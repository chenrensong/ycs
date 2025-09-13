package types

const YMapRefId = 1

type YMap struct {
	AbstractType
}

func NewYMap() *YMap {
	return &YMap{
		AbstractType: AbstractType{
			_item: nil,
			_map:  make(map[string]*structs.Item),
		},
	}
}

func (y *YMap) Write(encoder UpdateEncoder.IUpdateEncoder) {
	encoder.WriteTypeRef(YMapRefId)
	// Implementation for writing the map to an update encoder
}

func (y *YMap) Read(decoder UpdateEncoder.IUpdateDecoder) *YMap {
	// Implementation for reading the map from an update decoder
	return nil
}

func (y *YMap) Get(key string) interface{} {
	item, ok := y._map[key]
	if !ok || item.Deleted {
		return nil
	}
	content := item.Content.GetContent()
	if len(content) > 0 {
		return content[0]
	}
	return nil
}

func (y *YMap) Set(key string, value interface{}) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *Transaction) {
			y.TypeMapSet(tr, key, value)
		})
	} else {
		if y._prelimContent == nil {
			y._prelimContent = make(map[string]interface{})
		}
		y._prelimContent[key] = value
	}
}

func (y *YMap) Delete(key string) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *Transaction) {
			y.TypeMapDelete(tr, key)
		})
	} else if y._prelimContent != nil {
		delete(y._prelimContent, key)
	}
}

func (y *YMap) Has(key string) bool {
	item, ok := y._map[key]
	return ok && !item.Deleted
}

func (y *YMap) Keys() []string {
	keys := make([]string, 0, len(y._map))
	for key, item := range y._map {
		if !item.Deleted && item.Countable {
			keys = append(keys, key)
		}
	}
	return keys
}

func (y *YMap) Values() []interface{} {
	values := make([]interface{}, 0, len(y._map))
	for _, item := range y._map {
		if !item.Deleted && item.Countable {
			content := item.Content.GetContent()
			if len(content) > 0 {
				values = append(values, content[0])
			}
		}
	}
	return values
}

func (y *YMap) Entries() map[string]interface{} {
	entries := make(map[string]interface{})
	for key, item := range y._map {
		if !item.Deleted && item.Countable {
			content := item.Content.GetContent()
			if len(content) > 0 {
				entries[key] = content[0]
			}
		}
	}
	return entries
}

func (y *YMap) ForEach(f func(value interface{}, key string)) {
	for key, item := range y._map {
		if !item.Deleted && item.Countable {
			content := item.Content.GetContent()
			if len(content) > 0 {
				f(content[0], key)
			}
		}
	}
}

func (y *YMap) Size() int {
	return y.Count()
}

func (y *YMap) Clone() *YMap {
	return y.InternalClone().(*YMap)
}

func (y *YMap) InternalCopy() AbstractType {
	return NewYMap()
}

func (y *YMap) InternalClone() AbstractType {
	map_ := NewYMap()

	for key, item := range y._map {
		if !item.Deleted && item.Countable {
			content := item.Content.GetContent()
			if len(content) > 0 {
				map_.Set(key, content[0])
			}
		}
	}

	return map_
}

type YMapEvent struct {
	YEvent
	KeysChanged map[string]struct{}
}

func NewYMapEvent(map_ *YMap, transaction *Transaction, subs map[string]struct{}) *YMapEvent {
	return &YMapEvent{
		YEvent:      *NewYEvent(&map_.AbstractType, transaction),
		KeysChanged: subs,
	}
}

func (y *YMap) Count() int {
	if y._prelimContent != nil {
		return len(y._prelimContent)
	}
	// Count actual items in the map
	count := 0
	for _, item := range y._map {
		if !item.Deleted && item.Countable {
			count++
		}
	}
	return count
}

func (y *YMap) Integrate(doc *Doc, item *structs.Item) {
	// Call base class implementation
	y.AbstractType.Integrate(doc, item)

	// Set preliminary content if any exists
	for key, value := range y._prelimContent {
		y.Set(key, value)
	}

	y._prelimContent = nil
}

func (y *YMap) CallObserver(transaction *Transaction, parentSubs map[string]struct{}) {
	// Call base class implementation
	y.AbstractType.CallObserver(transaction, parentSubs)

	// Create and call YMapEvent
	observer := NewYMapEvent(y, transaction, parentSubs)
	for _, observerFunc := range y._observers {
		observerFunc(observer)
	}
}

func (y *YMap) TypeMapSet(transaction *Transaction, key string, value interface{}) {
	// Implementation for setting a value in the map within a transaction
	// This is a simplified implementation
	item := &structs.Item{
		Id:          ID{Client: 0, Clock: 0}, // Should be set by the document
		Left:        nil,                     // Should be set based on the document structure
		LeftOrigin:  nil,                     // Should be set based on the document structure
		Right:       nil,                     // Should be set based on the document structure
		RightOrigin: nil,                     // Should be set based on the document structure
		Parent:      y,
		ParentSub:   key,
		Content:     &structs.ContentAny{Content: value},
	}

	// Add the item to the document
	if doc := transaction.Doc; doc != nil {
		doc.Store.Integrate(transaction, item)
	}

	// Add the item to our map
	y._map[key] = item
}

func (y *YMap) TypeMapDelete(transaction *Transaction, key string) {
	// Implementation for deleting a value from the map within a transaction
	// This is a simplified implementation
	if item, ok := y._map[key]; ok {
		item.Delete(transaction)
	}
}
