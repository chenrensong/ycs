// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"errors"

	"github.com/chenrensong/ygo/contracts"
)

const StructGCRefNumber byte = 0

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

// GetID returns the ID
func (gc *StructGC) GetID() contracts.StructID {
	return gc.id
}

// GetLength returns the length
func (gc *StructGC) GetLength() int {
	return gc.length
}

// SetLength sets the length
func (gc *StructGC) SetLength(length int) {
	gc.length = length
}

// GetDeleted always returns true for GC structs
func (gc *StructGC) GetDeleted() bool {
	return true
}

// GetContent throws not implemented error
func (gc *StructGC) GetContent() contracts.IContentEx {
	panic("not implemented for StructGC")
}

// SetContent throws not implemented error
func (gc *StructGC) SetContent(content contracts.IContent) {
	panic("not implemented for StructGC")
}

// GetCountable throws not implemented error
func (gc *StructGC) GetCountable() bool {
	panic("not implemented for StructGC")
}

// GetKeep throws not implemented error
func (gc *StructGC) GetKeep() bool {
	panic("not implemented for StructGC")
}

// SetKeep throws not implemented error
func (gc *StructGC) SetKeep(value bool) {
	panic("not implemented for StructGC")
}

// GetLastID throws not implemented error
func (gc *StructGC) GetLastID() contracts.StructID {
	panic("not implemented for StructGC")
}

// GetLeft throws not implemented error
func (gc *StructGC) GetLeft() contracts.IStructItem {
	panic("not implemented for StructGC")
}

// SetLeft throws not implemented error
func (gc *StructGC) SetLeft(left contracts.IStructItem) {
	panic("not implemented for StructGC")
}

// GetLeftOrigin throws not implemented error
func (gc *StructGC) GetLeftOrigin() *contracts.StructID {
	panic("not implemented for StructGC")
}

// SetLeftOrigin throws not implemented error
func (gc *StructGC) SetLeftOrigin(leftOrigin *contracts.StructID) {
	panic("not implemented for StructGC")
}

// GetMarker throws not implemented error
func (gc *StructGC) GetMarker() bool {
	panic("not implemented for StructGC")
}

// SetMarker throws not implemented error
func (gc *StructGC) SetMarker(value bool) {
	panic("not implemented for StructGC")
}

// GetParent throws not implemented error
func (gc *StructGC) GetParent() interface{} {
	panic("not implemented for StructGC")
}

// SetParent throws not implemented error
func (gc *StructGC) SetParent(parent interface{}) {
	panic("not implemented for StructGC")
}

// GetParentSub throws not implemented error
func (gc *StructGC) GetParentSub() string {
	panic("not implemented for StructGC")
}

// SetParentSub throws not implemented error
func (gc *StructGC) SetParentSub(parentSub string) {
	panic("not implemented for StructGC")
}

// GetRedone throws not implemented error
func (gc *StructGC) GetRedone() *contracts.StructID {
	panic("not implemented for StructGC")
}

// SetRedone throws not implemented error
func (gc *StructGC) SetRedone(redone *contracts.StructID) {
	panic("not implemented for StructGC")
}

// GetRight throws not implemented error
func (gc *StructGC) GetRight() contracts.IStructItem {
	panic("not implemented for StructGC")
}

// SetRight throws not implemented error
func (gc *StructGC) SetRight(right contracts.IStructItem) {
	panic("not implemented for StructGC")
}

// GetRightOrigin throws not implemented error
func (gc *StructGC) GetRightOrigin() *contracts.StructID {
	panic("not implemented for StructGC")
}

// SetRightOrigin throws not implemented error
func (gc *StructGC) SetRightOrigin(rightOrigin *contracts.StructID) {
	panic("not implemented for StructGC")
}

// MergeWith merges with another struct
func (gc *StructGC) MergeWith(right contracts.IStructItem) bool {
	if rightGC, ok := right.(*StructGC); ok {
		gc.length += rightGC.length
		return true
	}
	return false
}

// Delete does nothing for GC structs
func (gc *StructGC) Delete(transaction contracts.ITransaction) {
	// Do nothing
}

// Integrate integrates the struct into the store
func (gc *StructGC) Integrate(transaction contracts.ITransaction, offset int) {
	if offset > 0 {
		gc.id = contracts.StructID{
			Client: gc.id.Client,
			Clock:  gc.id.Clock + int64(offset),
		}
		gc.length -= offset
	}

	transaction.GetDoc().GetStore().AddStruct(gc)
}

// GetMissing returns missing information
func (gc *StructGC) GetMissing(transaction contracts.ITransaction, store contracts.IStructStore) *int64 {
	return nil
}

// Write writes the struct to encoder
func (gc *StructGC) Write(encoder contracts.IUpdateEncoder, offset int) {
	encoder.WriteInfo(StructGCRefNumber)
	encoder.WriteLength(gc.length - offset)
}

// Gc garbage collects the struct
func (gc *StructGC) Gc(store contracts.IStructStore, parentGCd bool) {
	panic("not implemented for StructGC")
}

// IsVisible checks if the struct is visible in snapshot
func (gc *StructGC) IsVisible(snapshot contracts.ISnapshot) bool {
	panic("not implemented for StructGC")
}

// KeepItemAndParents keeps item and parents
func (gc *StructGC) KeepItemAndParents(value bool) {
	panic("not implemented for StructGC")
}

// MarkDeleted marks the struct as deleted
func (gc *StructGC) MarkDeleted() {
	panic("not implemented for StructGC")
}

// SplitItem splits the struct item
func (gc *StructGC) SplitItem(transaction contracts.ITransaction, diff int) (contracts.IStructItem, error) {
	return nil, errors.New("not implemented for StructGC")
}
