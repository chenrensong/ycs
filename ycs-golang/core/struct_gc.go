package core

import (
	"ycs/contracts"
)

// StructGC represents a garbage collected struct
type StructGC struct {
	id     contracts.StructID
	length int
}

// NewStructGC creates a new StructGC
func NewStructGC(id contracts.StructID, length int) *StructGC {
	return &StructGC{
		id:     id,
		length: length,
	}
}

// GetID returns the struct ID
func (gc *StructGC) GetID() contracts.StructID {
	return gc.id
}

// SetID sets the struct ID
func (gc *StructGC) SetID(id contracts.StructID) {
	gc.id = id
}

// GetLength returns the length
func (gc *StructGC) GetLength() int {
	return gc.length
}

// SetLength sets the length
func (gc *StructGC) SetLength(length int) {
	gc.length = length
}

// IsDeleted returns true since GC structs are considered deleted
func (gc *StructGC) IsDeleted() bool {
	return true
}

// GetDeleted returns true since GC structs are considered deleted
func (gc *StructGC) GetDeleted() bool {
	return true
}

// IsGC returns true since this is a GC struct
func (gc *StructGC) IsGC() bool {
	return true
}

// GetCountable returns false since GC structs are not countable
func (gc *StructGC) GetCountable() bool {
	return false
}

// SetCountable does nothing for GC structs
func (gc *StructGC) SetCountable(countable bool) {
	// Do nothing
}

// GetLeft returns nil since GC structs don't have left references
func (gc *StructGC) GetLeft() contracts.IStructItem {
	return nil
}

// SetLeft does nothing for GC structs
func (gc *StructGC) SetLeft(left contracts.IStructItem) {
	// Do nothing
}

// GetRight returns nil since GC structs don't have right references
func (gc *StructGC) GetRight() contracts.IStructItem {
	return nil
}

// SetRight does nothing for GC structs
func (gc *StructGC) SetRight(right contracts.IStructItem) {
	// Do nothing
}

// GetLeftOrigin returns nil
func (gc *StructGC) GetLeftOrigin() *contracts.StructID {
	return nil
}

// SetLeftOrigin does nothing for GC structs
func (gc *StructGC) SetLeftOrigin(leftOrigin *contracts.StructID) {
	// Do nothing
}

// GetRightOrigin returns nil
func (gc *StructGC) GetRightOrigin() *contracts.StructID {
	return nil
}

// SetRightOrigin does nothing for GC structs
func (gc *StructGC) SetRightOrigin(rightOrigin *contracts.StructID) {
	// Do nothing
}

// GetParent returns nil since GC structs don't have parents
func (gc *StructGC) GetParent() interface{} {
	return nil
}

// SetParent does nothing for GC structs
func (gc *StructGC) SetParent(parent interface{}) {
	// Do nothing
}

// GetParentSub returns empty string
func (gc *StructGC) GetParentSub() string {
	return ""
}

// SetParentSub does nothing for GC structs
func (gc *StructGC) SetParentSub(parentSub string) {
	// Do nothing
}

// GetRedone returns nil
func (gc *StructGC) GetRedone() *contracts.StructID {
	return nil
}

// SetRedone does nothing for GC structs
func (gc *StructGC) SetRedone(redone *contracts.StructID) {
	// Do nothing
}

// GetContent returns nil since GC structs don't have content
func (gc *StructGC) GetContent() contracts.IContentEx {
	return nil
}

// SetContent does nothing for GC structs
func (gc *StructGC) SetContent(content contracts.IContentEx) {
	// Do nothing
}

// GetLastID returns the last ID based on this struct's ID and length
func (gc *StructGC) GetLastID() contracts.StructID {
	if gc.length <= 1 {
		return gc.id
	}
	lastID := contracts.StructID{
		Client: gc.id.Client,
		Clock:  gc.id.Clock + int64(gc.length) - 1,
	}
	return lastID
}

// IsVisible returns false since GC structs are not visible
func (gc *StructGC) IsVisible(snapshot contracts.ISnapshot) bool {
	return false
}

// Delete does nothing for GC structs since they're already considered deleted
func (gc *StructGC) Delete(transaction contracts.ITransaction) {
	// Do nothing - already deleted
}

// KeepItemAndParents does nothing for GC structs
func (gc *StructGC) KeepItemAndParents(keep bool) {
	// Do nothing
}

// MarkDeleted does nothing for GC structs
func (gc *StructGC) MarkDeleted() {
	// Do nothing
}

// SplitItem returns itself since GC structs can't be split
func (gc *StructGC) SplitItem(transaction contracts.ITransaction, diff int) contracts.IStructItem {
	if diff == 0 {
		return gc
	}

	// Create a new GC struct for the split part
	rightID := contracts.StructID{
		Client: gc.id.Client,
		Clock:  gc.id.Clock + int64(diff),
	}

	rightGC := NewStructGC(rightID, gc.length-diff)
	gc.length = diff

	return rightGC
}

// Integrate does nothing for GC structs
func (gc *StructGC) Integrate(transaction contracts.ITransaction, offset int) {
	// Do nothing
}

// Write writes the GC struct to an encoder
func (gc *StructGC) Write(encoder contracts.IUpdateEncoder, offset int) error {
	encoder.WriteLength(gc.length - offset)
	return nil
}

// GetMissing returns nil since GC structs don't have missing dependencies
func (gc *StructGC) GetMissing(transaction contracts.ITransaction, store contracts.IStructStore) *int64 {
	return nil
}

// TryToMergeWithRight tries to merge with right struct (only works with other GC structs)
func (gc *StructGC) TryToMergeWithRight(right contracts.IStructItem) bool {
	if rightGC, ok := right.(*StructGC); ok {
		if gc.id.Client == rightGC.id.Client && gc.id.Clock+int64(gc.length) == rightGC.id.Clock {
			gc.length += rightGC.length
			return true
		}
	}
	return false
}

// MergeWith tries to merge with right struct
func (gc *StructGC) MergeWith(right contracts.IStructItem) bool {
	return gc.TryToMergeWithRight(right)
}

// Gc does nothing for GC structs
func (gc *StructGC) Gc(store contracts.IStructStore, parentGCd bool) {
	// Do nothing - already GC'd
}

// GetKeep returns false
func (gc *StructGC) GetKeep() bool {
	return false
}

// SetKeep does nothing
func (gc *StructGC) SetKeep(keep bool) {
	// Do nothing
}

// GetMarker returns false
func (gc *StructGC) GetMarker() bool {
	return false
}

// SetMarker does nothing
func (gc *StructGC) SetMarker(marker bool) {
	// Do nothing
}

// GetNext returns nil
func (gc *StructGC) GetNext() contracts.IStructItem {
	return nil
}

// GetPrev returns nil
func (gc *StructGC) GetPrev() contracts.IStructItem {
	return nil
}
