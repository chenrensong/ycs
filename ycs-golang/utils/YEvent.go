package utils

import (
	"github.com/chenrensong/ygo/structs"
	"github.com/chenrensong/ygo/types"
)

// YEvent represents a base event for Y types
type YEvent struct {
	Target      types.AbstractType
	Transaction *Transaction
	Path        []interface{}
}

// NewYEvent creates a new YEvent
func NewYEvent(target types.AbstractType, transaction *Transaction) *YEvent {
	return &YEvent{
		Target:      target,
		Transaction: transaction,
		Path:        make([]interface{}, 0),
	}
}

// GetPath gets the path to the target
func (e *YEvent) GetPath() []interface{} {
	path := make([]interface{}, len(e.Path))
	copy(path, e.Path)
	return path
}

// AddPath adds an element to the path
func (e *YEvent) AddPath(element interface{}) {
	e.Path = append(e.Path, element)
}

// Adds checks if an item was added in this transaction
func (e *YEvent) Adds(item structs.AbstractStruct) bool {
	beforeState, exists := e.Transaction.BeforeState[item.GetId().Client]
	if !exists {
		beforeState = 0
	}
	return item.GetId().Clock >= beforeState
}

// Deletes checks if an item was deleted in this transaction
func (e *YEvent) Deletes(item structs.AbstractStruct) bool {
	return e.Transaction.DeleteSet.IsDeleted(item.GetId())
}

// GetTarget gets the target of the event
func (e *YEvent) GetTarget() types.AbstractType {
	return e.Target
}

// GetTransaction gets the transaction of the event
func (e *YEvent) GetTransaction() *Transaction {
	return e.Transaction
}
