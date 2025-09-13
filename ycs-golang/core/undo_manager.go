// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"container/list"
	"time"

	"github.com/chenrensong/ygo/contracts"
)

// StackItem represents an item in the undo/redo stack
type StackItem struct {
	BeforeState map[int64]int64
	AfterState  map[int64]int64
	Meta        map[string]interface{} // Use this to save and restore metadata like selection range
	DeleteSet   contracts.IDeleteSet
}

// NewStackItem creates a new StackItem
func NewStackItem(ds contracts.IDeleteSet, beforeState, afterState map[int64]int64) *StackItem {
	return &StackItem{
		DeleteSet:   ds,
		BeforeState: beforeState,
		AfterState:  afterState,
		Meta:        make(map[string]interface{}),
	}
}

// OperationType represents the type of operation (Undo or Redo)
type OperationType int

const (
	OperationUndo OperationType = iota
	OperationRedo
)

// StackEventArgs represents event arguments for stack operations
type StackEventArgs struct {
	StackItem          *StackItem
	Type               OperationType
	ChangedParentTypes map[contracts.IAbstractType][]contracts.IYEvent
	Origin             interface{}
}

// NewStackEventArgs creates new StackEventArgs
func NewStackEventArgs(item *StackItem, opType OperationType, changedParentTypes map[contracts.IAbstractType][]contracts.IYEvent, origin interface{}) *StackEventArgs {
	return &StackEventArgs{
		StackItem:          item,
		Type:               opType,
		ChangedParentTypes: changedParentTypes,
		Origin:             origin,
	}
}

// UndoManager fires 'stack-item-added' event when a stack item was added to either the undo- or
// the redo-stack. You may store additional stack information via the metadata property
// on 'event.stackItem.meta' (it is a collection of metadata properties).
// Fires 'stack-item-popped' event when a stack item was popped from either the undo- or
// the redo-stack. You may restore the saved stack information from 'event.stackItem.Meta'.
type UndoManager struct {
	scope          []contracts.IAbstractType
	deleteFilter   func(contracts.IStructItem) bool
	trackedOrigins map[interface{}]struct{}
	undoStack      *list.List
	redoStack      *list.List
	undoing        bool
	redoing        bool
	doc            contracts.IYDoc
	lastChange     time.Time
	captureTimeout time.Duration

	// Event handlers
	stackItemAddedHandlers  []func(*StackEventArgs)
	stackItemPoppedHandlers []func(*StackEventArgs)
}

// NewUndoManager creates a new UndoManager with a single type scope
func NewUndoManager(typeScope contracts.IAbstractType) *UndoManager {
	return NewUndoManagerWithScopes([]contracts.IAbstractType{typeScope}, 500*time.Millisecond, nil, map[interface{}]struct{}{nil: {}})
}

// NewUndoManagerWithScopes creates a new UndoManager with multiple type scopes
func NewUndoManagerWithScopes(typeScopes []contracts.IAbstractType, captureTimeout time.Duration, deleteFilter func(contracts.IStructItem) bool, trackedOrigins map[interface{}]struct{}) *UndoManager {
	if deleteFilter == nil {
		deleteFilter = func(contracts.IStructItem) bool { return true }
	}

	if trackedOrigins == nil {
		trackedOrigins = make(map[interface{}]struct{})
	}

	um := &UndoManager{
		scope:                   typeScopes,
		deleteFilter:            deleteFilter,
		trackedOrigins:          trackedOrigins,
		undoStack:               list.New(),
		redoStack:               list.New(),
		undoing:                 false,
		redoing:                 false,
		doc:                     typeScopes[0].GetDoc(),
		lastChange:              time.Time{},
		captureTimeout:          captureTimeout,
		stackItemAddedHandlers:  make([]func(*StackEventArgs), 0),
		stackItemPoppedHandlers: make([]func(*StackEventArgs), 0),
	}

	// Add this UndoManager to tracked origins
	um.trackedOrigins[um] = struct{}{}

	// Subscribe to document events
	um.doc.AddAfterTransactionHandler(um.onAfterTransaction)

	return um
}

// AddStackItemAddedHandler adds a handler for stack item added events
func (um *UndoManager) AddStackItemAddedHandler(handler func(*StackEventArgs)) {
	um.stackItemAddedHandlers = append(um.stackItemAddedHandlers, handler)
}

// AddStackItemPoppedHandler adds a handler for stack item popped events
func (um *UndoManager) AddStackItemPoppedHandler(handler func(*StackEventArgs)) {
	um.stackItemPoppedHandlers = append(um.stackItemPoppedHandlers, handler)
}

// Count returns the number of items in the undo stack
func (um *UndoManager) Count() int {
	return um.undoStack.Len()
}

// Clear clears both undo and redo stacks
func (um *UndoManager) Clear() {
	um.doc.Transact(func(tr contracts.ITransaction) {
		clearItem := func(stackItem *StackItem) {
			stackItem.DeleteSet.IterateDeletedStructs(tr, func(item contracts.IStructItem) bool {
				for _, abstractType := range um.scope {
					if um.isParentOf(abstractType, item) {
						item.KeepItemAndParents(false)
						break
					}
				}
				return true
			})
		}

		// Clear undo stack
		for e := um.undoStack.Front(); e != nil; e = e.Next() {
			clearItem(e.Value.(*StackItem))
		}

		// Clear redo stack
		for e := um.redoStack.Front(); e != nil; e = e.Next() {
			clearItem(e.Value.(*StackItem))
		}
	}, um)

	um.undoStack.Init()
	um.redoStack.Init()
}

// StopCapturing stops capturing changes into the current stack item
// UndoManager merges Undo-StackItem if they are created within time-gap
// smaller than 'captureTimeout'. Call this method so that the next StackItem
// won't be merged.
func (um *UndoManager) StopCapturing() {
	um.lastChange = time.Time{}
}

// Undo undoes the last changes on type
// Returns stack item if a change was applied
func (um *UndoManager) Undo() *StackItem {
	um.undoing = true
	defer func() {
		um.undoing = false
	}()

	return um.popStackItem(um.undoStack, OperationUndo)
}

// Redo redoes the last changes on type
// Returns stack item if a change was applied
func (um *UndoManager) Redo() *StackItem {
	um.redoing = true
	defer func() {
		um.redoing = false
	}()

	return um.popStackItem(um.redoStack, OperationRedo)
}

// onAfterTransaction handles after transaction events
func (um *UndoManager) onAfterTransaction(transaction contracts.ITransaction) {
	// Only track certain transactions
	hasChangedScope := false
	for _, abstractType := range um.scope {
		if _, exists := transaction.GetChangedParentTypes()[abstractType]; exists {
			hasChangedScope = true
			break
		}
	}

	if !hasChangedScope {
		return
	}

	if _, tracked := um.trackedOrigins[transaction.GetOrigin()]; !tracked && transaction.GetOrigin() != um {
		return
	}

	var stack *list.List
	var operationType OperationType

	if um.undoing {
		stack = um.redoStack
		operationType = OperationRedo
	} else if um.redoing {
		stack = um.undoStack
		operationType = OperationUndo
	} else {
		stack = um.undoStack
		operationType = OperationUndo
	}

	if transaction.GetOrigin() == um {
		// This transaction was created by this UndoManager
		return
	}

	now := time.Now()

	// Check if we should merge with the last stack item
	var lastStackItem *StackItem
	if stack.Len() > 0 && now.Sub(um.lastChange) < um.captureTimeout {
		lastStackItem = stack.Back().Value.(*StackItem)
	}

	// Create new delete set for changes
	deleteSet := NewDeleteSetFromStructStore(transaction.GetDoc().GetStore())

	if lastStackItem != nil && um.canMerge(lastStackItem, transaction) {
		// Merge with existing stack item
		lastStackItem.AfterState = transaction.GetAfterState()
		for k, v := range deleteSet.GetClients() {
			if existing, exists := lastStackItem.DeleteSet.GetClients()[k]; exists {
				lastStackItem.DeleteSet.GetClients()[k] = append(existing, v...)
			} else {
				lastStackItem.DeleteSet.GetClients()[k] = v
			}
		}
	} else {
		// Create new stack item
		stackItem := NewStackItem(deleteSet, transaction.GetBeforeState(), transaction.GetAfterState())
		stack.PushBack(stackItem)

		// Fire stack item added event
		args := NewStackEventArgs(stackItem, operationType, transaction.GetChangedParentTypes(), transaction.GetOrigin())
		for _, handler := range um.stackItemAddedHandlers {
			handler(args)
		}
	}

	um.lastChange = now

	// Clear the other stack if this is not an undo/redo operation
	if !um.undoing && !um.redoing {
		um.redoStack.Init()
	}
}

// popStackItem pops an item from the specified stack
func (um *UndoManager) popStackItem(stack *list.List, operationType OperationType) *StackItem {
	if stack.Len() == 0 {
		return nil
	}

	elem := stack.Back()
	stack.Remove(elem)
	stackItem := elem.Value.(*StackItem)

	um.doc.Transact(func(tr contracts.ITransaction) {
		// Apply the reverse changes
		um.restoreSnapshot(stackItem, tr)
	}, um)

	// Fire stack item popped event
	args := NewStackEventArgs(stackItem, operationType, nil, um)
	for _, handler := range um.stackItemPoppedHandlers {
		handler(args)
	}

	return stackItem
}

// canMerge checks if a stack item can be merged with a transaction
func (um *UndoManager) canMerge(stackItem *StackItem, transaction contracts.ITransaction) bool {
	// Simple merge logic - can be extended
	return transaction.GetOrigin() != um
}

// restoreSnapshot restores a snapshot from a stack item
func (um *UndoManager) restoreSnapshot(stackItem *StackItem, transaction contracts.ITransaction) {
	// This is a simplified implementation
	// In a full implementation, this would restore the document state
	// to what it was before the changes in the stack item
}

// isParentOf checks if a type is a parent of an item
func (um *UndoManager) isParentOf(abstractType contracts.IAbstractType, item contracts.IStructItem) bool {
	parent := item.GetParent()
	for parent != nil {
		if parent == abstractType {
			return true
		}
		if parentType, ok := parent.(contracts.IAbstractType); ok {
			if parentType.GetItem() != nil {
				parent = parentType.GetItem().GetParent()
			} else {
				break
			}
		} else {
			break
		}
	}
	return false
}
