package types

import (
	"errors"

	"github.com/chenrensong/ygo/structs"
	"github.com/chenrensong/ygo/utils"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

// YEventArgs represents event arguments for type changes
type YEventArgs struct {
	Event       *utils.YEvent
	Transaction *utils.Transaction
}

// NewYEventArgs creates a new YEventArgs instance
func NewYEventArgs(evt *utils.YEvent, txn *utils.Transaction) *YEventArgs {
	return &YEventArgs{
		Event:       evt,
		Transaction: txn,
	}
}

// YDeepEventArgs represents deep event arguments for nested type changes
type YDeepEventArgs struct {
	Events      []*utils.YEvent
	Transaction *utils.Transaction
}

// NewYDeepEventArgs creates a new YDeepEventArgs instance
func NewYDeepEventArgs(events []*utils.YEvent, txn *utils.Transaction) *YDeepEventArgs {
	return &YDeepEventArgs{
		Events:      events,
		Transaction: txn,
	}
}

// AbstractType is the base type for all Yjs shared types
type AbstractType struct {
	item  *structs.Item
	start *structs.Item
	map_  map[string]*structs.Item

	eventHandlers     []func(*YEventArgs)
	deepEventHandlers []func(*YDeepEventArgs)
}

// NewAbstractType creates a new AbstractType instance
func NewAbstractType() *AbstractType {
	return &AbstractType{
		map_: make(map[string]*structs.Item),
	}
}

// Doc returns the document this type belongs to
func (t *AbstractType) Doc() *utils.YDoc {
	if t.item != nil {
		return t.item.Doc()
	}
	return nil
}

// Parent returns the parent type of this type
func (t *AbstractType) Parent() *AbstractType {
	if t.item != nil {
		if parent, ok := t.item.Parent().(*AbstractType); ok {
			return parent
		}
	}
	return nil
}

// Length returns the length of the type
func (t *AbstractType) Length() int {
	return 0 // To be overridden by concrete types
}

// Integrate integrates this type into the document
func (t *AbstractType) Integrate(doc *utils.YDoc, item *structs.Item) {
	// To be overridden by concrete types
}

// InternalCopy creates an internal copy of the type
func (t *AbstractType) InternalCopy() *AbstractType {
	panic(ErrNotImplemented)
}

// InternalClone creates an internal clone of the type
func (t *AbstractType) InternalClone() *AbstractType {
	panic(ErrNotImplemented)
}

// Write encodes the type to a writer
func (t *AbstractType) Write(encoder utils.Encoder) {
	panic(ErrNotImplemented)
}

// CallTypeObservers notifies observers of type changes
func (t *AbstractType) CallTypeObservers(txn *utils.Transaction, evt *utils.YEvent) {
	current := t
	for current != nil {
		if _, exists := txn.ChangedParentTypes[current]; !exists {
			txn.ChangedParentTypes[current] = []*utils.YEvent{}
		}
		txn.ChangedParentTypes[current] = append(txn.ChangedParentTypes[current], evt)

		if current.item == nil {
			break
		}
		current = current.Parent()
	}

	t.InvokeEventHandlers(evt, txn)
}

// CallObserver calls type observers
func (t *AbstractType) CallObserver(txn *utils.Transaction, parentSubs map[string]struct{}) {
	// To be overridden by concrete types
}

// first returns the first non-deleted item
func (t *AbstractType) first() *structs.Item {
	n := t.start
	for n != nil && n.Deleted() {
		n = n.Right()
	}
	return n
}

// InvokeEventHandlers invokes registered event handlers
func (t *AbstractType) InvokeEventHandlers(evt *utils.YEvent, txn *utils.Transaction) {
	args := NewYEventArgs(evt, txn)
	for _, handler := range t.eventHandlers {
		handler(args)
	}
}

// CallDeepEventHandlerListeners invokes deep event handlers
func (t *AbstractType) CallDeepEventHandlerListeners(events []*utils.YEvent, txn *utils.Transaction) {
	args := NewYDeepEventArgs(events, txn)
	for _, handler := range t.deepEventHandlers {
		handler(args)
	}
}

// FindRootTypeKey finds the root type key
func (t *AbstractType) FindRootTypeKey() string {
	if t.Doc() != nil {
		return t.Doc().FindRootTypeKey(t)
	}
	return ""
}

// TypeMapDelete deletes a value from the type map
func (t *AbstractType) TypeMapDelete(txn *utils.Transaction, key string) {
	if item, exists := t.map_[key]; exists {
		item.Delete(txn)
	}
}

// TypeMapSet sets a value in the type map
func (t *AbstractType) TypeMapSet(txn *utils.Transaction, key string, value interface{}) {
	var left *structs.Item
	if existing, exists := t.map_[key]; exists {
		left = existing
	}

	doc := txn.Doc
	ownClientID := doc.ClientID()
	var content structs.Content

	switch v := value.(type) {
	case *utils.YDoc:
		content = structs.NewContentDoc(v)
	case *AbstractType:
		content = structs.NewContentType(v)
	case []byte:
		content = structs.NewContentBinary(v)
	case nil:
		content = structs.NewContentAny([]interface{}{nil})
	default:
		content = structs.NewContentAny([]interface{}{value})
	}

	newItem := structs.NewItem(
		utils.NewID(ownClientID, doc.Store.GetState(ownClientID)),
		left,
		left.LastID(),
		nil,
		nil,
		t,
		key,
		content,
	)
	newItem.Integrate(txn, 0)
}

// TryTypeMapGet attempts to get a value from the type map
func (t *AbstractType) TryTypeMapGet(key string) (interface{}, bool) {
	if item, exists := t.map_[key]; exists && !item.Deleted() {
		content := item.Content().GetContent()
		if len(content) > 0 {
			return content[len(content)-1], true
		}
	}
	return nil, false
}

// TypeMapGetSnapshot gets a value from the type map at a specific snapshot
func (t *AbstractType) TypeMapGetSnapshot(key string, snapshot *utils.Snapshot) interface{} {
	item, exists := t.map_[key]
	if !exists {
		item = nil
	}

	for item != nil {
		clientState, hasClient := snapshot.StateVector[item.ID().Client]
		if !hasClient || item.ID().Clock >= clientState {
			item = item.Left()
			continue
		}

		if item.IsVisible(snapshot) {
			content := item.Content().GetContent()
			if len(content) > 0 {
				return content[len(content)-1]
			}
		}
		break
	}

	return nil
}

// TypeMapEnumerate enumerates all items in the type map
func (t *AbstractType) TypeMapEnumerate() map[string]*structs.Item {
	result := make(map[string]*structs.Item)
	for k, v := range t.map_ {
		if !v.Deleted() {
			result[k] = v
		}
	}
	return result
}

// TypeMapEnumerateValues enumerates all values in the type map
func (t *AbstractType) TypeMapEnumerateValues() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range t.TypeMapEnumerate() {
		content := v.Content().GetContent()
		if len(content) > 0 {
			result[k] = content[len(content)-1]
		}
	}
	return result
}
