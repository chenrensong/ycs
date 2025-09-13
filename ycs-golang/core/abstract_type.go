// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"github.com/chenrensong/ygo/contracts"
)

// YEventArgs represents event arguments for Y events
type YEventArgs struct {
	Event       contracts.IYEvent
	Transaction contracts.ITransaction
}

// NewYEventArgs creates a new YEventArgs
func NewYEventArgs(evt contracts.IYEvent, transaction contracts.ITransaction) *YEventArgs {
	return &YEventArgs{
		Event:       evt,
		Transaction: transaction,
	}
}

// YDeepEventArgs represents deep event arguments for Y events
type YDeepEventArgs struct {
	Events      []contracts.IYEvent
	Transaction contracts.ITransaction
}

// NewYDeepEventArgs creates a new YDeepEventArgs
func NewYDeepEventArgs(events []contracts.IYEvent, transaction contracts.ITransaction) *YDeepEventArgs {
	return &YDeepEventArgs{
		Events:      events,
		Transaction: transaction,
	}
}

// AbstractType is the base implementation of IAbstractType
type AbstractType struct {
	item              contracts.IStructItem
	start             contracts.IStructItem
	mapItems          map[string]contracts.IStructItem
	eventHandlers     []func(*YEventArgs)
	deepEventHandlers []func(*YDeepEventArgs)
	doc               contracts.IYDoc
	length            int
}

var contentFactory contracts.IContentFactory

// NewAbstractType creates a new AbstractType
func NewAbstractType() *AbstractType {
	return &AbstractType{
		mapItems:          make(map[string]contracts.IStructItem),
		eventHandlers:     make([]func(*YEventArgs), 0),
		deepEventHandlers: make([]func(*YDeepEventArgs), 0),
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
	return at.mapItems
}

// SetMap sets the map
func (at *AbstractType) SetMap(mapItems map[string]contracts.IStructItem) {
	at.mapItems = mapItems
}

// GetDoc returns the document
func (at *AbstractType) GetDoc() contracts.IYDoc {
	return at.doc
}

// GetParent returns the parent
func (at *AbstractType) GetParent() contracts.IAbstractType {
	if at.item != nil {
		return at.item.GetParent()
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

// InternalCopy creates an internal copy - to be implemented by subclasses
func (at *AbstractType) InternalCopy() contracts.IAbstractType {
	panic("InternalCopy must be implemented by subclasses")
}

// InternalClone creates an internal clone - to be implemented by subclasses
func (at *AbstractType) InternalClone() contracts.IAbstractType {
	panic("InternalClone must be implemented by subclasses")
}

// Write writes the type to an encoder - to be implemented by subclasses
func (at *AbstractType) Write(encoder contracts.IUpdateEncoder) {
	panic("Write must be implemented by subclasses")
}

// CallTypeObservers calls event listeners with an event. This will also add an event to all parents
// for observeDeep handlers.
func (at *AbstractType) CallTypeObservers(transaction contracts.ITransaction, evt contracts.IYEvent) {
	currentType := at

	for {
		values := transaction.GetChangedParentTypes()[currentType]
		if values == nil {
			values = make([]contracts.IYEvent, 0)
			transaction.GetChangedParentTypes()[currentType] = values
		}

		values = append(values, evt)
		transaction.GetChangedParentTypes()[currentType] = values

		if currentType.item == nil {
			break
		}

		parent := currentType.item.GetParent()
		if parent == nil {
			break
		}

		if parentType, ok := parent.(*AbstractType); ok {
			currentType = parentType
		} else {
			break
		}
	}

	at.InvokeEventHandlers(evt, transaction)
}

// CallObserver creates YEvent and calls all type observers.
// Must be implemented by each type.
func (at *AbstractType) CallObserver(transaction contracts.ITransaction, parentSubs map[string]struct{}) {
	// Do nothing in base implementation
}

// First returns the first non-deleted item
func (at *AbstractType) First() contracts.IStructItem {
	n := at.start
	for n != nil && n.GetDeleted() {
		n = n.GetRight()
	}
	return n
}

// InvokeEventHandlers invokes the event handlers
func (at *AbstractType) InvokeEventHandlers(evt contracts.IYEvent, transaction contracts.ITransaction) {
	args := NewYEventArgs(evt, transaction)
	for _, handler := range at.eventHandlers {
		handler(args)
	}
}

// CallDeepEventHandlerListeners calls deep event handler listeners
func (at *AbstractType) CallDeepEventHandlerListeners(events []contracts.IYEvent, transaction contracts.ITransaction) {
	args := NewYDeepEventArgs(events, transaction)
	for _, handler := range at.deepEventHandlers {
		handler(args)
	}
}

// FindRootTypeKey finds the root type key
func (at *AbstractType) FindRootTypeKey() string {
	return at.doc.FindRootTypeKey(at)
}

// TypeMapDelete deletes from the type map
func (at *AbstractType) TypeMapDelete(transaction contracts.ITransaction, key string) {
	if c, exists := at.mapItems[key]; exists {
		c.Delete(transaction)
	}
}

// TypeMapSet sets a value in the type map
func (at *AbstractType) TypeMapSet(transaction contracts.ITransaction, key string, value interface{}) {
	var left contracts.IStructItem
	if l, exists := at.mapItems[key]; exists {
		left = l
	}

	doc := transaction.GetDoc()
	ownClientId := doc.GetClientID()
	content := GetContentFactory().CreateContent(value)

	var lastId *contracts.StructID
	if left != nil {
		lastId = left.GetLastID()
	}

	newItem := NewStructItem(
		contracts.StructID{Client: ownClientId, Clock: doc.GetStore().GetState(ownClientId)},
		left,
		lastId,
		nil,
		nil,
		at,
		key,
		content,
	)
	newItem.Integrate(transaction, 0)
}

// TryTypeMapGet tries to get a value from the type map
func (at *AbstractType) TryTypeMapGet(key string) (interface{}, bool) {
	if val, exists := at.mapItems[key]; exists && !val.GetDeleted() {
		content := val.GetContent().GetContent()
		if len(content) > 0 {
			return content[val.GetLength()-1], true
		}
	}
	return nil, false
}

// TypeMapGetSnapshot gets a value from the type map at a specific snapshot
func (at *AbstractType) TypeMapGetSnapshot(key string, snapshot contracts.ISnapshot) interface{} {
	var v contracts.IStructItem
	if val, exists := at.mapItems[key]; exists {
		v = val
	}

	stateVector := snapshot.GetStateVector()
	for v != nil {
		clientState, hasClient := stateVector[v.GetId().Client]
		if !hasClient || v.GetId().Clock >= clientState {
			v = v.GetLeft()
		} else {
			break
		}
	}

	if v != nil && v.IsVisible(snapshot) {
		content := v.GetContent().GetContent()
		if len(content) > 0 {
			return content[v.GetLength()-1]
		}
	}
	return nil
}

// TypeMapEnumerate enumerates the type map
func (at *AbstractType) TypeMapEnumerate() map[string]contracts.IStructItem {
	result := make(map[string]contracts.IStructItem)
	for key, value := range at.mapItems {
		if !value.GetDeleted() {
			result[key] = value
		}
	}
	return result
}

// TypeMapEnumerateValues enumerates the values in the type map
func (at *AbstractType) TypeMapEnumerateValues() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range at.mapItems {
		if !value.GetDeleted() {
			content := value.GetContent().GetContent()
			if len(content) > 0 {
				result[key] = content[value.GetLength()-1]
			}
		}
	}
	return result
}

// AddEventHandler adds an event handler
func (at *AbstractType) AddEventHandler(handler func(*YEventArgs)) {
	at.eventHandlers = append(at.eventHandlers, handler)
}

// AddDeepEventHandler adds a deep event handler
func (at *AbstractType) AddDeepEventHandler(handler func(*YDeepEventArgs)) {
	at.deepEventHandlers = append(at.deepEventHandlers, handler)
}

// SetContentFactory sets the content factory
func SetContentFactory(factory contracts.IContentFactory) {
	contentFactory = factory
}

// GetContentFactory gets the content factory
func GetContentFactory() contracts.IContentFactory {
	return contentFactory
}
