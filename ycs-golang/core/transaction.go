package core

import (
	"sort"
	"ycs/contracts"
)

// Transaction represents a transaction that bundles changes on the Yjs model
// to minimize the number of messages sent and observer calls
type Transaction struct {
	doc                contracts.IYDoc
	deleteSet          contracts.IDeleteSet
	beforeState        map[int64]int64
	afterState         map[int64]int64
	changed            map[contracts.IAbstractType]map[string]struct{}
	changedParentTypes map[contracts.IAbstractType][]contracts.IYEvent
	mergeStructs       []contracts.IStructItem
	origin             interface{}
	meta               map[string]interface{}
	local              bool
	subdocsAdded       map[contracts.IYDoc]struct{}
	subdocsRemoved     map[contracts.IYDoc]struct{}
	subdocsLoaded      map[contracts.IYDoc]struct{}
}

// NewTransaction creates a new transaction
func NewTransaction(doc contracts.IYDoc, origin interface{}, local bool) *Transaction {
	return &Transaction{
		doc:                doc,
		deleteSet:          NewDeleteSet(),
		beforeState:        doc.GetStore().GetStateVector(),
		afterState:         make(map[int64]int64),
		changed:            make(map[contracts.IAbstractType]map[string]struct{}),
		changedParentTypes: make(map[contracts.IAbstractType][]contracts.IYEvent),
		mergeStructs:       make([]contracts.IStructItem, 0),
		origin:             origin,
		meta:               make(map[string]interface{}),
		local:              local,
		subdocsAdded:       make(map[contracts.IYDoc]struct{}),
		subdocsRemoved:     make(map[contracts.IYDoc]struct{}),
		subdocsLoaded:      make(map[contracts.IYDoc]struct{}),
	}
}

// GetDoc returns the Yjs document instance
func (tr *Transaction) GetDoc() contracts.IYDoc {
	return tr.doc
}

// GetOrigin returns the transaction origin
func (tr *Transaction) GetOrigin() interface{} {
	return tr.origin
}

// GetBeforeState returns the state before the transaction started
func (tr *Transaction) GetBeforeState() map[int64]int64 {
	return tr.beforeState
}

// SetBeforeState sets the before state
func (tr *Transaction) SetBeforeState(beforeState map[int64]int64) {
	tr.beforeState = beforeState
}

// GetAfterState returns the state after the transaction
func (tr *Transaction) GetAfterState() map[int64]int64 {
	return tr.afterState
}

// SetAfterState sets the after state
func (tr *Transaction) SetAfterState(afterState map[int64]int64) {
	tr.afterState = afterState
}

// GetChanged returns all types that were directly modified
func (tr *Transaction) GetChanged() map[contracts.IAbstractType]map[string]struct{} {
	return tr.changed
}

// GetChangedParentTypes returns the events for types that observe child elements
func (tr *Transaction) GetChangedParentTypes() map[contracts.IAbstractType][]contracts.IYEvent {
	return tr.changedParentTypes
}

// GetMeta returns the meta information on the transaction
func (tr *Transaction) GetMeta() map[string]interface{} {
	return tr.meta
}

// GetLocal returns whether this change originates from this doc
func (tr *Transaction) GetLocal() bool {
	return tr.local
}

// GetSubdocsAdded returns the subdocuments added in this transaction
func (tr *Transaction) GetSubdocsAdded() map[contracts.IYDoc]struct{} {
	return tr.subdocsAdded
}

// GetSubdocsRemoved returns the subdocuments removed in this transaction
func (tr *Transaction) GetSubdocsRemoved() map[contracts.IYDoc]struct{} {
	return tr.subdocsRemoved
}

// GetSubdocsLoaded returns the subdocuments loaded in this transaction
func (tr *Transaction) GetSubdocsLoaded() map[contracts.IYDoc]struct{} {
	return tr.subdocsLoaded
}

// GetDeleteSet returns the set of deleted items by IDs
func (tr *Transaction) GetDeleteSet() contracts.IDeleteSet {
	return tr.deleteSet
}

// GetMergeStructs returns the merge structs
func (tr *Transaction) GetMergeStructs() []contracts.IStructItem {
	return tr.mergeStructs
}

// GetNextID returns the next ID for this transaction
func (tr *Transaction) GetNextID() StructID {
	return StructID{
		Client: tr.doc.GetClientID(),
		Clock:  tr.doc.GetStore().GetState(tr.doc.GetClientID()),
	}
}

// AddChangedTypeToTransaction adds a changed type to the transaction
// If 'type.parent' was added in current transaction, 'type' technically did not change,
// it was just added and we should not fire events for 'type'.
func (tr *Transaction) AddChangedTypeToTransaction(yType contracts.IAbstractType, parentSub string) {
	item := yType.GetItem()
	if item == nil {
		// Add to changed types
		if tr.changed[yType] == nil {
			tr.changed[yType] = make(map[string]struct{})
		}
		tr.changed[yType][parentSub] = struct{}{}
		return
	}

	clock, exists := tr.beforeState[item.GetID().Client]
	if exists && item.GetID().Clock < clock && !item.IsDeleted() {
		if tr.changed[yType] == nil {
			tr.changed[yType] = make(map[string]struct{})
		}
		tr.changed[yType][parentSub] = struct{}{}
	}
}

// CleanupTransactions cleans up transactions and calls observers
func CleanupTransactions(transactionCleanups []contracts.ITransaction, i int) {
	if i >= len(transactionCleanups) {
		return
	}

	transaction := transactionCleanups[i]
	doc := transaction.GetDoc()
	store := doc.GetStore()
	deleteSet := transaction.GetDeleteSet()
	mergeStructs := transaction.GetMergeStructs()

	// Actions to be executed
	var actions []func()

	defer func() {
		// Replace deleted items with ItemDeleted / GC
		// This is where content is actually removed from the Yjs Doc
		if doc.GetGc() {
			deleteSet.TryGcDeleteSet(store, doc.GetGcFilter())
		}

		deleteSet.TryMergeDeleteSet(store)

		// On all affected store.clients props, try to merge
		for client, clock := range transaction.GetAfterState() {
			beforeClock := int64(0)
			if bc, exists := transaction.GetBeforeState()[client]; exists {
				beforeClock = bc
			}

			if beforeClock != clock {
				structs := store.GetClients()[client]
				firstChangePos := max(FindIndexSS(structs, beforeClock), 1)
				for j := len(structs) - 1; j >= firstChangePos; j-- {
					TryToMergeWithLeft(structs, j)
				}
			}
		}

		// Try to merge mergeStructs
		for _, item := range mergeStructs {
			client := item.GetID().Client
			clock := item.GetID().Clock
			structs := store.GetClients()[client]
			replacedStructPos := FindIndexSS(structs, clock)

			if replacedStructPos+1 < len(structs) {
				TryToMergeWithLeft(structs, replacedStructPos+1)
			}

			if replacedStructPos > 0 {
				TryToMergeWithLeft(structs, replacedStructPos)
			}
		}

		if !transaction.GetLocal() {
			afterClock := int64(-1)
			if ac, exists := transaction.GetAfterState()[doc.GetClientID()]; exists {
				afterClock = ac
			}

			beforeClock := int64(-1)
			if bc, exists := transaction.GetBeforeState()[doc.GetClientID()]; exists {
				beforeClock = bc
			}

			if afterClock != beforeClock {
				doc.SetClientID(generateNewClientID())
			}
		}

		doc.InvokeOnAfterTransactionCleanup(transaction)
		doc.InvokeUpdateV2(transaction)

		for subDoc := range transaction.GetSubdocsAdded() {
			doc.GetSubdocs()[subDoc] = struct{}{}
		}

		for subDoc := range transaction.GetSubdocsRemoved() {
			delete(doc.GetSubdocs(), subDoc)
		}

		doc.InvokeSubdocsChanged(
			transaction.GetSubdocsLoaded(),
			transaction.GetSubdocsAdded(),
			transaction.GetSubdocsRemoved(),
		)

		for subDoc := range transaction.GetSubdocsRemoved() {
			subDoc.Destroy()
		}

		if len(transactionCleanups) <= i+1 {
			// Clear transaction cleanups and invoke after all transactions
			cleanups := doc.GetTransactionCleanups()
			for len(cleanups) > 0 {
				cleanups = cleanups[:len(cleanups)-1]
			}
			doc.InvokeAfterAllTransactions(transactionCleanups)
		} else {
			CleanupTransactions(transactionCleanups, i+1)
		}
	}()

	// Sort and merge delete set
	deleteSet.SortAndMergeDeleteSet()
	transaction.SetAfterState(store.GetStateVector())
	doc.SetTransaction(nil)

	// Add observer call actions
	actions = append(actions, func() {
		doc.InvokeOnBeforeObserverCalls(transaction)
	})

	actions = append(actions, func() {
		for itemType, subs := range transaction.GetChanged() {
			if itemType.GetItem() == nil || !itemType.GetItem().IsDeleted() {
				itemType.CallObserver(transaction, subs)
			}
		}
	})

	actions = append(actions, func() {
		// Deep observe events
		for yType, events := range transaction.GetChangedParentTypes() {
			// We need to think about the possibility that the user transforms the YDoc in the event
			if yType.GetItem() == nil || !yType.GetItem().IsDeleted() {
				for _, evt := range events {
					if evt.GetTarget().GetItem() == nil || !evt.GetTarget().GetItem().IsDeleted() {
						evt.SetCurrentTarget(yType)
					}
				}

				// Sort events by path length so that top-level events are fired first
				sortedEvents := make([]contracts.IYEvent, len(events))
				copy(sortedEvents, events)
				sort.Slice(sortedEvents, func(i, j int) bool {
					return len(sortedEvents[i].GetPath()) < len(sortedEvents[j].GetPath())
				})

				if len(sortedEvents) > 0 {
					actions = append(actions, func() {
						yType.CallDeepEventHandlerListeners(sortedEvents, transaction)
					})
				}
			}
		}
	})

	actions = append(actions, func() {
		doc.InvokeOnAfterTransaction(transaction)
	})

	// Execute all actions
	callAll(actions, 0)
}

// RedoItem redoes the effect of an operation
func (tr *Transaction) RedoItem(item contracts.IStructItem, redoItems map[contracts.IStructItem]struct{}) contracts.IStructItem {
	doc := tr.doc
	store := doc.GetStore()
	ownClientID := doc.GetClientID()
	redone := item.GetRedone()

	if redone != nil {
		return store.GetItemCleanStart(tr, *redone)
	}

	parentItem := item.GetParent().(contracts.IAbstractType).GetItem()
	var left contracts.IStructItem
	var right contracts.IStructItem

	if item.GetParentSub() == nil {
		// Is an array item. Insert at the old position.
		left = item.GetLeft()
		right = item
	} else {
		// Is a map item. Insert at current value.
		left = item
		for left.GetRight() != nil {
			left = left.GetRight()
			if left.GetID().Client != ownClientID {
				// It is not possible to redo this item because it conflicts with a change from another client.
				return nil
			}
		}

		if left.GetRight() != nil {
			parent := item.GetParent().(contracts.IAbstractType)
			left = parent.GetMap()[*item.GetParentSub()]
		}

		right = nil
	}

	// Make sure that parent is redone
	if parentItem != nil && parentItem.IsDeleted() && parentItem.GetRedone() == nil {
		// Try to undo parent if it will be undone anyway
		if _, exists := redoItems[parentItem]; !exists || tr.RedoItem(parentItem, redoItems) == nil {
			return nil
		}
	}

	if parentItem != nil && parentItem.GetRedone() != nil {
		for parentItem.GetRedone() != nil {
			parentItem = store.GetItemCleanStart(tr, *parentItem.GetRedone())
		}

		// Find next cloned_redo items
		for left != nil {
			leftTrace := left
			for leftTrace != nil && leftTrace.GetParent().(contracts.IAbstractType).GetItem() != parentItem {
				if leftTrace.GetRedone() == nil {
					leftTrace = nil
				} else {
					leftTrace = store.GetItemCleanStart(tr, *leftTrace.GetRedone())
				}
			}

			if leftTrace != nil && leftTrace.GetParent().(contracts.IAbstractType).GetItem() == parentItem {
				left = leftTrace
				break
			}

			left = left.GetLeft()
		}

		for right != nil {
			rightTrace := right
			for rightTrace != nil && rightTrace.GetParent().(contracts.IAbstractType).GetItem() != parentItem {
				if rightTrace.GetRedone() == nil {
					rightTrace = nil
				} else {
					rightTrace = store.GetItemCleanStart(tr, *rightTrace.GetRedone())
				}
			}

			if rightTrace != nil && rightTrace.GetParent().(contracts.IAbstractType).GetItem() == parentItem {
				right = rightTrace
				break
			}

			right = right.GetRight()
		}
	}

	nextClock := store.GetState(ownClientID)
	nextID := StructID{Client: ownClientID, Clock: nextClock}

	var lastID *StructID
	if left != nil {
		lastID = left.GetLastID()
	}

	var rightID *StructID
	if right != nil {
		rightID = &right.GetID()
	}

	var parent interface{}
	if parentItem == nil {
		parent = item.GetParent()
	} else {
		// This would need proper content type handling
		parent = parentItem // Simplified
	}

	redoneItem := NewStructItem(
		nextID,
		left,
		lastID,
		right,
		rightID,
		parent,
		item.GetParentSub(),
		item.GetContent().Copy(),
	)

	item.SetRedone(&nextID)
	redoneItem.KeepItemAndParents(true)
	redoneItem.Integrate(tr, 0)

	return redoneItem
}

// WriteUpdateMessageFromTransaction writes update message from transaction
func (tr *Transaction) WriteUpdateMessageFromTransaction(encoder contracts.IUpdateEncoder) bool {
	if len(tr.deleteSet.GetClients()) == 0 {
		// Check if any state changed
		hasChanges := false
		for client, clockA := range tr.afterState {
			if clockB, exists := tr.beforeState[client]; !exists || clockA != clockB {
				hasChanges = true
				break
			}
		}
		if !hasChanges {
			return false
		}
	}

	tr.deleteSet.SortAndMergeDeleteSet()
	WriteClientsStructs(encoder, tr.doc.GetStore(), tr.beforeState)
	tr.deleteSet.Write(encoder)

	return true
}

// Helper functions
func callAll(funcs []func(), index int) {
	defer func() {
		if r := recover(); r != nil {
			if index < len(funcs) {
				callAll(funcs, index+1)
			}
		}
	}()

	for i := index; i < len(funcs); i++ {
		funcs[i]()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
