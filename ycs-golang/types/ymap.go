package types

import (
	"github.com/yjs/ycs-golang/structs"
	"github.com/yjs/ycs-golang/utils"
)

// YMapEvent represents an event for YMap
type YMapEvent struct {
	*YEvent
	KeysChanged map[string]struct{}
}

// NewYMapEvent creates a new YMapEvent
func NewYMapEvent(m *YMap, transaction *utils.Transaction, subs map[string]struct{}) *YMapEvent {
	return &YMapEvent{
		YEvent:      NewYEvent(m.AbstractType, transaction),
		KeysChanged: subs,
	}
}

// YMap represents a map type
type YMap struct {
	*AbstractType
	PrelimContent map[string]interface{}
}

// YMapRefId is the reference ID for YMap
const YMapRefId int = 1

// NewYMap creates a new YMap
func NewYMap(entries map[string]interface{}) *YMap {
	content := make(map[string]interface{})
	if entries != nil {
		for k, v := range entries {
			content[k] = v
		}
	}
	
	return &YMap{
		AbstractType:  NewAbstractType(),
		PrelimContent: content,
	}
}

// NewYMapEmpty creates a new empty YMap
func NewYMapEmpty() *YMap {
	return NewYMap(nil)
}

// Count returns the number of items in the map
func (y *YMap) Count() int {
	if y.PrelimContent != nil {
		return len(y.PrelimContent)
	}
	return len(y.TypeMapEnumerate())
}

// Get gets a value by key
func (y *YMap) Get(key string) (interface{}, error) {
	if value, ok := y.TryTypeMapGet(key); ok {
		return value, nil
	}
	
	// In Go, we return an error instead of throwing an exception
	return nil, &KeyNotFoundError{Key: key}
}

// Set sets a value by key
func (y *YMap) Set(key string, value interface{}) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			y.TypeMapSet(tr, key, value)
		})
	} else {
		y.PrelimContent[key] = value
	}
}

// Delete deletes a value by key
func (y *YMap) Delete(key string) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			y.TypeMapDelete(tr, key)
		})
	} else {
		delete(y.PrelimContent, key)
	}
}

// ContainsKey checks if a key exists in the map
func (y *YMap) ContainsKey(key string) bool {
	if val, ok := y.Map[key]; ok {
		return !val.Deleted()
	}
	return false
}

// Keys returns the keys of the map
func (y *YMap) Keys() []string {
	enumerated := y.TypeMapEnumerate()
	keys := make([]string, 0, len(enumerated))
	
	for key := range enumerated {
		keys = append(keys, key)
	}
	
	return keys
}

// Values returns the values of the map
func (y *YMap) Values() []interface{} {
	enumerated := y.TypeMapEnumerate()
	values := make([]interface{}, 0, len(enumerated))
	
	for _, item := range enumerated {
		content := item.Content.GetContent()
		values = append(values, content[item.Length-1])
	}
	
	return values
}

// Clone creates a clone of the map
func (y *YMap) Clone() *YMap {
	// In a real implementation, you would need to cast the result
	// For now, we'll just return nil as a placeholder
	return nil
}

// InternalCopy creates an internal copy of the map
func (y *YMap) InternalCopy() *AbstractType {
	return NewYMapEmpty().AbstractType
}

// InternalClone creates an internal clone of the map
func (y *YMap) InternalClone() *AbstractType {
	m := NewYMapEmpty()
	
	for key, item := range y.TypeMapEnumerate() {
		// TODO: [alekseyk] Yjs checks for the AbstractType here, but _map can only have 'Item' values. Might be an error?
		m.Set(key, item)
	}
	
	return m.AbstractType
}

// Integrate integrates the map
func (y *YMap) Integrate(doc *utils.YDoc, item *structs.Item) {
	y.AbstractType.Integrate(doc, item)
	
	for key, value := range y.PrelimContent {
		y.Set(key, value)
	}
	
	y.PrelimContent = nil
}

// CallObserver creates YMapEvent and calls observers
func (y *YMap) CallObserver(transaction *utils.Transaction, parentSubs map[string]struct{}) {
	y.CallTypeObservers(transaction, NewYMapEvent(y, transaction, parentSubs).YEvent)
}

// Write writes the map to an encoder
func (y *YMap) Write(encoder utils.IUpdateEncoder) {
	encoder.WriteTypeRef(YMapRefId)
}

// Read reads a YMap from a decoder
func ReadYMap(decoder utils.IUpdateDecoder) *YMap {
	return NewYMapEmpty()
}

// KeyNotFoundError represents an error when a key is not found
type KeyNotFoundError struct {
	Key string
}

func (e *KeyNotFoundError) Error() string {
	return "key not found: " + e.Key
}