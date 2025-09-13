package utils

import "sync"

// UndoStackItem represents an item in the undo stack
type UndoStackItem struct {
	DeleteSet   *DeleteSet
	BeforeState map[uint64]int
	AfterState  map[uint64]int
	Meta        map[string]interface{}
}

// NewUndoStackItem creates a new undo stack item
func NewUndoStackItem(deleteSet *DeleteSet, beforeState, afterState map[uint64]int) *UndoStackItem {
	return &UndoStackItem{
		DeleteSet:   deleteSet,
		BeforeState: beforeState,
		AfterState:  afterState,
		Meta:        make(map[string]interface{}),
	}
}

// UndoManager manages undo/redo operations
type UndoManager struct {
	doc            *YDoc
	scope          []IAbstractType
	undoStack      []*UndoStackItem
	redoStack      []*UndoStackItem
	undoing        bool
	redoing        bool
	trackedOrigins map[interface{}]struct{}
	mutex          sync.RWMutex
}

// NewUndoManager creates a new undo manager
func NewUndoManager(doc *YDoc, scope []IAbstractType) *UndoManager {
	um := &UndoManager{
		doc:            doc,
		scope:          scope,
		undoStack:      make([]*UndoStackItem, 0),
		redoStack:      make([]*UndoStackItem, 0),
		undoing:        false,
		redoing:        false,
		trackedOrigins: make(map[interface{}]struct{}),
	}

	// Register transaction observer
	doc.Observe("transaction", um.onTransaction)

	return um
}

// AddToScope adds a type to the undo manager scope
func (um *UndoManager) AddToScope(type_ IAbstractType) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	um.scope = append(um.scope, type_)
}

// Undo performs an undo operation
func (um *UndoManager) Undo() bool {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	if len(um.undoStack) == 0 {
		return false
	}

	um.undoing = true
	defer func() { um.undoing = false }()

	// Get the last undo item
	undoItem := um.undoStack[len(um.undoStack)-1]
	um.undoStack = um.undoStack[:len(um.undoStack)-1]

	// Perform undo operation
	um.doc.Transact(func(tr *Transaction) {
		// Restore the before state
		// This is a simplified implementation
		for client, state := range undoItem.BeforeState {
			tr.BeforeState[client] = state
		}
	}, "undo")

	// Move to redo stack
	um.redoStack = append(um.redoStack, undoItem)

	return true
}

// Redo performs a redo operation
func (um *UndoManager) Redo() bool {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	if len(um.redoStack) == 0 {
		return false
	}

	um.redoing = true
	defer func() { um.redoing = false }()

	// Get the last redo item
	redoItem := um.redoStack[len(um.redoStack)-1]
	um.redoStack = um.redoStack[:len(um.redoStack)-1]

	// Perform redo operation
	um.doc.Transact(func(tr *Transaction) {
		// Restore the after state
		// This is a simplified implementation
		for client, state := range redoItem.AfterState {
			tr.AfterState[client] = state
		}
	}, "redo")

	// Move to undo stack
	um.undoStack = append(um.undoStack, redoItem)

	return true
}

// Clear clears both undo and redo stacks
func (um *UndoManager) Clear() {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	um.undoStack = make([]*UndoStackItem, 0)
	um.redoStack = make([]*UndoStackItem, 0)
}

// CanUndo returns true if undo is possible
func (um *UndoManager) CanUndo() bool {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	return len(um.undoStack) > 0
}

// CanRedo returns true if redo is possible
func (um *UndoManager) CanRedo() bool {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	return len(um.redoStack) > 0
}

// onTransaction is called when a transaction occurs
func (um *UndoManager) onTransaction(tr *Transaction, doc *YDoc) {
	if um.undoing || um.redoing {
		return
	}

	// Check if any of the tracked types were changed
	hasRelevantChanges := false
	for type_ := range tr.Changed {
		for _, scopeType := range um.scope {
			if type_ == scopeType {
				hasRelevantChanges = true
				break
			}
		}
		if hasRelevantChanges {
			break
		}
	}

	if !hasRelevantChanges {
		return
	}

	// Create undo stack item
	undoItem := NewUndoStackItem(tr.DeleteSet, tr.BeforeState, tr.AfterState)

	um.mutex.Lock()
	defer um.mutex.Unlock()

	// Add to undo stack
	um.undoStack = append(um.undoStack, undoItem)

	// Clear redo stack
	um.redoStack = make([]*UndoStackItem, 0)
}
