// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"ycs-golang/structs"
	"ycs-golang/types"
)

// Transaction represents a transaction in the Yjs model
type Transaction struct {
	Doc                 *YDoc
	Origin              interface{}
	BeforeState         map[int64]int64
	AfterState          map[int64]int64
	Changed             map[*types.AbstractType]map[string]bool
	ChangedParentTypes  map[*types.AbstractType][]*YEvent
	Meta                map[string]interface{}
	Local               bool
	SubdocsAdded        map[*YDoc]bool
	SubdocsRemoved      map[*YDoc]bool
	SubdocsLoaded       map[*YDoc]bool
	DeleteSet           *DeleteSet
	mergeStructs        []*structs.AbstractStruct
}

// NewTransaction creates a new Transaction
func NewTransaction(doc *YDoc, origin interface{}, local bool) *Transaction {
	return &Transaction{
		Doc:                doc,
		Origin:             origin,
		BeforeState:        doc.Store.GetStateVector(),
		AfterState:         make(map[int64]int64),
		Changed:            make(map[*types.AbstractType]map[string]bool),
		ChangedParentTypes: make(map[*types.AbstractType][]*YEvent),
		Meta:               make(map[string]interface{}),
		Local:              local,
		SubdocsAdded:       make(map[*YDoc]bool),
		SubdocsRemoved:     make(map[*YDoc]bool),
		SubdocsLoaded:      make(map[*YDoc]bool),
		DeleteSet:          NewDeleteSet(),
		mergeStructs:       make([]*structs.AbstractStruct, 0),
	}
}

// GetNextId returns the next ID for this transaction
func (t *Transaction) GetNextId() ID {
	return ID{Client: t.Doc.ClientId, Clock: t.Doc.Store.GetState(t.Doc.ClientId)}
}

// AddChangedTypeToTransaction adds a changed type to the transaction
func (t *Transaction) AddChangedTypeToTransaction(typ *types.AbstractType, parentSub string) {
	item := typ.Item
	if item == nil || (func() bool {
		if clock, exists := t.BeforeState[item.Id.Client]; exists && item.Id.Clock < clock && !item.Deleted {
			return true
		}
		return false
	}()) {
		if _, exists := t.Changed[typ]; !exists {
			t.Changed[typ] = make(map[string]bool)
		}
		t.Changed[typ][parentSub] = true
	}
}

// CleanupTransactions cleans up transactions
func CleanupTransactions(transactionCleanups []*Transaction, i int) {
	if i < len(transactionCleanups) {
		transaction := transactionCleanups[i]
		doc := transaction.Doc
		store := doc.Store
		ds := transaction.DeleteSet
		mergeStructs := transaction.mergeStructs
		actions := make([]func(), 0)

		defer func() {
			// Replace deleted items with ItemDeleted / GC.
			// This is where content is actually removed from the Yjs Doc.
			if doc.Gc {
				ds.TryGcDeleteSet(store, func(item *structs.Item) bool {
					// This is a placeholder for the GC filter function
					return true
				})
			}

			ds.TryMergeDeleteSet(store)

			// On all affected store.clients props, try to merge.
			for client, clock := range transaction.AfterState {
				var beforeClock int64
				if bc, exists := transaction.BeforeState[client]; exists {
					beforeClock = bc
				}

				if beforeClock != clock {
					structs := store.Clients[client]
					firstChangePos := max(StructStoreFindIndexSS(structs, beforeClock), 1)
					for j := len(structs) - 1; j >= firstChangePos; j-- {
						TryToMergeWithLeft(structs, j)
					}
				}
			}

			// Try to merge mergeStructs.
			// TODO: It makes more sense to transform mergeStructs to a DS, sort it, and merge from right to left
			//       but at the moment DS does not handle duplicates.
			for j := 0; j < len(mergeStructs); j++ {
				client := mergeStructs[j].Id.Client
				clock := mergeStructs[j].Id.Clock
				structs := store.Clients[client]
				replacedStructPos := StructStoreFindIndexSS(structs, clock)

				if replacedStructPos+1 < len(structs) {
					TryToMergeWithLeft(structs, replacedStructPos+1)
				}

				if replacedStructPos > 0 {
					TryToMergeWithLeft(structs, replacedStructPos)
				}
			}

			if !transaction.Local {
				var afterClock, beforeClock int64
				if ac, exists := transaction.AfterState[doc.ClientId]; exists {
					afterClock = ac
				} else {
					afterClock = -1
				}

				if bc, exists := transaction.BeforeState[doc.ClientId]; exists {
					beforeClock = bc
				} else {
					beforeClock = -1
				}

				if afterClock != beforeClock {
					doc.ClientId = GenerateNewClientId()
					// Debug.WriteLine($"{nameof(Transaction)}: Changed the client-id because another client seems to be using it.");
				}
			}

			// @todo: Merge all the transactions into one and provide send the data as a single update message.
			doc.InvokeOnAfterTransactionCleanup(transaction)

			doc.InvokeUpdateV2(transaction)

			for subDoc := range transaction.SubdocsAdded {
				doc.Subdocs[subDoc] = true
			}

			for subDoc := range transaction.SubdocsRemoved {
				delete(doc.Subdocs, subDoc)
			}

			doc.InvokeSubdocsChanged(transaction.SubdocsLoaded, transaction.SubdocsAdded, transaction.SubdocsRemoved)

			for subDoc := range transaction.SubdocsRemoved {
				subDoc.Destroy()
			}

			if len(transactionCleanups) <= i+1 {
				// Clear transaction cleanups
				doc.transactionCleanups = make([]*Transaction, 0)
				doc.InvokeAfterAllTransactions(transactionCleanups)
			} else {
				CleanupTransactions(transactionCleanups, i+1)
			}
		}()

		ds.SortAndMergeDeleteSet()
		transaction.AfterState = store.GetStateVector()
		doc.transaction = nil

		actions = append(actions, func() {
			doc.InvokeOnBeforeObserverCalls(transaction)
		})

		actions = append(actions, func() {
			for itemType, subs := range transaction.Changed {
				if itemType.Item == nil || !itemType.Item.Deleted {
					itemType.CallObserver(transaction, subs)
				}
			}
		})

		actions = append(actions, func() {
			// Deep observe events.
			for typ, events := range transaction.ChangedParentTypes {
				// We need to think about the possibility that the user transforms the YDoc in the event.
				if typ.Item == nil || !typ.Item.Deleted {
					for _, evt := range events {
						if evt.Target.Item == nil || !evt.Target.Item.Deleted {
							evt.CurrentTarget = typ
						}
					}

					// Sort events by path length so that top-level events are fired first.
					// In Go, we need to implement our own sort function
					// This is a simplified version - you may need to implement a proper sort
					sortedEvents := make([]*YEvent, len(events))
					copy(sortedEvents, events)

					actions = append(actions, func() {
						typ.CallDeepEventHandlerListeners(sortedEvents, transaction)
					})
				}
			}
		})

		actions = append(actions, func() {
			doc.InvokeOnAfterTransaction(transaction)
		})

		CallAll(actions)
	}
}

// RedoItem redoes the effect of an operation
func (t *Transaction) RedoItem(item *structs.Item, redoItems map[*structs.Item]bool) *structs.AbstractStruct {
	doc := t.Doc
	store := doc.Store
	ownClientId := doc.ClientId
	redone := item.Redone

	if redone != nil {
		return store.GetItemCleanStart(t, *redone)
	}

	var parentItem *structs.Item
	if parentType, ok := item.Parent.(*types.AbstractType); ok && parentType.Item != nil {
		parentItem = parentType.Item
	}

	var left, right *structs.AbstractStruct

	if item.ParentSub == "" {
		// Is an array item. Insert at the old position.
		left = item.Left
		right = item
	} else {
		// Is a map item. Insert at current value.
		left = item
		for {
			if leftItem, ok := left.(*structs.Item); ok && leftItem.Right != nil {
				left = leftItem.Right
				if leftItem.Id.Client != ownClientId {
					// It is not possible to redo this item because it conflicts with a change from another client.
					return nil
				}
			} else {
				break
			}
		}

		if leftItem, ok := left.(*structs.Item); ok && leftItem.Right != nil {
			if parentType, ok := item.Parent.(*types.AbstractType); ok {
				left = parentType.Map[item.ParentSub]
			}
		}

		right = nil
	}

	// Make sure that parent is redone.
	if parentItem != nil && parentItem.Deleted && parentItem.Redone == nil {
		// Try to undo parent if it will be undone anyway.
		if !redoItems[parentItem] || t.RedoItem(parentItem, redoItems) == nil {
			return nil
		}
	}

	if parentItem != nil && parentItem.Redone != nil {
		for parentItem.Redone != nil {
			parentItem = store.GetItemCleanStart(t, *parentItem.Redone).(*structs.Item)
		}

		// Find next cloned_redo items.
		for left != nil {
			leftTrace := left
			for leftTrace != nil {
				var leftParentItem *structs.Item
				if leftItem, ok := leftTrace.(*structs.Item); ok && leftItem.Parent != nil {
					if leftParentType, ok := leftItem.Parent.(*types.AbstractType); ok && leftParentType.Item != nil {
						leftParentItem = leftParentType.Item
					}
				}

				if leftParentItem != parentItem {
					if leftItem, ok := leftTrace.(*structs.Item); ok && leftItem.Redone != nil {
						leftTrace = store.GetItemCleanStart(t, *leftItem.Redone)
					} else {
						leftTrace = nil
					}
				} else {
					break
				}
			}

			if leftTrace != nil {
				var leftParentItem *structs.Item
				if leftItem, ok := leftTrace.(*structs.Item); ok && leftItem.Parent != nil {
					if leftParentType, ok := leftItem.Parent.(*types.AbstractType); ok && leftParentType.Item != nil {
						leftParentItem = leftParentType.Item
					}
				}

				if leftParentItem == parentItem {
					left = leftTrace
					break
				}
			}

			if leftItem, ok := left.(*structs.Item); ok {
				left = leftItem.Left
			} else {
				left = nil
			}
		}

		for right != nil {
			rightTrace := right
			for rightTrace != nil {
				var rightParentItem *structs.Item
				if rightItem, ok := rightTrace.(*structs.Item); ok && rightItem.Parent != nil {
					if rightParentType, ok := rightItem.Parent.(*types.AbstractType); ok && rightParentType.Item != nil {
						rightParentItem = rightParentType.Item
					}
				}

				if rightParentItem != parentItem {
					if rightItem, ok := rightTrace.(*structs.Item); ok && rightItem.Redone != nil {
						rightTrace = store.GetItemCleanStart(t, *rightItem.Redone)
					} else {
						rightTrace = nil
					}
				} else {
					break
				}
			}

			if rightTrace != nil {
				var rightParentItem *structs.Item
				if rightItem, ok := rightTrace.(*structs.Item); ok && rightItem.Parent != nil {
					if rightParentType, ok := rightItem.Parent.(*types.AbstractType); ok && rightParentType.Item != nil {
						rightParentItem = rightParentType.Item
					}
				}

				if rightParentItem == parentItem {
					right = rightTrace
					break
				}
			}

			if rightItem, ok := right.(*structs.Item); ok {
				right = rightItem.Right
			} else {
				right = nil
			}
		}
	}

	nextClock := store.GetState(ownClientId)
	nextId := ID{Client: ownClientId, Clock: nextClock}

	redoneItem := structs.NewItem(
		nextId,
		left,
		func() *ID {
			if leftItem, ok := left.(*structs.Item); ok {
				return leftItem.LastId()
			}
			return nil
		}(),
		right,
		func() *ID {
			if right != nil {
				return right.Id
			}
			return nil
		}(),
		func() interface{} {
			if parentItem == nil {
				return item.Parent
			}
			if contentType, ok := parentItem.Content.(*structs.ContentType); ok {
				return contentType.Type
			}
			return nil
		}(),
		item.ParentSub,
		item.Content.Copy())

	item.Redone = &nextId

	redoneItem.KeepItemAndParents(true)
	redoneItem.Integrate(t, 0)

	return redoneItem
}

// SplitSnapshotAffectedStructs splits snapshot affected structs
func SplitSnapshotAffectedStructs(transaction *Transaction, snapshot *Snapshot) {
	var metaObj interface{}
	var exists bool
	if metaObj, exists = transaction.Meta["splitSnapshotAffectedStructs"]; !exists {
		metaObj = make(map[*Snapshot]bool)
		transaction.Meta["splitSnapshotAffectedStructs"] = metaObj
	}

	meta := metaObj.(map[*Snapshot]bool)
	store := transaction.Doc.Store

	// Check if we already split for this snapshot.
	if !meta[snapshot] {
		for client, clock := range snapshot.StateVector {
			if clock < store.GetState(client) {
				store.GetItemCleanStart(transaction, ID{Client: client, Clock: clock})
			}
		}

		snapshot.DeleteSet.IterateDeletedStructs(transaction, func(item *structs.AbstractStruct) bool {
			return true
		})
		meta[snapshot] = true
	}
}

// WriteUpdateMessageFromTransaction writes an update message from a transaction
func (t *Transaction) WriteUpdateMessageFromTransaction(encoder IUpdateEncoder) bool {
	// Check if there are any changes
	hasChanges := len(t.DeleteSet.Clients) > 0
	if !hasChanges {
		for key, value := range t.AfterState {
			if clockB, exists := t.BeforeState[key]; !exists || value != clockB {
				hasChanges = true
				break
			}
		}
	}

	if !hasChanges {
		return false
	}

	t.DeleteSet.SortAndMergeDeleteSet()
	// Note: EncodingUtils.WriteClientsStructs needs to be implemented
	// EncodingUtils.WriteClientsStructs(encoder, t.Doc.Store, t.BeforeState)
	t.DeleteSet.Write(encoder)

	return true
}

// CallAll calls all functions in the list
func CallAll(funcs []func(), index int) {
	defer func() {
		if r := recover(); r != nil && index < len(funcs)-1 {
			CallAll(funcs, index+1)
		}
	}()

	for i := index; i < len(funcs); i++ {
		funcs[i]()
	}
}

// Helper function to find maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}