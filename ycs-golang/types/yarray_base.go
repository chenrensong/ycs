// ------------------------------------------------------------------------------
//  Copyright (c) Microsoft Corporation.  All rights reserved.
// ------------------------------------------------------------------------------

package types

import (
	"math"
	"sync/atomic"

	"github.com/chenrensong/ygo/contracts"
	"github.com/chenrensong/ygo/core"
)

// ArraySearchMarker helps us to find positions in the associative array faster
type ArraySearchMarker struct {
	P         contracts.IStructItem
	Index     int
	Timestamp int64
}

// Global search marker timestamp (assigned to -1, so the first timestamp is 0)
var globalSearchMarkerTimestamp int64 = -1

// NewArraySearchMarker creates a new ArraySearchMarker instance
func NewArraySearchMarker(p contracts.IStructItem, index int) *ArraySearchMarker {
	p.SetMarker(true)
	marker := &ArraySearchMarker{
		P:     p,
		Index: index,
	}
	marker.RefreshTimestamp()
	return marker
}

// RefreshTimestamp updates the timestamp of the marker
func (asm *ArraySearchMarker) RefreshTimestamp() {
	asm.Timestamp = atomic.AddInt64(&globalSearchMarkerTimestamp, 1)
}

// Update updates the marker with new position
func (asm *ArraySearchMarker) Update(p contracts.IStructItem, index int) {
	asm.P.SetMarker(false)
	asm.P = p
	p.SetMarker(true)
	asm.Index = index
	asm.RefreshTimestamp()
}

// ArraySearchMarkerCollection manages a collection of search markers
type ArraySearchMarkerCollection struct {
	searchMarkers []*ArraySearchMarker
}

// NewArraySearchMarkerCollection creates a new ArraySearchMarkerCollection
func NewArraySearchMarkerCollection() *ArraySearchMarkerCollection {
	return &ArraySearchMarkerCollection{
		searchMarkers: make([]*ArraySearchMarker, 0),
	}
}

// Count returns the number of search markers
func (asmc *ArraySearchMarkerCollection) Count() int {
	return len(asmc.searchMarkers)
}

// Clear removes all search markers
func (asmc *ArraySearchMarkerCollection) Clear() {
	asmc.searchMarkers = asmc.searchMarkers[:0]
}

// MarkPosition marks a position with a search marker
func (asmc *ArraySearchMarkerCollection) MarkPosition(p contracts.IStructItem, index int) *ArraySearchMarker {
	if len(asmc.searchMarkers) >= MaxSearchMarkers {
		// Override oldest marker (we don't want to create more objects)
		var oldest *ArraySearchMarker
		for _, marker := range asmc.searchMarkers {
			if oldest == nil || marker.Timestamp < oldest.Timestamp {
				oldest = marker
			}
		}
		oldest.Update(p, index)
		return oldest
	} else {
		// Create a new marker
		marker := NewArraySearchMarker(p, index)
		asmc.searchMarkers = append(asmc.searchMarkers, marker)
		return marker
	}
}

// UpdateMarkerChanges updates markers when a change happened
// This should be called before doing a deletion!
func (asmc *ArraySearchMarkerCollection) UpdateMarkerChanges(index int, length int) {
	for i := len(asmc.searchMarkers) - 1; i >= 0; i-- {
		m := asmc.searchMarkers[i]

		if length > 0 {
			p := m.P
			p.SetMarker(false)

			// Ideally we just want to do a simple position comparison, but this will only work if
			// search markers don't point to deleted items for formats.
			// Iterate marker to prev undeleted countable position so we know what to do when updating a position.
			for p != nil && (p.GetDeleted() || !p.GetCountable()) {
				p = p.GetLeft()
				if p != nil && !p.GetDeleted() && p.GetCountable() {
					// Adjust position. The loop should break now.
					m.Index -= p.GetLength()
				}
			}

			if p == nil || p.GetMarker() {
				// Remove search marker if updated position is null or if position is already marked
				asmc.searchMarkers = append(asmc.searchMarkers[:i], asmc.searchMarkers[i+1:]...)
				continue
			}

			m.P = p
			p.SetMarker(true)
		}

		// A simple index <= m.Index check would actually suffice
		if index < m.Index || (length > 0 && index == m.Index) {
			if index > m.Index+length {
				m.Index = index
			} else {
				m.Index = m.Index + length
			}
		}
	}
}

const MaxSearchMarkers = 80

// YArrayBase represents the base class for array-like types
type YArrayBase struct {
	*core.AbstractType
	searchMarkers *ArraySearchMarkerCollection
}

// NewYArrayBase creates a new YArrayBase instance
func NewYArrayBase() *YArrayBase {
	return &YArrayBase{
		AbstractType:  core.NewAbstractType(),
		searchMarkers: NewArraySearchMarkerCollection(),
	}
}

// ClearSearchMarkers clears all search markers
func (yab *YArrayBase) ClearSearchMarkers() {
	yab.searchMarkers.Clear()
}

// CallObserver creates YArrayEvent and calls observers
func (yab *YArrayBase) CallObserver(transaction contracts.ITransaction, parentSubs map[string]struct{}) {
	if !transaction.GetLocal() {
		yab.searchMarkers.Clear()
	}
}

// InsertGenerics inserts generic content at the specified index
func (yab *YArrayBase) InsertGenerics(transaction contracts.ITransaction, index int, content []interface{}) {
	if index == 0 {
		if yab.searchMarkers.Count() > 0 {
			yab.searchMarkers.UpdateMarkerChanges(index, len(content))
		}
		yab.InsertGenericsAfter(transaction, nil, content)
		return
	}

	startIndex := index
	marker := yab.FindMarker(index)
	n := yab.AbstractType.GetStart()

	if marker != nil {
		n = marker.P
		index -= marker.Index

		// We need to iterate one to the left so that the algorithm works
		if index == 0 {
			// @todo: refactor this as it actually doesn't consider formats
			n = n.GetLeft()
			if n != nil && n.GetCountable() && !n.GetDeleted() {
				index += n.GetLength()
			} else {
				index += 0
			}
		}
	}

	for ; n != nil; n = n.GetRight() {
		if !n.GetDeleted() && n.GetCountable() {
			if index <= n.GetLength() {
				if index < n.GetLength() {
					// insert in-between
					transaction.GetDoc().GetStore().GetItemCleanStart(transaction, contracts.NewStructID(n.GetID().Client, n.GetID().Clock+int64(index)))
				}
				break
			}
			index -= n.GetLength()
		}
	}

	if yab.searchMarkers.Count() > 0 {
		yab.searchMarkers.UpdateMarkerChanges(startIndex, len(content))
	}

	yab.InsertGenericsAfter(transaction, n, content)
}

// InsertGenericsAfter inserts generic content after the reference item
func (yab *YArrayBase) InsertGenericsAfter(transaction contracts.ITransaction, referenceItem contracts.IStructItem, content []interface{}) {
	left := referenceItem
	doc := transaction.GetDoc()
	ownClientID := doc.GetClientID()
	store := doc.GetStore()
	var right contracts.IStructItem
	if referenceItem == nil {
		right = yab.AbstractType.GetStart()
	} else {
		right = referenceItem.GetRight()
	}

	var jsonContent []interface{}

	packJsonContent := func() {
		if len(jsonContent) > 0 {
			var leftLastID *contracts.StructID
			var rightID *contracts.StructID
			if left != nil {
				leftLastID = &left.GetLastID()
			}
			if right != nil {
				rightID = &right.GetID()
			}

			structID := contracts.NewStructID(ownClientID, store.GetState(ownClientID))
			contentAny := contracts.NewContentAny(jsonContent)
			left = contracts.NewStructItem(structID, left, leftLastID, right, rightID, yab, "", contentAny)
			left.Integrate(transaction, 0)
			jsonContent = jsonContent[:0]
		}
	}

	for _, c := range content {
		switch v := c.(type) {
		case []byte:
			packJsonContent()
			var leftLastID *contracts.StructID
			var rightID *contracts.StructID
			if left != nil {
				leftLastID = &left.GetLastID()
			}
			if right != nil {
				rightID = &right.GetID()
			}
			structID := contracts.NewStructID(ownClientID, store.GetState(ownClientID))
			contentBinary := contracts.NewContentBinary(v)
			left = contracts.NewStructItem(structID, left, leftLastID, right, rightID, yab, "", contentBinary)
			left.Integrate(transaction, 0)

		case contracts.IYDoc:
			packJsonContent()
			var leftLastID *contracts.StructID
			var rightID *contracts.StructID
			if left != nil {
				leftLastID = &left.GetLastID()
			}
			if right != nil {
				rightID = &right.GetID()
			}
			structID := contracts.NewStructID(ownClientID, store.GetState(ownClientID))
			contentDoc := contracts.NewContentDoc(v)
			left = contracts.NewStructItem(structID, left, leftLastID, right, rightID, yab, "", contentDoc)
			left.Integrate(transaction, 0)

		case contracts.IAbstractType:
			packJsonContent()
			var leftLastID *contracts.StructID
			var rightID *contracts.StructID
			if left != nil {
				leftLastID = &left.GetLastID()
			}
			if right != nil {
				rightID = &right.GetID()
			}
			structID := contracts.NewStructID(ownClientID, store.GetState(ownClientID))
			contentType := contracts.NewContentType(v)
			left = contracts.NewStructItem(structID, left, leftLastID, right, rightID, yab, "", contentType)
			left.Integrate(transaction, 0)

		default:
			jsonContent = append(jsonContent, c)
		}
	}

	packJsonContent()
}

// Delete deletes content from the specified index
func (yab *YArrayBase) Delete(transaction contracts.ITransaction, index int, length int) {
	if length == 0 {
		return
	}

	startIndex := index
	startLength := length
	marker := yab.FindMarker(index)
	n := yab.AbstractType.GetStart()

	if marker != nil {
		n = marker.P
		index -= marker.Index
	}

	// Compute the first item to be deleted
	for n != nil && index > 0 {
		if !n.GetDeleted() && n.GetCountable() {
			if index < n.GetLength() {
				transaction.GetDoc().GetStore().GetItemCleanStart(transaction, contracts.NewStructID(n.GetID().Client, n.GetID().Clock+int64(index)))
			}
			index -= n.GetLength()
		}
		n = n.GetRight()
	}

	// Delete all items until done
	for length > 0 && n != nil {
		if !n.GetDeleted() {
			if length < n.GetLength() {
				transaction.GetDoc().GetStore().GetItemCleanStart(transaction, contracts.NewStructID(n.GetID().Client, n.GetID().Clock+int64(length)))
			}
			n.Delete(transaction)
			length -= n.GetLength()
		}
		n = n.GetRight()
	}

	if length > 0 {
		panic("Array length exceeded")
	}

	if yab.searchMarkers.Count() > 0 {
		yab.searchMarkers.UpdateMarkerChanges(startIndex, -startLength+length)
	}
}

// InternalSlice returns a slice of the array content
func (yab *YArrayBase) InternalSlice(start int, end int) []interface{} {
	if start < 0 {
		start += yab.AbstractType.GetLength()
	}
	if end < 0 {
		end += yab.AbstractType.GetLength()
	}
	if start < 0 {
		panic("start index out of range")
	}
	if end < 0 {
		panic("end index out of range")
	}
	if start > end {
		panic("end index must be greater than start index")
	}

	length := end - start
	var cs []interface{}
	n := yab.AbstractType.GetStart()

	for n != nil && length > 0 {
		if n.GetCountable() && !n.GetDeleted() {
			c := n.GetContent().GetContent()
			if len(c) <= start {
				start -= len(c)
			} else {
				for i := start; i < len(c) && length > 0; i++ {
					cs = append(cs, c[i])
					length--
				}
				start = 0
			}
		}
		n = n.GetRight()
	}

	return cs
}

// ForEach iterates over each element in the array
func (yab *YArrayBase) ForEach(fun func(interface{}, int, *YArrayBase)) {
	index := 0
	n := yab.AbstractType.GetStart()

	for n != nil {
		if n.GetCountable() && !n.GetDeleted() {
			c := n.GetContent().GetContent()
			for _, cItem := range c {
				fun(cItem, index, yab)
				index++
			}
		}
		n = n.GetRight()
	}
}

// ForEachSnapshot iterates over each element in the array for a specific snapshot
func (yab *YArrayBase) ForEachSnapshot(fun func(interface{}, int, *YArrayBase), snapshot contracts.ISnapshot) {
	index := 0
	n := yab.AbstractType.GetStart()

	for n != nil {
		if n.GetCountable() && n.IsVisible(snapshot) {
			c := n.GetContent().GetContent()
			for _, value := range c {
				fun(value, index, yab)
				index++
			}
		}
		n = n.GetRight()
	}
}

// EnumerateContent enumerates the content of the array
func (yab *YArrayBase) EnumerateContent() []interface{} {
	var result []interface{}
	n := yab.AbstractType.GetStart()

	for n != nil {
		for n != nil && n.GetDeleted() {
			n = n.GetRight()
		}

		// Check if we reached the end
		if n == nil {
			break
		}

		currentContent := n.GetContent().GetContent()
		for _, c := range currentContent {
			result = append(result, c)
		}

		// We used content of n, now iterate to next
		n = n.GetRight()
	}

	return result
}

// FindMarker finds the best search marker for the given index
// Search markers help us to find positions in the associative array faster.
// They speed up the process of finding a position without much bookkeeping.
// A maximum of 'MaxSearchMarker' objects are created.
// This function always returns a refreshed marker (updated timestamp).
func (yab *YArrayBase) FindMarker(index int) *ArraySearchMarker {
	if yab.AbstractType.GetStart() == nil || index == 0 || yab.searchMarkers == nil || yab.searchMarkers.Count() == 0 {
		return nil
	}

	var marker *ArraySearchMarker
	if yab.searchMarkers.Count() > 0 {
		// Find the marker with the closest index
		for _, m := range yab.searchMarkers.searchMarkers {
			if marker == nil || int(math.Abs(float64(index-m.Index))) < int(math.Abs(float64(index-marker.Index))) {
				marker = m
			}
		}
	}

	p := yab.AbstractType.GetStart()
	pIndex := 0

	if marker != nil {
		p = marker.P
		pIndex = marker.Index
		// We used it, we might need to use it again
		marker.RefreshTimestamp()
	}

	// Iterate to right if possible
	for p.GetRight() != nil && pIndex < index {
		if !p.GetDeleted() && p.GetCountable() {
			if index < pIndex+p.GetLength() {
				break
			}
			pIndex += p.GetLength()
		}
		p = p.GetRight()
	}

	// Iterate to left if necessary (might be that pIndex > index)
	for p.GetLeft() != nil && pIndex > index {
		p = p.GetLeft()
		if p == nil {
			break
		} else if !p.GetDeleted() && p.GetCountable() {
			pIndex -= p.GetLength()
		}
	}

	// We want to make sure that p can't be merged with left, because that would screw up everything.
	// In that case just return what we have (it is most likely the best marker anyway).
	// Iterate to left until p can't be merged with left.
	for p.GetLeft() != nil && p.GetLeft().GetID().Client == p.GetID().Client && p.GetLeft().GetID().Clock+int64(p.GetLeft().GetLength()) == p.GetID().Clock {
		p = p.GetLeft()
		if p == nil {
			break
		} else if !p.GetDeleted() && p.GetCountable() {
			pIndex -= p.GetLength()
		}
	}

	if marker != nil && int(math.Abs(float64(marker.Index-pIndex))) < p.GetParent().(contracts.IAbstractType).GetLength()/MaxSearchMarkers {
		// Adjust existing marker
		marker.Update(p, pIndex)
		return marker
	} else {
		// Create a new marker
		return yab.searchMarkers.MarkPosition(p, pIndex)
	}
}
