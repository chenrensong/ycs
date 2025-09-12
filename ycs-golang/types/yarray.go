package types

import (
	"github.com/yjs/ycs-golang/structs"
	"github.com/yjs/ycs-golang/utils"
)

// YArrayEvent represents an event for YArray
type YArrayEvent struct {
	*YEvent
}

// NewYArrayEvent creates a new YArrayEvent
func NewYArrayEvent(arr *YArray, transaction *utils.Transaction) *YArrayEvent {
	return &YArrayEvent{
		YEvent: NewYEvent(arr.AbstractType, transaction),
	}
}

// YArray represents an array type
type YArray struct {
	*YArrayBase
	PrelimContent []interface{}
}

// YArrayRefId is the reference ID for YArray
const YArrayRefId byte = 0

// NewYArray creates a new YArray
func NewYArray(prelimContent []interface{}) *YArray {
	content := make([]interface{}, 0)
	if prelimContent != nil {
		content = append(content, prelimContent...)
	}
	
	return &YArray{
		YArrayBase:    NewYArrayBase(),
		PrelimContent: content,
	}
}

// NewYArrayEmpty creates a new empty YArray
func NewYArrayEmpty() *YArray {
	return NewYArray(nil)
}

// Length returns the length of the array
func (y *YArray) Length() int {
	if y.PrelimContent != nil {
		return len(y.PrelimContent)
	}
	return y.YArrayBase.Length
}

// Clone creates a clone of the array
func (y *YArray) Clone() *YArray {
	// In a real implementation, you would need to cast the result
	// For now, we'll just return nil as a placeholder
	return nil
}

// Integrate integrates the array
func (y *YArray) Integrate(doc *utils.YDoc, item *structs.Item) {
	y.YArrayBase.Integrate(doc, item)
	y.Insert(0, y.PrelimContent)
	y.PrelimContent = nil
}

// InternalCopy creates an internal copy of the array
func (y *YArray) InternalCopy() *AbstractType {
	return NewYArrayEmpty().AbstractType
}

// InternalClone creates an internal clone of the array
func (y *YArray) InternalClone() *AbstractType {
	arr := NewYArrayEmpty()
	
	for _, item := range y.EnumerateList() {
		if at, ok := item.(*AbstractType); ok {
			// In a real implementation, you would need to add the cloned type
			// arr.Add([]interface{}{at.InternalClone()})
		} else {
			arr.Add([]interface{}{item})
		}
	}
	
	return arr.AbstractType
}

// Write writes the array to an encoder
func (y *YArray) Write(encoder utils.IUpdateEncoder) {
	encoder.WriteTypeRef(YArrayRefId)
}

// Read reads a YArray from a decoder
func ReadYArray(decoder utils.IUpdateDecoder) *YArray {
	return NewYArrayEmpty()
}

// CallObserver creates YArrayEvent and calls observers
func (y *YArray) CallObserver(transaction *utils.Transaction, parentSubs map[string]struct{}) {
	y.YArrayBase.CallObserver(transaction, parentSubs)
	y.CallTypeObservers(transaction, NewYArrayEvent(y, transaction).YEvent)
}

// Insert inserts new content at an index
func (y *YArray) Insert(index int, content []interface{}) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			y.InsertGenerics(tr, index, content)
		})
	} else {
		// Insert into preliminary content
		newContent := make([]interface{}, len(y.PrelimContent)+len(content))
		copy(newContent, y.PrelimContent[:index])
		copy(newContent[index:], content)
		copy(newContent[index+len(content):], y.PrelimContent[index:])
		y.PrelimContent = newContent
	}
}

// Add adds content to the end of the array
func (y *YArray) Add(content []interface{}) {
	y.Insert(y.Length(), content)
}

// Unshift adds content to the beginning of the array
func (y *YArray) Unshift(content []interface{}) {
	y.Insert(0, content)
}

// Delete deletes content at the specified index with the specified length
func (y *YArray) Delete(index, length int) {
	if length <= 0 {
		length = 1
	}
	
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			y.YArrayBase.Delete(tr, index, length)
		})
	} else {
		// Remove from preliminary content
		if index+length <= len(y.PrelimContent) {
			newContent := make([]interface{}, len(y.PrelimContent)-length)
			copy(newContent, y.PrelimContent[:index])
			copy(newContent[index:], y.PrelimContent[index+length:])
			y.PrelimContent = newContent
		}
	}
}

// Slice returns a slice of the array
func (y *YArray) Slice(start, end int) []interface{} {
	if end == 0 {
		end = y.Length()
	}
	return y.InternalSlice(start, end)
}

// Get gets an item at the specified index
func (y *YArray) Get(index int) interface{} {
	marker := y.FindMarker(index)
	n := y.Start
	
	if marker != nil {
		n = marker.P
		index -= marker.Index
	}
	
	for ; n != nil; n = n.Right {
		if !n.Deleted() && n.Countable() {
			if index < n.Length {
				content := n.Content.GetContent()
				return content[index]
			}
			index -= n.Length
		}
		
		// Type assert to Item to access Right
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			break
		}
	}
	
	return nil
}

// ToArray converts the array to a slice
func (y *YArray) ToArray() []interface{} {
	cs := make([]interface{}, 0)
	for _, item := range y.EnumerateList() {
		cs = append(cs, item)
	}
	return cs
}

// EnumerateList enumerates the list items
func (y *YArray) EnumerateList() []interface{} {
	result := make([]interface{}, 0)
	n := y.Start
	
	for n != nil {
		if n.Countable() && !n.Deleted() {
			c := n.Content.GetContent()
			for _, item := range c {
				result = append(result, item)
			}
		}
		
		// Type assert to Item to access Right
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			break
		}
	}
	
	return result
}