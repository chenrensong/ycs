// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"github.com/chenrensong/ygo/contracts"
)

// InfoEnum flags for StructItem
type InfoEnum int

const (
	Keep      InfoEnum = 1 << 0
	Countable InfoEnum = 1 << 1
	Deleted   InfoEnum = 1 << 2
	Marker    InfoEnum = 1 << 3
)

// StructItem represents a structure item
type StructItem struct {
	id          contracts.StructID
	leftOrigin  *contracts.StructID
	left        contracts.IStructItem
	rightOrigin *contracts.StructID
	right       contracts.IStructItem
	parent      interface{}
	parentSub   string
	redone      *contracts.StructID
	content     contracts.IContent
	info        InfoEnum
	length      int
}

// NewStructItem creates a new StructItem
func NewStructItem(
	id contracts.StructID,
	left contracts.IStructItem,
	leftOrigin *contracts.StructID,
	right contracts.IStructItem,
	rightOrigin *contracts.StructID,
	parent interface{},
	parentSub string,
	content contracts.IContent,
) *StructItem {
	item := &StructItem{
		id:          id,
		length:      content.GetLength(),
		leftOrigin:  leftOrigin,
		left:        left,
		right:       right,
		rightOrigin: rightOrigin,
		parent:      parent,
		parentSub:   parentSub,
		redone:      nil,
		content:     content,
	}

	if content.GetCountable() {
		item.info |= Countable
	}

	return item
}

// GetID returns the ID
func (si *StructItem) GetID() contracts.StructID {
	return si.id
}

// GetLeftOrigin returns the left origin
func (si *StructItem) GetLeftOrigin() *contracts.StructID {
	return si.leftOrigin
}

// SetLeftOrigin sets the left origin
func (si *StructItem) SetLeftOrigin(leftOrigin *contracts.StructID) {
	si.leftOrigin = leftOrigin
}

// GetLeft returns the left item
func (si *StructItem) GetLeft() contracts.IStructItem {
	return si.left
}

// SetLeft sets the left item
func (si *StructItem) SetLeft(left contracts.IStructItem) {
	si.left = left
}

// GetRightOrigin returns the right origin
func (si *StructItem) GetRightOrigin() *contracts.StructID {
	return si.rightOrigin
}

// SetRightOrigin sets the right origin
func (si *StructItem) SetRightOrigin(rightOrigin *contracts.StructID) {
	si.rightOrigin = rightOrigin
}

// GetRight returns the right item
func (si *StructItem) GetRight() contracts.IStructItem {
	return si.right
}

// SetRight sets the right item
func (si *StructItem) SetRight(right contracts.IStructItem) {
	si.right = right
}

// GetParent returns the parent
func (si *StructItem) GetParent() interface{} {
	return si.parent
}

// SetParent sets the parent
func (si *StructItem) SetParent(parent interface{}) {
	si.parent = parent
}

// GetParentSub returns the parent sub
func (si *StructItem) GetParentSub() string {
	return si.parentSub
}

// SetParentSub sets the parent sub
func (si *StructItem) SetParentSub(parentSub string) {
	si.parentSub = parentSub
}

// GetRedone returns the redone ID
func (si *StructItem) GetRedone() *contracts.StructID {
	return si.redone
}

// SetRedone sets the redone ID
func (si *StructItem) SetRedone(redone *contracts.StructID) {
	si.redone = redone
}

// GetContent returns the content
func (si *StructItem) GetContent() contracts.IContentEx {
	if contentEx, ok := si.content.(contracts.IContentEx); ok {
		return contentEx
	}
	return nil
}

// SetContent sets the content
func (si *StructItem) SetContent(content contracts.IContent) {
	si.content = content
}

// GetMarker returns the marker flag
func (si *StructItem) GetMarker() bool {
	return si.info&Marker != 0
}

// SetMarker sets the marker flag
func (si *StructItem) SetMarker(value bool) {
	if value {
		si.info |= Marker
	} else {
		si.info &= ^Marker
	}
}

// GetKeep returns the keep flag
func (si *StructItem) GetKeep() bool {
	return si.info&Keep != 0
}

// SetKeep sets the keep flag
func (si *StructItem) SetKeep(value bool) {
	if value {
		si.info |= Keep
	} else {
		si.info &= ^Keep
	}
}

// GetCountable returns the countable flag
func (si *StructItem) GetCountable() bool {
	return si.info&Countable != 0
}

// GetDeleted returns the deleted flag
func (si *StructItem) GetDeleted() bool {
	return si.info&Deleted != 0
}

// SetDeleted sets the deleted flag
func (si *StructItem) SetDeleted(value bool) {
	if value {
		si.info |= Deleted
	} else {
		si.info &= ^Deleted
	}
}

// GetLength returns the length
func (si *StructItem) GetLength() int {
	return si.length
}

// SetLength sets the length
func (si *StructItem) SetLength(length int) {
	si.length = length
}

// GetLastID returns the last ID
func (si *StructItem) GetLastID() contracts.StructID {
	if si.length == 1 {
		return si.id
	}
	return contracts.StructID{
		Client: si.id.Client,
		Clock:  si.id.Clock + int64(si.length) - 1,
	}
}

// Delete marks the item as deleted
func (si *StructItem) Delete(transaction contracts.ITransaction) {
	if !si.GetDeleted() {
		parent := si.GetParent()
		if abstractType, ok := parent.(contracts.IAbstractType); ok {
			// Remove from parent's map if it has a parentSub
			if si.parentSub != "" {
				delete(abstractType.GetMap(), si.parentSub)
			}
		}
		si.SetDeleted(true)
		transaction.GetDeleteSet().Add(si.id.Client, si.id.Clock, int64(si.length))
	}
}

// Gc garbage collects the item
func (si *StructItem) Gc(store contracts.IStructStore, parentGCd bool) {
	if !si.GetDeleted() && !si.GetKeep() {
		si.content = nil
	}
}

// IsVisible checks if the item is visible in a snapshot
func (si *StructItem) IsVisible(snapshot contracts.ISnapshot) bool {
	stateVector := snapshot.GetStateVector()
	clientState, exists := stateVector[si.id.Client]
	return exists && si.id.Clock < clientState && !si.GetDeleted()
}

// Integrate integrates the item into the document structure
func (si *StructItem) Integrate(transaction contracts.ITransaction, offset int) {
	doc := transaction.GetDoc()
	store := doc.GetStore()

	// Add to store
	store.AddStruct(si)

	// Handle parent integration
	parent := si.GetParent()
	if abstractType, ok := parent.(contracts.IAbstractType); ok {
		if si.parentSub != "" {
			// Map integration
			abstractType.GetMap()[si.parentSub] = si
		} else {
			// List integration
			if abstractType.GetStart() == nil {
				abstractType.SetStart(si)
			} else {
				// Find insertion point and update left/right links
				si.IntegrateIntoList(abstractType)
			}
		}

		// Integrate content if it's a type
		if contentType, ok := si.content.(interface {
			GetType() contracts.IAbstractType
		}); ok {
			contentType.GetType().Integrate(doc, si)
		}
	}
}

// IntegrateIntoList integrates the item into a list structure
func (si *StructItem) IntegrateIntoList(parent contracts.IAbstractType) {
	// This is a simplified version - full implementation would handle conflict resolution
	if parent.GetStart() == nil {
		parent.SetStart(si)
	} else {
		// Find the correct position based on left/right origins
		current := parent.GetStart()
		for current.GetRight() != nil {
			current = current.GetRight()
		}
		current.SetRight(si)
		si.SetLeft(current)
	}
}
