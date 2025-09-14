package core

import (
	"ycs/contracts"
)

// YEventArgs represents event arguments for Y events
type YEventArgs struct {
	Event       contracts.IYEvent
	Transaction contracts.ITransaction
}

// NewYEventArgs creates new YEventArgs
func NewYEventArgs(evt contracts.IYEvent, transaction contracts.ITransaction) YEventArgs {
	return YEventArgs{
		Event:       evt,
		Transaction: transaction,
	}
}

// YDeepEventArgs represents deep event arguments
type YDeepEventArgs struct {
	Events      []contracts.IYEvent
	Transaction contracts.ITransaction
}

// NewYDeepEventArgs creates new YDeepEventArgs
func NewYDeepEventArgs(events []contracts.IYEvent, transaction contracts.ITransaction) YDeepEventArgs {
	return YDeepEventArgs{
		Events:      events,
		Transaction: transaction,
	}
}

// Delta represents a change delta
type Delta struct {
	Delete     *int
	Retain     *int
	Insert     []interface{}
	Attributes map[string]interface{}
}

// ChangeKey represents a key change
type ChangeKey struct {
	Action   string
	OldValue interface{}
	NewValue interface{}
}

// ChangesCollection represents a collection of changes
type ChangesCollection struct {
	Added   map[contracts.IStructItem]struct{}
	Deleted map[contracts.IStructItem]struct{}
	Delta   []Delta
	Keys    map[string]ChangeKey
}

// YEvent represents a Y event
type YEvent struct {
	target        contracts.IAbstractType
	currentTarget contracts.IAbstractType
	transaction   contracts.ITransaction
	changes       *ChangesCollection
}

// NewYEvent creates a new YEvent
func NewYEvent(target contracts.IAbstractType, transaction contracts.ITransaction) *YEvent {
	return &YEvent{
		target:        target,
		currentTarget: target,
		transaction:   transaction,
	}
}

// GetTarget returns the target
func (ye *YEvent) GetTarget() contracts.IAbstractType {
	return ye.target
}

// SetTarget sets the target
func (ye *YEvent) SetTarget(target contracts.IAbstractType) {
	ye.target = target
}

// GetCurrentTarget returns the current target
func (ye *YEvent) GetCurrentTarget() contracts.IAbstractType {
	return ye.currentTarget
}

// SetCurrentTarget sets the current target
func (ye *YEvent) SetCurrentTarget(currentTarget contracts.IAbstractType) {
	ye.currentTarget = currentTarget
}

// GetTransaction returns the transaction
func (ye *YEvent) GetTransaction() contracts.ITransaction {
	return ye.transaction
}

// SetTransaction sets the transaction
func (ye *YEvent) SetTransaction(transaction contracts.ITransaction) {
	ye.transaction = transaction
}

// GetPath returns the path from current target to target
func (ye *YEvent) GetPath() []interface{} {
	return ye.getPathTo(ye.currentTarget, ye.target)
}

// GetChanges returns the changes collection
func (ye *YEvent) GetChanges() *contracts.ChangesCollection {
	changes := ye.collectChanges()
	// Convert to contracts.ChangesCollection
	return &contracts.ChangesCollection{
		Added:   changes.Added,
		Deleted: changes.Deleted,
		Delta:   convertDelta(changes.Delta),
		Keys:    convertKeys(changes.Keys),
	}
}

// Helper function to convert Delta slice
func convertDelta(delta []Delta) []contracts.Delta {
	result := make([]contracts.Delta, len(delta))
	for i, d := range delta {
		result[i] = contracts.Delta{
			Delete:     d.Delete,
			Retain:     d.Retain,
			Insert:     d.Insert,
			Attributes: d.Attributes,
		}
	}
	return result
}

// Helper function to convert Keys map
func convertKeys(keys map[string]ChangeKey) map[string]contracts.ChangeKey {
	result := make(map[string]contracts.ChangeKey)
	for k, v := range keys {
		// Convert string action to ChangeAction enum
		var action contracts.ChangeAction
		switch v.Action {
		case "Add":
			action = contracts.ChangeActionAdd
		case "Update":
			action = contracts.ChangeActionUpdate
		case "Delete":
			action = contracts.ChangeActionDelete
		default:
			action = contracts.ChangeActionUpdate
		}

		result[k] = contracts.ChangeKey{
			Action:   action,
			OldValue: v.OldValue,
		}
	}
	return result
}

// Deletes checks if a struct is deleted by this event
func (ye *YEvent) deletes(str contracts.IStructItem) bool {
	return ye.transaction.GetDeleteSet().IsDeleted(str.GetID())
}

// Adds checks if a struct is added by this event
func (ye *YEvent) adds(str contracts.IStructItem) bool {
	beforeClock, exists := ye.transaction.GetBeforeState()[str.GetID().Client]
	return !exists || str.GetID().Clock >= beforeClock
}

// collectChanges collects all changes for this event
func (ye *YEvent) collectChanges() *ChangesCollection {
	if ye.changes == nil {
		added := make(map[contracts.IStructItem]struct{})
		deleted := make(map[contracts.IStructItem]struct{})
		delta := make([]Delta, 0)
		keys := make(map[string]ChangeKey)

		ye.changes = &ChangesCollection{
			Added:   added,
			Deleted: deleted,
			Delta:   delta,
			Keys:    keys,
		}

		changed, exists := ye.transaction.GetChanged()[ye.target]
		if !exists {
			changed = make(map[string]struct{})
			ye.transaction.GetChanged()[ye.target] = changed
		}

		// Check if null key exists (indicating array changes)
		if _, hasNullKey := changed[""]; hasNullKey {
			var lastOp *Delta

			packOp := func() {
				if lastOp != nil {
					delta = append(delta, *lastOp)
					ye.changes.Delta = delta
				}
			}

			for item := ye.target.GetStart(); item != nil; item = item.GetRight() {
				if item.GetDeleted() {
					if ye.deletes(item) && !ye.adds(item) {
						if lastOp == nil || lastOp.Delete == nil {
							packOp()
							deleteVal := 0
							lastOp = &Delta{Delete: &deleteVal}
						}
						*lastOp.Delete += item.GetLength()
						deleted[item] = struct{}{}
					}
					// else: do nothing for items that were added and then deleted in the same transaction
				} else {
					if ye.adds(item) {
						if lastOp == nil || lastOp.Insert == nil {
							packOp()
							lastOp = &Delta{Insert: make([]interface{}, 0)}
						}
						content := item.GetContent().GetContent()
						lastOp.Insert = append(lastOp.Insert, content...)
						added[item] = struct{}{}
					} else {
						if lastOp == nil || lastOp.Retain == nil {
							packOp()
							retainVal := 0
							lastOp = &Delta{Retain: &retainVal}
						}
						*lastOp.Retain += item.GetLength()
					}
				}
			}

			if lastOp != nil && (lastOp.Retain != nil && *lastOp.Retain > 0) {
				packOp()
			}
		}

		// Handle key changes for maps
		for key := range changed {
			if key != "" { // Skip the null key we handled above
				// Simplified key change handling
				keys[key] = ChangeKey{
					Action: "update", // This would need more sophisticated logic
				}
			}
		}
	}

	return ye.changes
}

// getPathTo returns the path from one type to another
func (ye *YEvent) getPathTo(parent, child contracts.IAbstractType) []interface{} {
	path := make([]interface{}, 0)
	// This is a simplified implementation
	// In a full implementation, you'd traverse the parent-child relationship
	return path
}
