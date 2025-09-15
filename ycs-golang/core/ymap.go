package core

import (
	"ycs/contracts"
)

const YMapRefID = 1

// YMapEvent represents an event that describes changes on a YMap
type YMapEvent struct {
	*YEvent
	KeysChanged map[string]struct{}
}

// NewYMapEvent creates a new YMapEvent
func NewYMapEvent(ymap *YMap, transaction contracts.ITransaction, subs map[string]struct{}) *YMapEvent {
	return &YMapEvent{
		YEvent:      NewYEvent(ymap, transaction),
		KeysChanged: subs,
	}
}

// YMap represents a shared Map implementation
type YMap struct {
	*AbstractType
	prelimContent map[string]interface{}
}

// NewYMap creates a new YMap
func NewYMap(entries map[string]interface{}) *YMap {
	ym := &YMap{
		AbstractType:  NewAbstractType(),
		prelimContent: make(map[string]interface{}),
	}

	if entries != nil {
		for k, v := range entries {
			ym.prelimContent[k] = v
		}
	}

	return ym
}

// GetCount returns the number of entries in the map
func (ym *YMap) GetCount() int {
	if ym.prelimContent != nil {
		return len(ym.prelimContent)
	}

	count := 0
	for _, value := range ym.GetMap() {
		if !value.GetDeleted() {
			count++
		}
	}
	return count
}

// Get returns the value for the specified key
func (ym *YMap) Get(key string) interface{} {
	value, exists := ym.tryTypeMapGet(key)
	if !exists {
		return nil
	}
	return value
}

// Set sets a value for the specified key
func (ym *YMap) Set(key string, value interface{}) {
	if ym.GetDoc() != nil {
		ym.GetDoc().Transact(func(tr contracts.ITransaction) {
			ym.typeMapSet(tr, key, value)
		}, nil, true)
	} else {
		ym.prelimContent[key] = value
	}
}

// Delete removes the specified key from the map
func (ym *YMap) Delete(key string) {
	if ym.GetDoc() != nil {
		ym.GetDoc().Transact(func(tr contracts.ITransaction) {
			ym.typeMapDelete(tr, key)
		}, nil, true)
	} else {
		delete(ym.prelimContent, key)
	}
}

// ContainsKey checks if the map contains the specified key
func (ym *YMap) ContainsKey(key string) bool {
	val, exists := ym.GetMap()[key]
	return exists && !val.GetDeleted()
}

// Keys returns all keys in the map
func (ym *YMap) Keys() []string {
	var keys []string
	for key := range ym.typeMapEnumerate() {
		keys = append(keys, key)
	}
	return keys
}

// Values returns all values in the map
func (ym *YMap) Values() []interface{} {
	var values []interface{}
	for _, item := range ym.typeMapEnumerate() {
		content := item.GetContent().GetContent()
		values = append(values, content[item.GetLength()-1])
	}
	return values
}

// Clone creates a clone of the YMap
func (ym *YMap) Clone() contracts.IYMap {
	return ym.InternalClone().(contracts.IYMap)
}

// InternalCopy creates an internal copy
func (ym *YMap) InternalCopy() contracts.IAbstractType {
	return NewYMap(nil)
}

// InternalClone creates an internal clone
func (ym *YMap) InternalClone() contracts.IAbstractType {
	ymap := NewYMap(nil)

	for key, item := range ym.typeMapEnumerate() {
		// TODO: Check if this should handle AbstractType cloning
		ymap.Set(key, item)
	}

	return ymap
}

// Integrate integrates the map with a document and item
func (ym *YMap) Integrate(doc contracts.IYDoc, item contracts.IStructItem) {
	ym.AbstractType.Integrate(doc, item)

	for key, value := range ym.prelimContent {
		ym.Set(key, value)
	}

	ym.prelimContent = nil
}

// CallObserver creates YMapEvent and calls observers
func (ym *YMap) CallObserver(transaction contracts.ITransaction, parentSubs map[string]struct{}) {
	ym.CallTypeObservers(transaction, NewYMapEvent(ym, transaction, parentSubs))
}

// Write writes the map to an encoder
func (ym *YMap) Write(encoder contracts.IUpdateEncoder) {
	encoder.WriteTypeRef(YMapRefID)
}

// ReadYMap reads a YMap from decoder
func ReadYMap(decoder contracts.IUpdateDecoder) contracts.IYMap {
	return NewYMap(nil)
}

// Entries returns all key-value pairs in the map
func (ym *YMap) Entries() map[string]interface{} {
	return ym.typeMapEnumerateValues()
}

// GetEnumerator returns all key-value pairs (implements IYMap interface)
func (ym *YMap) GetEnumerator() map[string]interface{} {
	return ym.typeMapEnumerateValues()
}
