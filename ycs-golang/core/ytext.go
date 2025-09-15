package core

import (
	"strings"
	"ycs/contracts"
)

const YTextRefID = 2

// ChangeType represents the type of change in YText
type ChangeType int

const (
	ChangeTypeInsert ChangeType = iota
	ChangeTypeDelete
	ChangeTypeRetain
)

// YTextEvent represents an event for YText changes
type YTextEvent struct {
	*YEvent
	Subs             map[string]struct{}
	KeysChanged      map[string]struct{}
	ChildListChanged bool
	delta            []contracts.Delta
}

// NewYTextEvent creates a new YTextEvent
func NewYTextEvent(text contracts.IYText, transaction contracts.ITransaction, subs map[string]struct{}) *YTextEvent {
	event := &YTextEvent{
		YEvent:           NewYEvent(text.(contracts.IAbstractType), transaction),
		Subs:             subs,
		KeysChanged:      make(map[string]struct{}),
		ChildListChanged: false,
	}

	if subs != nil {
		for sub := range subs {
			if sub == "" {
				event.ChildListChanged = true
			} else {
				event.KeysChanged[sub] = struct{}{}
			}
		}
	}

	return event
}

// GetDelta returns the delta representation of changes
func (yte *YTextEvent) GetDelta() []contracts.Delta {
	if yte.delta == nil {
		yte.computeDelta()
	}
	return yte.delta
}

// computeDelta computes the delta representation (simplified version)
func (yte *YTextEvent) computeDelta() {
	// This is a simplified implementation
	// The full implementation would be much more complex
	yte.delta = []contracts.Delta{}
}

// YText represents a shared text implementation
type YText struct {
	*AbstractType
	prelimContent []interface{}
}

// NewYText creates a new YText
func NewYText(prelimContent []interface{}) *YText {
	yt := &YText{
		AbstractType:  NewAbstractType(),
		prelimContent: make([]interface{}, 0),
	}

	if prelimContent != nil {
		yt.prelimContent = append(yt.prelimContent, prelimContent...)
	}

	return yt
}

// GetLength returns the length of the text
func (yt *YText) GetLength() int {
	if yt.prelimContent != nil {
		return len(yt.prelimContent)
	}
	return yt.AbstractType.GetLength()
}

// Clone creates a clone of the YText
func (yt *YText) Clone() contracts.IYText {
	return yt.InternalClone().(contracts.IYText)
}

// Integrate integrates the YText into a document
func (yt *YText) Integrate(doc contracts.IYDoc, item contracts.IStructItem) {
	yt.AbstractType.Integrate(doc, item)
	if len(yt.prelimContent) > 0 {
		yt.InsertContent(0, yt.prelimContent)
		yt.prelimContent = nil
	}
}

// InternalCopy creates an internal copy
func (yt *YText) InternalCopy() contracts.IAbstractType {
	return NewYText(nil)
}

// InternalClone creates an internal clone
func (yt *YText) InternalClone() contracts.IAbstractType {
	clone := NewYText(nil)
	// Clone the content - simplified implementation
	return clone
}

// Write writes the YText to an encoder
func (yt *YText) Write(encoder contracts.IUpdateEncoder) {
	encoder.WriteTypeRef(YTextRefID)
}

// ReadYText reads a YText from a decoder
func ReadYText(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
	return NewYText(nil)
}

// CallObserver creates YTextEvent and calls observers
func (yt *YText) CallObserver(transaction contracts.ITransaction, parentSubs map[string]struct{}) {
	yt.AbstractType.CallObserver(transaction, parentSubs)
	// Create interface reference for YText
	var ytInterface contracts.IYText = yt
	yt.CallTypeObservers(transaction, NewYTextEvent(ytInterface, transaction, parentSubs))
}

// InsertContent inserts content array at the specified index (internal method)
func (yt *YText) InsertContent(index int, content []interface{}) {
	yt.insertContent(index, content)
}

// insertAt performs the actual insertion during a transaction
func (yt *YText) insertAt(transaction contracts.ITransaction, index int, content []interface{}) {
	// This is a simplified implementation
	// The full implementation would handle complex text operations
}

// Delete deletes content from the specified range
func (yt *YText) Delete(index int, length int) {
	if yt.GetDoc() != nil {
		yt.GetDoc().Transact(func(tr contracts.ITransaction) {
			yt.deleteAt(tr, index, length)
		}, "delete")
	} else {
		// Handle preliminary content deletion
		if index < len(yt.prelimContent) {
			end := index + length
			if end > len(yt.prelimContent) {
				end = len(yt.prelimContent)
			}
			newContent := make([]interface{}, 0, len(yt.prelimContent)-(end-index))
			newContent = append(newContent, yt.prelimContent[:index]...)
			newContent = append(newContent, yt.prelimContent[end:]...)
			yt.prelimContent = newContent
		}
	}
}

// deleteAt performs the actual deletion during a transaction
func (yt *YText) deleteAt(transaction contracts.ITransaction, index int, length int) {
	// This is a simplified implementation
	// The full implementation would handle complex text operations
}

// ToString returns the string representation of the text
func (yt *YText) ToString() string {
	if yt.prelimContent != nil {
		var builder strings.Builder
		for _, item := range yt.prelimContent {
			if str, ok := item.(string); ok {
				builder.WriteString(str)
			}
		}
		return builder.String()
	}

	// For integrated YText, iterate through the items
	var builder strings.Builder
	item := yt.GetStart()
	for item != nil {
		if !item.GetDeleted() && item.GetCountable() {
			content := item.GetContent()
			contentData := content.GetContent()
			if len(contentData) > 0 {
				if str, ok := contentData[0].(string); ok {
					builder.WriteString(str)
				}
			}
		}
		item = item.GetNext()
	}
	return builder.String()
}

// Format applies formatting to a range of text
func (yt *YText) Format(index int, length int, attributes map[string]interface{}) {
	if yt.GetDoc() != nil {
		yt.GetDoc().Transact(func(tr contracts.ITransaction) {
			yt.formatAt(tr, index, length, attributes)
		}, "format")
	}
}

// formatAt performs the actual formatting during a transaction
func (yt *YText) formatAt(transaction contracts.ITransaction, index int, length int, attributes map[string]interface{}) {
	// This is a simplified implementation
	// The full implementation would handle complex formatting operations
}

// GetAttributesAt returns the attributes at the specified index
func (yt *YText) GetAttributesAt(index int) map[string]interface{} {
	// This is a simplified implementation
	return make(map[string]interface{})
}

// TryGc attempts to garbage collect the YText
func (yt *YText) TryGc(store contracts.IStructStore) {
	// This is a simplified implementation
	// The full implementation would handle garbage collection
}

// ApplyDelta applies delta operations to the text
func (yt *YText) ApplyDelta(delta []contracts.Delta, sanitize ...bool) {
	// TODO: Implement delta application
}

// GetAttribute returns a single attribute value
func (yt *YText) GetAttribute(name string) interface{} {
	// TODO: Implement attribute retrieval
	return nil
}

// GetAttributes returns all attributes (implements IYText interface)
func (yt *YText) GetAttributes() map[string]interface{} {
	// TODO: Implement attributes retrieval
	return make(map[string]interface{})
}

// Insert with string parameter (implements IYText interface)
func (yt *YText) Insert(index int, text string, attributes ...map[string]interface{}) {
	content := []interface{}{text}
	yt.insertContent(index, content)
}

// insertContent is the internal method for inserting content
func (yt *YText) insertContent(index int, content []interface{}) {
	if yt.GetDoc() != nil {
		yt.GetDoc().Transact(func(tr contracts.ITransaction) {
			yt.insertAt(tr, index, content)
		}, "insert")
	} else {
		// Store as preliminary content
		if index == len(yt.prelimContent) {
			yt.prelimContent = append(yt.prelimContent, content...)
		} else {
			// Insert at specific position
			newContent := make([]interface{}, 0, len(yt.prelimContent)+len(content))
			newContent = append(newContent, yt.prelimContent[:index]...)
			newContent = append(newContent, content...)
			newContent = append(newContent, yt.prelimContent[index:]...)
			yt.prelimContent = newContent
		}
	}
}

// InsertEmbed inserts an embed object at the specified index
func (yt *YText) InsertEmbed(index int, embed interface{}, attributes ...map[string]interface{}) {
	content := []interface{}{embed}
	yt.insertContent(index, content)
}

// RemoveAttribute removes an attribute
func (yt *YText) RemoveAttribute(name string) {
	// TODO: Implement attribute removal
}

// SetAttribute sets an attribute
func (yt *YText) SetAttribute(name string, value interface{}) {
	// TODO: Implement attribute setting
}

// ToDelta converts the text to delta format
func (yt *YText) ToDelta(snapshot contracts.ISnapshot, prevSnapshot contracts.ISnapshot, computeYChange func(contracts.YTextChangeType, contracts.StructID, contracts.YTextChangeAttributes) interface{}) []contracts.Delta {
	// TODO: Implement delta conversion
	return []contracts.Delta{}
}
