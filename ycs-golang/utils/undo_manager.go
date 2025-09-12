// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"container/list"
	"time"
	"ycs-golang/structs"
	"ycs-golang/types"
)

// OperationType represents the type of operation
type OperationType int

const (
	Undo OperationType = iota
	Redo
)

// StackItem represents an item in the undo/redo stack
type StackItem struct {
	BeforeState        map[int64]int64
	AfterState         map[int64]int64
	Meta               map[string]interface{}
	DeleteSet          *DeleteSet
}

// NewStackItem creates a new StackItem
func NewStackItem(ds *DeleteSet, beforeState, afterState map[int64]int64) *StackItem {
	return &StackItem{
		DeleteSet:          ds,
		BeforeState:        beforeState,
		AfterState:         afterState,
		Meta:               make(map[string]interface{}),
	}
}

// StackEventArgs represents arguments for stack events
type StackEventArgs struct {
	StackItem         *StackItem
	Type              OperationType
	ChangedParentTypes map[*types.AbstractType][]*YEvent
	Origin            interface{}
}

// UndoManager manages undo/redo operations
type UndoManager struct {
	scope              []*types.AbstractType
	deleteFilter       func(*structs.Item) bool
	trackedOrigins     map[interface{}]bool
	undoStack          *list.List
	redoStack          *list.List
	undoing            bool
	redoing            bool
	doc                *YDoc
	lastChange         time.Time
	captureTimeout     int
	stackItemAddedHandlers []func(*StackEventArgs)
	stackItemPoppedHandlers []func(*StackEventArgs)
}

// NewUndoManager creates a new UndoManager
func NewUndoManager(typeScope *types.AbstractType) *UndoManager {
	return NewUndoManagerWithOptions([]*types.AbstractType{typeScope}, 500, func(it *structs.Item) bool { return true }, make(map[interface{}]bool))
}

// NewUndoManagerWithOptions creates a new UndoManager with options
func NewUndoManagerWithOptions(typeScopes []*types.AbstractType, captureTimeout int, deleteFilter func(*structs.Item) bool, trackedOrigins map[interface{}]bool) *UndoManager {
	um := &UndoManager{
		scope:              typeScopes,
		deleteFilter:       deleteFilter,
		trackedOrigins:     trackedOrigins,
		undoStack:          list.New(),
		redoStack:          list.New(),
		undoing:            false,
		redoing:            false,
		doc:                typeScopes[0].Doc,
		lastChange:         time.Time{},
		captureTimeout:     captureTimeout,
		stackItemAddedHandlers: []func(*StackEventArgs){},
		stackItemPoppedHandlers: []func(*StackEventArgs){},
	}
	
	// Add this undo manager to tracked origins
	um.trackedOrigins[um] = true
	
	// Register after transaction handler
	um.doc.AfterTransaction = append(um.doc.AfterTransaction, um.onAfterTransaction)
	
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

// Clear clears the undo/redo stacks
func (um *UndoManager) Clear() {
	um.doc.Transact(func(tr *Transaction) {
		clearItem := func(stackItem *StackItem) {
			stackItem.DeleteSet.IterateDeletedStructs(tr, func(i *structs.AbstractStruct) bool {
				if item, ok := i.(*structs.Item); ok {
					for _, typeItem := range um.scope {
						if um.isParentOf(typeItem, item) {
							item.KeepItemAndParents(false)
							break
						}
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
	})

	um.undoStack.Init()
	um.redoStack.Init()
}

// StopCapturing stops capturing changes for the current operation
func (um *UndoManager) StopCapturing() {
	um.lastChange = time.Time{}
}

// Undo undoes the last change
func (um *UndoManager) Undo() *StackItem {
	um.undoing = true
	defer func() {
		um.undoing = false
	}()
	
	return um.popStackItem(um.undoStack, Undo)
}

// Redo redoes the last change
func (um *UndoManager) Redo() *StackItem {
	um.redoing = true
	defer func() {
		um.redoing = false
	}()
	
	return um.popStackItem(um.redoStack, Redo)
}

// onAfterTransaction handles after transaction events
func (um *UndoManager) onAfterTransaction(transaction *Transaction) {
	// Only track certain transactions.
	shouldTrack := false
	for _, typeItem := range um.scope {
		if _, exists := transaction.ChangedParentTypes[typeItem]; exists {
			shouldTrack = true
			break
		}
	}
	
	if !shouldTrack {
		// Check if origin is tracked
		shouldTrack = um.trackedOrigins[transaction.Origin]
		// TODO: Implement type checking for origins
	}
	
	if !shouldTrack {
		return
	}

	undoing := um.undoing
	redoing := um.redoing
	var stack *list.List
	if undoing {
		stack = um.redoStack
	} else {
		stack = um.undoStack
	}

	if undoing {
		// Next undo should not be appended to last stack item.
		um.StopCapturing()
	} else if !redoing {
		// Neither undoing nor redoing: delete redoStack.
		um.redoStack.Init()
	}

	beforeState := transaction.BeforeState
	afterState := transaction.AfterState

	now := time.Now()
	if now.Sub(um.lastChange).Milliseconds() < int64(um.captureTimeout) && stack.Len() > 0 && !undoing && !redoing {
		// Append change to last stack op.
		lastOp := stack.Back().Value.(*StackItem)
		dss := []*DeleteSet{lastOp.DeleteSet, transaction.DeleteSet}
		lastOp.DeleteSet = NewDeleteSetFromList(dss)
		lastOp.AfterState = afterState
	} else {
		// Create a new stack op.
		item := NewStackItem(transaction.DeleteSet, beforeState, afterState)
		stack.PushBack(item)
	}

	if !undoing && !redoing {
		um.lastChange = now
	}

	// Make sure that deleted structs are not GC'd.
	transaction.DeleteSet.IterateDeletedStructs(transaction, func(i *structs.AbstractStruct) bool {
		if item, ok := i.(*structs.Item); ok {
			for _, typeItem := range um.scope {
				if um.isParentOf(typeItem, item) {
					item.KeepItemAndParents(true)
					break
				}
			}
		}
		return true
	})

	// Fire stack item added event
	if stack.Len() > 0 {
		eventArgs := &StackEventArgs{
			StackItem: stack.Back().Value.(*StackItem),
			Type: func() OperationType {
				if undoing {
					return Redo
				}
				return Undo
			}(),
			ChangedParentTypes: transaction.ChangedParentTypes,
			Origin: transaction.Origin,
		}
		
		for _, handler := range um.stackItemAddedHandlers {
			handler(eventArgs)
		}
	}
}

// popStackItem pops an item from the stack
func (um *UndoManager) popStackItem(stack *list.List, eventType OperationType) *StackItem {
	var result *StackItem
	var tr *Transaction

	um.doc.Transact(func(transaction *Transaction) {
		tr = transaction

		for stack.Len() > 0 && result == nil {
			// Get and remove the last element
			element := stack.Back()
			stack.Remove(element)
			stackItem := element.Value.(*StackItem)
			
			itemsToRedo := make(map[*structs.Item]bool)
			itemsToDelete := make([]*structs.Item, 0)
			performedChange := false

			for client, endClock := range stackItem.AfterState {
				startClock := int64(0)
				if sc, exists := stackItem.BeforeState[client]; exists {
					startClock = sc
				}

				length := endClock - startClock
				structs := um.doc.Store.Clients[client]

				if startClock != endClock {
					// Make sure structs don't overlap with the range of created operations [stackItem.start, stackItem.start + stackItem.end).
					// This must be executed before deleted structs are iterated.
					um.doc.Store.GetItemCleanStart(transaction, ID{Client: client, Clock: startClock})

					if endClock < um.doc.Store.GetState(client) {
						um.doc.Store.GetItemCleanStart(transaction, ID{Client: client, Clock: endClock})
					}

					um.doc.Store.IterateStructs(transaction, structs, startClock, length, func(str *structs.AbstractStruct) bool {
						if it, ok := str.(*structs.Item); ok {
							if it.Redone != nil {
								redoneResult, diff := um.doc.Store.FollowRedone(str.Id)
								
								if diff > 0 {
									redoneResult = um.doc.Store.GetItemCleanStart(transaction, ID{Client: redoneResult.Id.Client, Clock: redoneResult.Id.Clock + int64(diff)})
								}

								if redoneResult.Length > length {
									um.doc.Store.GetItemCleanStart(transaction, ID{Client: redoneResult.Id.Client, Clock: endClock})
								}

								str = redoneResult
								if item, ok := redoneResult.(*structs.Item); ok {
									it = item
								}
							}

							if !it.Deleted {
								for _, typeItem := range um.scope {
									if um.isParentOf(typeItem, it) {
										itemsToDelete = append(itemsToDelete, it)
										break
									}
								}
							}
						}
						return true
					})
				}
			}

			stackItem.DeleteSet.IterateDeletedStructs(transaction, func(str *structs.AbstractStruct) bool {
				id := str.Id
				clock := id.Clock
				client := id.Client

				startClock := int64(0)
				if sc, exists := stackItem.BeforeState[client]; exists {
					startClock = sc
				}

				endClock := int64(0)
				if ec, exists := stackItem.AfterState[client]; exists {
					endClock = ec
				}

				if item, ok := str.(*structs.Item); ok {
					// Check if item is in scope
					inScope := false
					for _, typeItem := range um.scope {
						if um.isParentOf(typeItem, item) {
							inScope = true
							break
						}
					}
					
					// Never redo structs in [stackItem.start, stackItem.start + stackItem.end), because they were created and deleted in the same capture interval.
					if inScope && !(clock >= startClock && clock < endClock) {
						itemsToRedo[item] = true
					}
				}
				return true
			})

			for item := range itemsToRedo {
				performedChange = transaction.RedoItem(item, itemsToRedo) != nil || performedChange
			}

			// We want to delete in reverse order so that children are deleted before
			// parents, so we have more information available when items are filtered.
			for i := len(itemsToDelete) - 1; i >= 0; i-- {
				item := itemsToDelete[i]
				if um.deleteFilter(item) {
					item.Delete(transaction)
					performedChange = true
				}
			}

			result = stackItem
		}

		for typ, subProps := range transaction.Changed {
			// Destroy search marker if necessary.
			if _, exists := subProps[""]; exists {
				if arr, ok := typ.(*types.YArrayBase); ok {
					arr.ClearSearchMarkers()
				}
			}
		}
	}, um)

	if result != nil {
		eventArgs := &StackEventArgs{
			StackItem: result,
			Type: eventType,
			ChangedParentTypes: tr.ChangedParentTypes,
			Origin: tr.Origin,
		}
		
		for _, handler := range um.stackItemPoppedHandlers {
			handler(eventArgs)
		}
	}

	return result
}

// isParentOf checks if a parent type is a parent of a child item
func (um *UndoManager) isParentOf(parent *types.AbstractType, child *structs.Item) bool {
	for child != nil {
		if child.Parent == parent {
			return true
		}

		if parentType, ok := child.Parent.(*types.AbstractType); ok && parentType.Item != nil {
			child = parentType.Item
		} else {
			break
		}
	}

	return false
}