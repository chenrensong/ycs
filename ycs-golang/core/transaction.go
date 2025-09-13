// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"math"
	"sort"

	"github.com/chenrensong/ygo/contracts"
)

// Transaction is created for every change on the Yjs model. It is possible
// to bundle changes on the Yjs model in a single transaction to minimize
// the number of messages sent and the number of observer calls.
// If possible the user of this library should bundle as many changes as possible.
type Transaction struct {
	doc                contracts.IYDoc
	origin             interface{}
	beforeState        map[int64]int64
	afterState         map[int64]int64
	changed            map[contracts.IAbstractType]map[string]struct{}
	changedParentTypes map[contracts.IAbstractType][]contracts.IYEvent
	mergeStructs       []contracts.IStructItem
	meta               map[string]interface{}
	local              bool
	subdocsAdded       map[contracts.IYDoc]struct{}
	subdocsRemoved     map[contracts.IYDoc]struct{}
	subdocsLoaded      map[contracts.IYDoc]struct{}
	deleteSet          contracts.IDeleteSet
}

// NewTransaction creates a new transaction
func NewTransaction(doc contracts.IYDoc, origin interface{}, local bool) *Transaction {
	return &Transaction{
		doc:                doc,
		origin:             origin,
		beforeState:        doc.GetStore().GetStateVector(),
		afterState:         make(map[int64]int64),
		changed:            make(map[contracts.IAbstractType]map[string]struct{}),
		changedParentTypes: make(map[contracts.IAbstractType][]contracts.IYEvent),
		mergeStructs:       make([]contracts.IStructItem, 0),
		meta:               make(map[string]interface{}),
		local:              local,
		subdocsAdded:       make(map[contracts.IYDoc]struct{}),
		subdocsRemoved:     make(map[contracts.IYDoc]struct{}),
		subdocsLoaded:      make(map[contracts.IYDoc]struct{}),
		deleteSet:          NewDeleteSet(),
	}
}

// GetDoc returns the document
func (t *Transaction) GetDoc() contracts.IYDoc {
	return t.doc
}

// GetOrigin returns the origin
func (t *Transaction) GetOrigin() interface{} {
	return t.origin
}

// GetBeforeState returns the before state
func (t *Transaction) GetBeforeState() map[int64]int64 {
	return t.beforeState
}

// SetBeforeState sets the before state
func (t *Transaction) SetBeforeState(beforeState map[int64]int64) {
	t.beforeState = beforeState
}

// GetAfterState returns the after state
func (t *Transaction) GetAfterState() map[int64]int64 {
	return t.afterState
}

// SetAfterState sets the after state
func (t *Transaction) SetAfterState(afterState map[int64]int64) {
	t.afterState = afterState
}

// GetChanged returns the changed types
func (t *Transaction) GetChanged() map[contracts.IAbstractType]map[string]struct{} {
	return t.changed
}

// GetChangedParentTypes returns the changed parent types
func (t *Transaction) GetChangedParentTypes() map[contracts.IAbstractType][]contracts.IYEvent {
	return t.changedParentTypes
}

// GetMergeStructs returns the merge structs
func (t *Transaction) GetMergeStructs() []contracts.IStructItem {
	return t.mergeStructs
}

// GetMeta returns the meta information
func (t *Transaction) GetMeta() map[string]interface{} {
	return t.meta
}

// GetLocal returns whether this is a local transaction
func (t *Transaction) GetLocal() bool {
	return t.local
}

// GetSubdocsAdded returns the added subdocs
func (t *Transaction) GetSubdocsAdded() map[contracts.IYDoc]struct{} {
	return t.subdocsAdded
}

// GetSubdocsRemoved returns the removed subdocs
func (t *Transaction) GetSubdocsRemoved() map[contracts.IYDoc]struct{} {
	return t.subdocsRemoved
}

// GetSubdocsLoaded returns the loaded subdocs
func (t *Transaction) GetSubdocsLoaded() map[contracts.IYDoc]struct{} {
	return t.subdocsLoaded
}

// GetDeleteSet returns the delete set
func (t *Transaction) GetDeleteSet() contracts.IDeleteSet {
	return t.deleteSet
}

// GetNextID returns the next ID for this transaction
func (t *Transaction) GetNextID() contracts.StructID {
	return contracts.StructID{
		Client: t.doc.GetClientID(),
		Clock:  t.doc.GetStore().GetState(t.doc.GetClientID()),
	}
}

// AddChangedTypeToTransaction adds a changed type to the transaction
// If 'type.parent' was added in current transaction, 'type' technically did not change,
// it was just added and we should not fire events for 'type'.
func (t *Transaction) AddChangedTypeToTransaction(abstractType contracts.IAbstractType, parentSub string) {
	item := abstractType.GetItem()

	var shouldAdd bool
	if item == nil {
		shouldAdd = true
	} else {
		if clock, exists := t.beforeState[item.GetID().Client]; exists {
			shouldAdd = item.GetID().Clock < clock && !item.GetDeleted()
		} else {
			shouldAdd = !item.GetDeleted()
		}
	}

	if shouldAdd {
		set, exists := t.changed[abstractType]
		if !exists {
			set = make(map[string]struct{})
			t.changed[abstractType] = set
		}
		set[parentSub] = struct{}{}
	}
}

// CleanupTransactions cleans up transactions
func CleanupTransactions(transactionCleanups []contracts.ITransaction, i int) {
	if i < len(transactionCleanups) {
		transaction := transactionCleanups[i]
		doc := transaction.GetDoc()
		store := doc.GetStore()
		ds := transaction.GetDeleteSet()
		mergeStructs := transaction.GetMergeStructs()
		var actions []func()

		// Sort and merge delete set
		ds.SortAndMergeDeleteSet()
		transaction.SetAfterState(store.GetStateVector())
		doc.SetTransaction(nil)

		// Add observer actions
		actions = append(actions, func() {
			doc.InvokeOnBeforeObserverCalls(transaction)
		})

		actions = append(actions, func() {
			// Call observers for changed types
			for itemType, subs := range transaction.GetChanged() {
				if itemType.GetItem() == nil || !itemType.GetItem().GetDeleted() {
					itemType.CallObserver(transaction, subs)
				}
			}
		})

		actions = append(actions, func() {
			// Deep observe events
			for abstractType, events := range transaction.GetChangedParentTypes() {
				// We need to think about the possibility that the user transforms the YDoc in the event
				if abstractType.GetItem() == nil || !abstractType.GetItem().GetDeleted() {
					for _, evt := range events {
						if evt.GetTarget().GetItem() == nil || !evt.GetTarget().GetItem().GetDeleted() {
							evt.SetCurrentTarget(abstractType)
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
							abstractType.CallDeepEventHandlerListeners(sortedEvents, transaction)
						})
					}
				}
			}
		})

		actions = append(actions, func() {
			doc.InvokeOnAfterTransaction(transaction)
		})

		// Execute all actions
		for _, action := range actions {
			action()
		}

		// Cleanup phase
		// Replace deleted items with ItemDeleted / GC
		// This is where content is actually removed from the Yjs Doc
		if doc.GetGc() {
			ds.TryGcDeleteSet(store, doc.GetGcFilter())
		}

		ds.TryMergeDeleteSet(store)

		// On all affected store.clients props, try to merge
		for client, clock := range transaction.GetAfterState() {
			beforeClock, exists := transaction.GetBeforeState()[client]
			if !exists {
				beforeClock = 0
			}

			if beforeClock != clock {
				structs := store.GetClients()[client]
				firstChangePos := int(math.Max(float64(FindIndexSS(structs, beforeClock)), 1))
				for j := len(structs) - 1; j >= firstChangePos; j-- {
					TryToMergeWithLeft(structs, j)
				}
			}
		}

		// Try to merge mergeStructs
		for _, mergeStruct := range mergeStructs {
			client := mergeStruct.GetID().Client
			clock := mergeStruct.GetID().Clock
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
			afterClock, afterExists := transaction.GetAfterState()[doc.GetClientID()]
			if !afterExists {
				afterClock = -1
			}

			beforeClock, beforeExists := transaction.GetBeforeState()[doc.GetClientID()]
			if !beforeExists {
				beforeClock = -1
			}

			if afterClock != beforeClock {
				doc.SetClientID(doc.GenerateNewClientID())
				// Debug: Changed the client-id because another client seems to be using it
			}
		}

		// Invoke cleanup and update handlers
		doc.InvokeOnAfterTransactionCleanup(transaction)
		doc.InvokeUpdateV2(transaction)

		// Handle subdocs
		for subDoc := range transaction.GetSubdocsAdded() {
			doc.GetSubdocs()[subDoc] = struct{}{}
		}

		for subDoc := range transaction.GetSubdocsRemoved() {
			delete(doc.GetSubdocs(), subDoc)
		}

		doc.InvokeSubdocsChanged(transaction.GetSubdocsLoaded(), transaction.GetSubdocsAdded(), transaction.GetSubdocsRemoved())

		for subDoc := range transaction.GetSubdocsRemoved() {
			subDoc.Destroy()
		}

		if len(transactionCleanups) <= i+1 {
			doc.GetTransactionCleanups().Clear()
			doc.InvokeAfterAllTransactions(transactionCleanups)
		} else {
			CleanupTransactions(transactionCleanups, i+1)
		}
	}
}

// CallAll executes all actions in sequence
func CallAll(actions []func()) {
	for _, action := range actions {
		action()
	}
}
