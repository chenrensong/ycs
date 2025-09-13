// ------------------------------------------------------------------------------
//  Copyright (c) Microsoft Corporation.  All rights reserved.
// ------------------------------------------------------------------------------

package types

import (
	"github.com/chenrensong/ygo/contracts"
)

// YEvent represents an event that describes changes on a Y type
type YEvent struct {
	Target        contracts.IAbstractType
	CurrentTarget contracts.IAbstractType
	Transaction   contracts.ITransaction
	changes       *contracts.ChangesCollection
}

// NewYEvent creates a new YEvent instance
func NewYEvent(target contracts.IAbstractType, transaction contracts.ITransaction) *YEvent {
	return &YEvent{
		Target:        target,
		CurrentTarget: target,
		Transaction:   transaction,
		changes:       nil,
	}
}

// Path returns the path from CurrentTarget to Target
func (ye *YEvent) Path() []interface{} {
	return ye.getPathTo(ye.CurrentTarget, ye.Target)
}

// Changes returns the changes collection
func (ye *YEvent) Changes() *contracts.ChangesCollection {
	return ye.collectChanges()
}

// Deletes checks if a struct is deleted by this event
func (ye *YEvent) Deletes(str contracts.IStructItem) bool {
	return ye.Transaction.DeleteSet().IsDeleted(str.ID())
}

// Adds checks if a struct is added by this event
func (ye *YEvent) Adds(str contracts.IStructItem) bool {
	if clock, exists := ye.Transaction.BeforeState()[str.ID().Client]; exists {
		return str.ID().Clock >= clock
	}
	return true
}

// collectChanges computes and caches the changes collection
func (ye *YEvent) collectChanges() *contracts.ChangesCollection {
	if ye.changes == nil {
		target := ye.Target
		added := make(map[contracts.IStructItem]bool)
		deleted := make(map[contracts.IStructItem]bool)
		delta := make([]*contracts.Delta, 0)
		keys := make(map[string]*contracts.ChangeKey)

		ye.changes = &contracts.ChangesCollection{
			Added:   added,
			Deleted: deleted,
			Delta:   delta,
			Keys:    keys,
		}

		changed, exists := ye.Transaction.Changed()[target]
		if !exists {
			changed = make(map[string]bool)
			ye.Transaction.Changed()[target] = changed
		}

		if changed[""] { // Check if null key exists (represents content changes)
			var lastOp *contracts.Delta

			packOp := func() {
				if lastOp != nil {
					delta = append(delta, lastOp)
					ye.changes.Delta = delta
				}
			}

			for item := target.Start(); item != nil; item = item.Right().(contracts.IStructItem) {
				if item.Deleted() {
					if ye.Deletes(item) && !ye.Adds(item) {
						if lastOp == nil || lastOp.Delete == nil {
							packOp()
							deleteLen := 0
							lastOp = &contracts.Delta{Delete: &deleteLen}
						}
						*lastOp.Delete += item.Length()
						deleted[item] = true
					}
				} else {
					if ye.Adds(item) {
						if lastOp == nil || lastOp.Insert == nil {
							packOp()
							insertList := make([]interface{}, 0, 1)
							lastOp = &contracts.Delta{Insert: &insertList}
						}
						content := item.Content().GetContent()
						*lastOp.Insert = append(*lastOp.Insert, content...)
						added[item] = true
					} else {
						if lastOp == nil || lastOp.Retain == nil {
							packOp()
							retainLen := 0
							lastOp = &contracts.Delta{Retain: &retainLen}
						}
						*lastOp.Retain += item.Length()
					}
				}
			}

			if lastOp != nil && lastOp.Retain == nil {
				packOp()
			}
		}

		for key := range changed {
			if key != "" {
				var action contracts.ChangeAction
				var oldValue interface{}
				item := target.Map()[key]

				if ye.Adds(item) {
					prev := item.Left()
					for prev != nil && ye.Adds(prev.(contracts.IStructItem)) {
						prev = prev.(contracts.IStructItem).Left()
					}

					if ye.Deletes(item) {
						if prev != nil && ye.Deletes(prev.(contracts.IStructItem)) {
							action = contracts.ChangeActionDelete
							prevContent := prev.(contracts.IStructItem).Content().GetContent()
							oldValue = prevContent[len(prevContent)-1]
						} else {
							continue
						}
					} else {
						if prev != nil && ye.Deletes(prev.(contracts.IStructItem)) {
							action = contracts.ChangeActionUpdate
							prevContent := prev.(contracts.IStructItem).Content().GetContent()
							oldValue = prevContent[len(prevContent)-1]
						} else {
							action = contracts.ChangeActionAdd
							oldValue = nil
						}
					}
				} else {
					if ye.Deletes(item) {
						action = contracts.ChangeActionDelete
						itemContent := item.Content().GetContent()
						oldValue = itemContent[len(itemContent)-1]
					} else {
						continue
					}
				}

				keys[key] = &contracts.ChangeKey{
					Action:   action,
					OldValue: oldValue,
				}
			}
		}
	}

	return ye.changes
}

// getPathTo computes the path from parent to child
func (ye *YEvent) getPathTo(parent contracts.IAbstractType, child contracts.IAbstractType) []interface{} {
	path := make([]interface{}, 0)

	for child.Item() != nil && child != parent {
		if child.Item().ParentSub() != "" {
			// Parent is map-ish
			path = append([]interface{}{child.Item().ParentSub()}, path...)
		} else {
			// Parent is array-ish
			i := 0
			c := child.Item().Parent().(contracts.IAbstractType).Start()
			for c != child.Item() && c != nil {
				if !c.(contracts.IStructItem).Deleted() {
					i++
				}
				c = c.(contracts.IStructItem).Right()
			}
			path = append([]interface{}{i}, path...)
		}

		child = child.Item().Parent().(contracts.IAbstractType)
	}

	return path
}
