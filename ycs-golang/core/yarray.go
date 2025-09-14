package core

import (
	"ycs/contracts"
)

const YArrayRefID = 0

// YArrayBase represents the base functionality for YArray
type YArrayBase struct {
	*AbstractType
	searchMarkers []interface{} // Placeholder for search markers
}

// NewYArrayBase creates a new YArrayBase
func NewYArrayBase() *YArrayBase {
	return &YArrayBase{
		AbstractType:  NewAbstractType(),
		searchMarkers: make([]interface{}, 0),
	}
}

// ClearSearchMarkers clears search markers
func (yab *YArrayBase) ClearSearchMarkers() {
	yab.searchMarkers = yab.searchMarkers[:0]
}

// FindMarker finds a marker (placeholder implementation)
func (yab *YArrayBase) FindMarker(index int) interface{} {
	// Placeholder implementation
	return nil
}

// IsCountable checks if an item is countable
func (yab *YArrayBase) IsCountable() bool {
	return true
}

// InsertGenerics inserts generic content
func (yab *YArrayBase) InsertGenerics(transaction contracts.ITransaction, index int, content []interface{}) {
	// Placeholder implementation
}

// DeleteRange deletes a range of items
func (yab *YArrayBase) DeleteRange(transaction contracts.ITransaction, index, length int) {
	// Placeholder implementation
}

// internalSlice returns internal slice
func (yab *YArrayBase) internalSlice(start, end int) []interface{} {
	result := make([]interface{}, 0)
	n := yab.GetStart()
	currentIndex := 0

	for n != nil && currentIndex < end {
		if !n.GetDeleted() && n.GetCountable() {
			itemLength := n.GetLength()
			if currentIndex+itemLength > start {
				content := n.GetContent().GetContent()
				startOffset := 0
				if currentIndex < start {
					startOffset = start - currentIndex
				}
				endOffset := itemLength
				if currentIndex+itemLength > end {
					endOffset = end - currentIndex
				}

				for i := startOffset; i < endOffset; i++ {
					if i < len(content) {
						result = append(result, content[i])
					}
				}
			}
			currentIndex += itemLength
		}
		n = n.GetRight()
	}

	return result
}

// YArrayEvent represents an event for YArray changes
type YArrayEvent struct {
	*YEvent
}

// NewYArrayEvent creates a new YArrayEvent
func NewYArrayEvent(arr contracts.IYArray, transaction contracts.ITransaction) *YArrayEvent {
	return &YArrayEvent{
		YEvent: NewYEvent(arr, transaction),
	}
}

// YArray represents a shared array implementation
type YArray struct {
	*YArrayBase
	prelimContent []interface{}
}

// NewYArray creates a new YArray
func NewYArray(prelimContent []interface{}) *YArray {
	ya := &YArray{
		YArrayBase:    NewYArrayBase(),
		prelimContent: make([]interface{}, 0),
	}

	if prelimContent != nil {
		ya.prelimContent = append(ya.prelimContent, prelimContent...)
	}

	return ya
}

// GetLength returns the length of the array
func (ya *YArray) GetLength() int {
	if ya.prelimContent != nil {
		return len(ya.prelimContent)
	}
	return ya.YArrayBase.GetLength()
}

// Clone creates a clone of the YArray
func (ya *YArray) Clone() contracts.IYArray {
	return ya.InternalClone().(contracts.IYArray)
}

// Integrate integrates the array with a document and item
func (ya *YArray) Integrate(doc contracts.IYDoc, item contracts.IStructItem) {
	ya.YArrayBase.Integrate(doc, item)
	ya.Insert(0, ya.prelimContent)
	ya.prelimContent = nil
}

// InternalCopy creates an internal copy
func (ya *YArray) InternalCopy() contracts.IAbstractType {
	return NewYArray(nil)
}

// InternalClone creates an internal clone
func (ya *YArray) InternalClone() contracts.IAbstractType {
	arr := NewYArray(nil)

	for _, item := range ya.enumerateList() {
		if at, ok := item.(contracts.IAbstractType); ok {
			arr.Add([]interface{}{at.InternalClone()})
		} else {
			arr.Add([]interface{}{item})
		}
	}

	return arr
}

// Write writes the array to an encoder
func (ya *YArray) Write(encoder contracts.IUpdateEncoder) {
	encoder.WriteTypeRef(YArrayRefID)
}

// ReadYArray reads a YArray from decoder
func ReadYArray(decoder contracts.IUpdateDecoder) contracts.IYArray {
	return NewYArray(nil)
}

// CallObserver creates YArrayEvent and calls observers
func (ya *YArray) CallObserver(transaction contracts.ITransaction, parentSubs map[string]struct{}) {
	ya.YArrayBase.CallObserver(transaction, parentSubs)
	ya.CallTypeObservers(transaction, NewYArrayEvent(ya, transaction))
}

// Insert inserts new content at an index
func (ya *YArray) Insert(index int, content []interface{}) {
	if ya.GetDoc() != nil {
		ya.GetDoc().Transact(func(tr contracts.ITransaction) {
			ya.InsertGenerics(tr, index, content)
		}, nil, true)
	} else {
		// Insert into preliminary content
		if index > len(ya.prelimContent) {
			index = len(ya.prelimContent)
		}

		// Extend slice if needed
		newContent := make([]interface{}, len(ya.prelimContent)+len(content))
		copy(newContent[:index], ya.prelimContent[:index])
		copy(newContent[index:index+len(content)], content)
		copy(newContent[index+len(content):], ya.prelimContent[index:])
		ya.prelimContent = newContent
	}
}

// Add adds content to the end of the array
func (ya *YArray) Add(content []interface{}) {
	ya.Insert(ya.GetLength(), content)
}

// Unshift adds content to the beginning of the array
func (ya *YArray) Unshift(content []interface{}) {
	ya.Insert(0, content)
}

// Delete deletes elements from the array
func (ya *YArray) Delete(index int, length ...int) {
	deleteLength := 1
	if len(length) > 0 {
		deleteLength = length[0]
	}

	if deleteLength <= 0 {
		return
	}

	if ya.GetDoc() != nil {
		ya.GetDoc().Transact(func(tr contracts.ITransaction) {
			ya.DeleteRange(tr, index, deleteLength)
		}, nil, true)
	} else {
		if index < 0 || index >= len(ya.prelimContent) {
			return
		}

		end := index + deleteLength
		if end > len(ya.prelimContent) {
			end = len(ya.prelimContent)
		}

		newContent := make([]interface{}, 0, len(ya.prelimContent)-(end-index))
		newContent = append(newContent, ya.prelimContent[:index]...)
		newContent = append(newContent, ya.prelimContent[end:]...)
		ya.prelimContent = newContent
	}
}

// Slice returns a slice of the array
func (ya *YArray) Slice(start ...int) []interface{} {
	startIndex := 0
	if len(start) > 0 {
		startIndex = start[0]
	}
	return ya.internalSlice(startIndex, ya.GetLength())
}

// Get returns the element at the specified index
func (ya *YArray) Get(index int) interface{} {
	marker := ya.FindMarker(index)
	n := ya.GetStart()

	if marker != nil {
		// This is a placeholder - actual marker implementation would be needed
		// n = marker.GetP()
		// index -= marker.GetIndex()
	}

	for n != nil {
		if !n.GetDeleted() && n.GetCountable() {
			if index < n.GetLength() {
				return n.GetContent().GetContent()[index]
			}
			index -= n.GetLength()
		}
		n = n.GetRight()
	}

	return nil
}

// ToArray converts the array to a Go slice
func (ya *YArray) ToArray() []interface{} {
	result := make([]interface{}, 0)
	for _, item := range ya.enumerateList() {
		result = append(result, item)
	}
	return result
}

// enumerateList enumerates all items in the list
func (ya *YArray) enumerateList() []interface{} {
	result := make([]interface{}, 0)
	n := ya.GetStart()

	for n != nil {
		if n.GetCountable() && !n.GetDeleted() {
			content := n.GetContent().GetContent()
			result = append(result, content...)
		}
		n = n.GetRight()
	}

	return result
}
