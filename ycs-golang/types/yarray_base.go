package types

import (
	"sync"
	"sync/atomic"
	"github.com/yjs/ycs-golang/structs"
	"github.com/yjs/ycs-golang/utils"
)

// ArraySearchMarker represents a search marker for array operations
type ArraySearchMarker struct {
	P         *structs.Item
	Index     int
	Timestamp int64
}

// GlobalSearchMarkerTimestamp is assigned to '-1', so the first timestamp is '0'
var globalSearchMarkerTimestamp int64 = -1

// NewArraySearchMarker creates a new ArraySearchMarker
func NewArraySearchMarker(p *structs.Item, index int) *ArraySearchMarker {
	marker := &ArraySearchMarker{
		P:     p,
		Index: index,
	}
	
	marker.RefreshTimestamp()
	return marker
}

// RefreshTimestamp refreshes the timestamp of the marker
func (m *ArraySearchMarker) RefreshTimestamp() {
	m.Timestamp = atomic.AddInt64(&globalSearchMarkerTimestamp, 1)
}

// Update updates the marker with new position
func (m *ArraySearchMarker) Update(p *structs.Item, index int) {
	m.P.Marker = false
	
	m.P = p
	m.P.Marker = true
	m.Index = index
	
	m.RefreshTimestamp()
}

// ArraySearchMarkerCollection represents a collection of search markers
type ArraySearchMarkerCollection struct {
	searchMarkers []*ArraySearchMarker
	mutex         sync.RWMutex
}

// NewArraySearchMarkerCollection creates a new ArraySearchMarkerCollection
func NewArraySearchMarkerCollection() *ArraySearchMarkerCollection {
	return &ArraySearchMarkerCollection{
		searchMarkers: make([]*ArraySearchMarker, 0),
	}
}

// Count returns the number of search markers
func (c *ArraySearchMarkerCollection) Count() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.searchMarkers)
}

// Clear clears all search markers
func (c *ArraySearchMarkerCollection) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.searchMarkers = c.searchMarkers[:0]
}

// MarkPosition marks a position with a search marker
func (c *ArraySearchMarkerCollection) MarkPosition(p *structs.Item, index int) *ArraySearchMarker {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	const maxSearchMarkers = 80
	
	if len(c.searchMarkers) >= maxSearchMarkers {
		// Override oldest marker (we don't want to create more objects)
		var marker *ArraySearchMarker
		minTimestamp := c.searchMarkers[0].Timestamp
		
		for _, m := range c.searchMarkers {
			if m.Timestamp < minTimestamp {
				minTimestamp = m.Timestamp
				marker = m
			}
		}
		
		marker.Update(p, index)
		return marker
	} else {
		// Create a new marker
		pm := NewArraySearchMarker(p, index)
		c.searchMarkers = append(c.searchMarkers, pm)
		return pm
	}
}

// UpdateMarkerChanges updates markers when a change happened
func (c *ArraySearchMarkerCollection) UpdateMarkerChanges(index, len int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i := len(c.searchMarkers) - 1; i >= 0; i-- {
		m := c.searchMarkers[i]
		
		if len > 0 {
			p := m.P
			p.SetMarker(false)
			
			// Ideally we just want to do a simple position comparison, but this will only work if
			// search markers don't point to deleted items for formats.
			// Iterate marker to prev undeleted countable position so we know what to do when updating a position.
			for p != nil && (p.Deleted() || !p.Countable()) {
				p = p.Prev()
				if p != nil && !p.Deleted() && p.Countable() {
					// Adjust position. The loop should break now.
					m.Index -= p.Length
				}
			}
			
			if p == nil || p.Marker() {
				// Remove search marker if updated position is null or if position is already marked.
				c.searchMarkers = append(c.searchMarkers[:i], c.searchMarkers[i+1:]...)
				continue
			}
			
			m.P = p
			p.SetMarker(true)
		}
		
		// A simple index <= m.Index check would actually suffice.
		if index < m.Index || (len > 0 && index == m.Index) {
			if index > m.Index+len {
				m.Index = index
			} else {
				m.Index = m.Index + len
			}
		}
	}
}

// YArrayBase represents the base class for array types
type YArrayBase struct {
	*AbstractType
	searchMarkers *ArraySearchMarkerCollection
}

// NewYArrayBase creates a new YArrayBase
func NewYArrayBase() *YArrayBase {
	return &YArrayBase{
		AbstractType:  NewAbstractType(),
		searchMarkers: NewArraySearchMarkerCollection(),
	}
}

// ClearSearchMarkers clears all search markers
func (y *YArrayBase) ClearSearchMarkers() {
	y.searchMarkers.Clear()
}

// CallObserver creates YArrayEvent and calls observers
func (y *YArrayBase) CallObserver(transaction *utils.Transaction, parentSubs map[string]struct{}) {
	if !transaction.Local {
		y.searchMarkers.Clear()
	}
}

// InsertGenerics inserts generic content at the specified index
func (y *YArrayBase) InsertGenerics(transaction *utils.Transaction, index int, content []interface{}) {
	if index == 0 {
		if y.searchMarkers.Count() > 0 {
			y.searchMarkers.UpdateMarkerChanges(index, len(content))
		}
		
		y.InsertGenericsAfter(transaction, nil, content)
		return
	}
	
	startIndex := index
	marker := y.FindMarker(index)
	n := y.Start
	
	if marker != nil {
		n = marker.P
		index -= marker.Index
		
		// We need to iterate one to the left so that the algorithm works.
		if index == 0 {
			// @todo: refactor this as it actually doesn't consider formats.
			n = n.Prev()
			if n != nil && n.Countable() && !n.Deleted() {
				index += n.Length
			} else {
				index += 0
			}
		}
	}
	
	for ; n != nil; n = n.Right {
		if !n.Deleted() && n.Countable() {
			if index <= n.Length {
				if index < n.Length {
					// insert in-between
					transaction.Doc.Store.GetItemCleanStart(transaction, &utils.ID{Client: n.Id.Client, Clock: n.Id.Clock + index})
				}
				break
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
	
	if y.searchMarkers.Count() > 0 {
		y.searchMarkers.UpdateMarkerChanges(startIndex, len(content))
	}
	
	y.InsertGenericsAfter(transaction, n, content)
}

// InsertGenericsAfter inserts generic content after the specified reference item
func (y *YArrayBase) InsertGenericsAfter(transaction *utils.Transaction, referenceItem *structs.Item, content []interface{}) {
	left := referenceItem
	doc := transaction.Doc
	ownClientId := doc.ClientId
	store := doc.Store
	var right *structs.Item
	
	if referenceItem == nil {
		right = y.Start
	} else {
		// Type assert to Item to access Right
		if rightItem, ok := referenceItem.Right.(*structs.Item); ok {
			right = rightItem
		}
	}
	
	jsonContent := make([]interface{}, 0)
	
	packJsonContent := func() {
		if len(jsonContent) > 0 {
			newLeft := structs.NewItem(
				&utils.ID{Client: ownClientId, Clock: store.GetState(ownClientId)},
				left,
				func() *utils.ID {
					if left != nil {
						return left.LastId()
					}
					return nil
				}(),
				right,
				func() *utils.ID {
					if right != nil {
						return right.Id
					}
					return nil
				}(),
				y,
				"",
				structs.NewContentAny(jsonContent),
			)
			newLeft.Integrate(transaction, 0)
			left = newLeft
			jsonContent = jsonContent[:0]
		}
	}
	
	for _, c := range content {
		switch v := c.(type) {
		case []byte:
			packJsonContent()
			newLeft := structs.NewItem(
				&utils.ID{Client: ownClientId, Clock: store.GetState(ownClientId)},
				left,
				func() *utils.ID {
					if left != nil {
						return left.LastId()
					}
					return nil
				}(),
				right,
				func() *utils.ID {
					if right != nil {
						return right.Id
					}
					return nil
				}(),
				y,
				"",
				structs.NewContentBinary(v),
			)
			newLeft.Integrate(transaction, 0)
			left = newLeft
		case *utils.YDoc:
			packJsonContent()
			newLeft := structs.NewItem(
				&utils.ID{Client: ownClientId, Clock: store.GetState(ownClientId)},
				left,
				func() *utils.ID {
					if left != nil {
						return left.LastId()
					}
					return nil
				}(),
				right,
				func() *utils.ID {
					if right != nil {
						return right.Id
					}
					return nil
				}(),
				y,
				"",
				structs.NewContentDoc(v),
			)
			newLeft.Integrate(transaction, 0)
			left = newLeft
		case *AbstractType:
			packJsonContent()
			newLeft := structs.NewItem(
				&utils.ID{Client: ownClientId, Clock: store.GetState(ownClientId)},
				left,
				func() *utils.ID {
					if left != nil {
						return left.LastId()
					}
					return nil
				}(),
				right,
				func() *utils.ID {
					if right != nil {
						return right.Id
					}
					return nil
				}(),
				y,
				"",
				structs.NewContentType(v),
			)
			newLeft.Integrate(transaction, 0)
			left = newLeft
		default:
			jsonContent = append(jsonContent, v)
		}
	}
	
	packJsonContent()
}

// Delete deletes content at the specified index with the specified length
func (y *YArrayBase) Delete(transaction *utils.Transaction, index, length int) {
	if length == 0 {
		return
	}
	
	startIndex := index
	startLength := length
	marker := y.FindMarker(index)
	n := y.Start
	
	if marker != nil {
		n = marker.P
		index -= marker.Index
	}
	
	// Compute the first item to be deleted.
	for ; n != nil && index > 0; n = n.Right {
		if !n.Deleted() && n.Countable() {
			if index < n.Length {
				transaction.Doc.Store.GetItemCleanStart(transaction, &utils.ID{Client: n.Id.Client, Clock: n.Id.Clock + index})
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
	
	// Delete all items until done.
	for length > 0 && n != nil {
		if !n.Deleted() {
			if length < n.Length {
				transaction.Doc.Store.GetItemCleanStart(transaction, &utils.ID{Client: n.Id.Client, Clock: n.Id.Clock + length})
			}
			
			n.Delete(transaction)
			length -= n.Length
		}
		
		// Type assert to Item to access Right
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			break
		}
	}
	
	if length > 0 {
		// In Go, we'll just return without throwing an exception
		// panic("Array length exceeded")
		return
	}
	
	if y.searchMarkers.Count() > 0 {
		y.searchMarkers.UpdateMarkerChanges(startIndex, -startLength+length)
	}
}

// InternalSlice returns a slice of the array
func (y *YArrayBase) InternalSlice(start, end int) []interface{} {
	if start < 0 {
		start += y.Length
	}
	
	if end < 0 {
		end += y.Length
	}
	
	if start < 0 {
		start = 0
	}
	
	if end < 0 {
		end = 0
	}
	
	if start > end {
		end = start
	}
	
	length := end - start
	
	cs := make([]interface{}, 0)
	n := y.Start
	
	for n != nil && length > 0 {
		if n.Countable() && !n.Deleted() {
			c := n.Content.GetContent()
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
		
		// Type assert to Item to access Right
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			break
		}
	}
	
	return cs
}

// FindMarker finds a marker for the specified index
func (y *YArrayBase) FindMarker(index int) *ArraySearchMarker {
	if y.Start == nil || index == 0 || y.searchMarkers == nil || y.searchMarkers.Count() == 0 {
		return nil
	}
	
	var marker *ArraySearchMarker
	if y.searchMarkers.Count() > 0 {
		// Find the marker with the closest index
		minDiff := abs(index - y.searchMarkers.searchMarkers[0].Index)
		marker = y.searchMarkers.searchMarkers[0]
		
		for _, m := range y.searchMarkers.searchMarkers {
			diff := abs(index - m.Index)
			if diff < minDiff {
				minDiff = diff
				marker = m
			}
		}
	}
	
	p := y.Start
	pIndex := 0
	
	if marker != nil {
		p = marker.P
		pIndex = marker.Index
		
		// We used it, we might need to use it again.
		marker.RefreshTimestamp()
	}
	
	// Iterate to right if possible.
	for p.Right != nil && pIndex < index {
		if !p.Deleted() && p.Countable() {
			if index < pIndex+p.Length {
				break
			}
			pIndex += p.Length
		}
		
		// Type assert to Item to access Right
		if rightItem, ok := p.Right.(*structs.Item); ok {
			p = rightItem
		} else {
			break
		}
	}
	
	// Iterate to left if necessary (might be that pIndex > index).
	for p.Left != nil && pIndex > index {
		// Type assert to Item to access Left
		if leftItem, ok := p.Left.(*structs.Item); ok {
			p = leftItem
		} else {
			break
		}
		
		if p == nil {
			break
		} else if !p.Deleted() && p.Countable() {
			pIndex -= p.Length
		}
	}
	
	// We want to make sure that p can't be merged with left, because that would screw up everything.
	// In that case just return what we have (it is most likely the best marker anyway).
	// Iterate to left until p can't be merged with left.
	for p.Left != nil && p.Left.Id.Client == p.Id.Client && p.Left.Id.Clock+p.Left.Length == p.Id.Clock {
		// Type assert to Item to access Left
		if leftItem, ok := p.Left.(*structs.Item); ok {
			p = leftItem
		} else {
			break
		}
		
		if p == nil {
			break
		} else if !p.Deleted() && p.Countable() {
			pIndex -= p.Length
		}
	}
	
	const maxSearchMarkers = 80
	
	if marker != nil && abs(marker.Index-pIndex) < (y.Length/maxSearchMarkers) {
		// Adjust existing marker.
		marker.Update(p, pIndex)
		return marker
	} else {
		// Create a new marker.
		return y.searchMarkers.MarkPosition(p, pIndex)
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}