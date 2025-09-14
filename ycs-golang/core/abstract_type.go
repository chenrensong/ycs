package core

import (
	"ycs/content"
	"ycs/contracts"
)

// AbstractType represents the base type for all Y types
type AbstractType struct {
	item             contracts.IStructItem
	start            contracts.IStructItem
	m                map[string]contracts.IStructItem
	doc              contracts.IYDoc
	length           int
	eventHandler     func(contracts.YEventArgs)
	deepEventHandler func(contracts.YDeepEventArgs)
}

// NewAbstractType creates a new AbstractType
func NewAbstractType() *AbstractType {
	return &AbstractType{
		m: make(map[string]contracts.IStructItem),
	}
}

// GetItem returns the item
func (at *AbstractType) GetItem() contracts.IStructItem {
	return at.item
}

// SetItem sets the item
func (at *AbstractType) SetItem(item contracts.IStructItem) {
	at.item = item
}

// GetStart returns the start item
func (at *AbstractType) GetStart() contracts.IStructItem {
	return at.start
}

// SetStart sets the start item
func (at *AbstractType) SetStart(start contracts.IStructItem) {
	at.start = start
}

// GetMap returns the map
func (at *AbstractType) GetMap() map[string]contracts.IStructItem {
	return at.m
}

// SetMap sets the map
func (at *AbstractType) SetMap(m map[string]contracts.IStructItem) {
	at.m = m
}

// GetDoc returns the document
func (at *AbstractType) GetDoc() contracts.IYDoc {
	return at.doc
}

// GetParent returns the parent
func (at *AbstractType) GetParent() contracts.IAbstractType {
	if at.item != nil {
		return at.item.GetParent().(contracts.IAbstractType)
	}
	return nil
}

// GetLength returns the length
func (at *AbstractType) GetLength() int {
	return at.length
}

// SetLength sets the length
func (at *AbstractType) SetLength(length int) {
	at.length = length
}

// Integrate integrates the type with a document and item
func (at *AbstractType) Integrate(doc contracts.IYDoc, item contracts.IStructItem) {
	at.doc = doc
	at.item = item
}

// InternalCopy creates an internal copy (to be overridden by subclasses)
func (at *AbstractType) InternalCopy() contracts.IAbstractType {
	panic("InternalCopy not implemented")
}

// InternalClone creates an internal clone (to be overridden by subclasses)
func (at *AbstractType) InternalClone() contracts.IAbstractType {
	panic("InternalClone not implemented")
}

// Write writes the type to an encoder (to be overridden by subclasses)
func (at *AbstractType) Write(encoder contracts.IUpdateEncoder) {
	panic("Write not implemented")
}

// CallTypeObservers calls event listeners with an event. This will also add an event to all parents
// for observeDeep handlers.
func (at *AbstractType) CallTypeObservers(transaction contracts.ITransaction, evt contracts.IYEvent) {
	currentType := at

	for {
		values, exists := transaction.GetChangedParentTypes()[currentType]
		if !exists {
			values = make([]contracts.IYEvent, 0)
			transaction.GetChangedParentTypes()[currentType] = values
		}

		transaction.GetChangedParentTypes()[currentType] = append(values, evt)

		if currentType.item == nil {
			break
		}

		currentType = currentType.item.GetParent().(*AbstractType)
	}

	at.InvokeEventHandlers(evt, transaction)
}

// CallObserver creates YEvent and calls all type observers.
// Must be implemented by each type.
func (at *AbstractType) CallObserver(transaction contracts.ITransaction, parentSubs map[string]struct{}) {
	// Default implementation does nothing
}

// First returns the first non-deleted item
func (at *AbstractType) First() contracts.IStructItem {
	n := at.start
	for n != nil && n.GetDeleted() {
		n = n.GetRight()
	}
	return n
}

// InvokeEventHandlers invokes event handlers
func (at *AbstractType) InvokeEventHandlers(evt contracts.IYEvent, transaction contracts.ITransaction) {
	if at.eventHandler != nil {
		at.eventHandler(contracts.YEventArgs{Event: evt, Transaction: transaction})
	}
}

// CallDeepEventHandlerListeners calls deep event handler listeners
func (at *AbstractType) CallDeepEventHandlerListeners(events []contracts.IYEvent, transaction contracts.ITransaction) {
	if at.deepEventHandler != nil {
		at.deepEventHandler(contracts.YDeepEventArgs{Events: events, Transaction: transaction})
	}
}

// FindRootTypeKey finds the root type key
func (at *AbstractType) FindRootTypeKey() string {
	return at.doc.FindRootTypeKey(at)
}

// typeMapDelete deletes a key from the type map
func (at *AbstractType) typeMapDelete(transaction contracts.ITransaction, key string) {
	if c, exists := at.m[key]; exists {
		c.Delete(transaction)
	}
}

// typeMapSet sets a value in the type map
func (at *AbstractType) typeMapSet(transaction contracts.ITransaction, key string, value interface{}) {
	var left contracts.IStructItem
	if l, exists := at.m[key]; exists {
		left = l
	}

	doc := transaction.GetDoc()
	ownClientID := doc.GetClientID()
	contentFactory := content.MustGetGlobalFactory()
	contentObj := contentFactory.CreateContent(value)

	newItem := NewStructItem(
		StructID{Client: int64(ownClientID), Clock: doc.GetStore().GetState(int64(ownClientID))},
		left,
		FromContractsStructID(left.GetLastID()).ToPointer(),
		nil,
		nil,
		at,
		&key,
		contentObj,
	)
	newItem.Integrate(transaction, 0)
}

// tryTypeMapGet tries to get a value from the type map
func (at *AbstractType) tryTypeMapGet(key string) (interface{}, bool) {
	if val, exists := at.m[key]; exists && !val.GetDeleted() {
		content := val.GetContent().GetContent()
		return content[val.GetLength()-1], true
	}
	return nil, false
}

// typeMapGetSnapshot gets a value from the type map at a specific snapshot
func (at *AbstractType) typeMapGetSnapshot(key string, snapshot contracts.ISnapshot) interface{} {
	var v contracts.IStructItem
	if val, exists := at.m[key]; exists {
		v = val
	}

	stateVector := snapshot.GetStateVector()
	for v != nil {
		clientClock, hasClient := stateVector[v.GetID().Client]
		if !hasClient || v.GetID().Clock >= clientClock {
			v = v.GetLeft()
		} else {
			break
		}
	}

	if v != nil && v.IsVisible(snapshot) {
		content := v.GetContent().GetContent()
		return content[v.GetLength()-1]
	}
	return nil
}

// typeMapEnumerate enumerates non-deleted items in the type map
func (at *AbstractType) typeMapEnumerate() map[string]contracts.IStructItem {
	result := make(map[string]contracts.IStructItem)
	for key, value := range at.m {
		if !value.GetDeleted() {
			result[key] = value
		}
	}
	return result
}

// typeMapEnumerateValues enumerates values in the type map
func (at *AbstractType) typeMapEnumerateValues() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range at.typeMapEnumerate() {
		content := value.GetContent().GetContent()
		result[key] = content[value.GetLength()-1]
	}
	return result
}
