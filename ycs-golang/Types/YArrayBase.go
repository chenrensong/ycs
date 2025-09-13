package types

import (
	"math"
	"sync"

	"github.com/chenrensong/ygo/structs"
	"github.com/chenrensong/ygo/utils"
)

var _globalSearchMarkerTimestamp int64 = -1

type ArraySearchMarker struct {
	p         *structs.Item
	index     int
	timestamp int64
}

func (m *ArraySearchMarker) RefreshTimestamp() {
	_globalSearchMarkerTimestamp++
	m.timestamp = _globalSearchMarkerTimestamp
}

func (m *ArraySearchMarker) Update(p *structs.Item, index int) {
	if m.p != nil {
		m.p.Marker = false
	}

	m.p = p
	m.p.Marker = true
	m.index = index
	m.RefreshTimestamp()
}

const MaxSearchMarkers = 80

type ArraySearchMarkerCollection struct {
	searchMarkers []*ArraySearchMarker
	mutex         sync.RWMutex
}

func NewArraySearchMarkerCollection() *ArraySearchMarkerCollection {
	return &ArraySearchMarkerCollection{
		searchMarkers: make([]*ArraySearchMarker, 0),
	}
}

func (c *ArraySearchMarkerCollection) Count() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.searchMarkers)
}

func (c *ArraySearchMarkerCollection) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.searchMarkers = nil
}

func (c *ArraySearchMarkerCollection) MarkPosition(p *structs.Item, index int) *ArraySearchMarker {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.searchMarkers) >= MaxSearchMarkers {
		// Override oldest marker
		var oldest *ArraySearchMarker
		for _, m := range c.searchMarkers {
			if oldest == nil || m.timestamp < oldest.timestamp {
				oldest = m
			}
		}
		oldest.Update(p, index)
		return oldest
	} else {
		// Create a new marker
		p.Marker = true
		m := &ArraySearchMarker{p: p, index: index}
		m.RefreshTimestamp()
		c.searchMarkers = append(c.searchMarkers, m)
		return m
	}
}

func (c *ArraySearchMarkerCollection) UpdateMarkerChanges(index int, length int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i := len(c.searchMarkers) - 1; i >= 0; i-- {
		m := c.searchMarkers[i]

		if length > 0 {
			p := m.p
			p.Marker = false

			// Iterate marker to prev undeleted countable position
			for p != nil && (p.Deleted || !p.Countable) {
				if leftItem, ok := p.Left.(*structs.Item); ok {
					p = leftItem
					if p != nil && !p.Deleted && p.Countable {
						// Adjust position
						m.index -= p.Length
					}
				} else {
					p = nil
				}
			}

			if p == nil || p.Marker {
				// Remove search marker
				c.searchMarkers = append(c.searchMarkers[:i], c.searchMarkers[i+1:]...)
				continue
			}

			m.p = p
			p.Marker = true
		}

		if index < m.index || (length > 0 && index == m.index) {
			if index > m.index+length {
				m.index = index
			} else {
				m.index = m.index + length
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type YArrayBase struct {
	AbstractType
	_searchMarkers *ArraySearchMarkerCollection
}

func NewYArrayBase() *YArrayBase {
	return &YArrayBase{
		AbstractType:   NewAbstractType(),
		_searchMarkers: NewArraySearchMarkerCollection(),
	}
}

// Integrate integrates this type with a document and item
func (y *YArrayBase) Integrate(doc *utils.YDoc, item *structs.Item) {
	y.AbstractType.Integrate(doc, item)
}

func (y *YArrayBase) ClearSearchMarkers() {
	y._searchMarkers.Clear()
}

func (y *YArrayBase) CallObserver(transaction *utils.Transaction, parentSubs map[string]struct{}) {
	if !transaction.Local {
		y._searchMarkers.Clear()
	}
}

func (y *YArrayBase) InsertGenerics(transaction *utils.Transaction, index int, content []interface{}) {
	if index == 0 {
		if y._searchMarkers != nil && y._searchMarkers.Count() > 0 {
			y._searchMarkers.UpdateMarkerChanges(index, len(content))
		}

		y.InsertGenericsAfter(transaction, nil, content)
		return
	}

	startIndex := index
	var marker *ArraySearchMarker
	if y._searchMarkers != nil {
		marker = y.FindMarker(index)
	}

	n := y._start
	if marker != nil {
		n = marker.p
		index -= marker.index

		// We need to iterate one to the left so that the algorithm works.
		if index == 0 {
			if leftItem, ok := n.Prev.(*structs.Item); ok {
				n = leftItem
				if n != nil && n.Countable && !n.Deleted {
					index += n.Length
				}
			}
		}
	}

	for n != nil {
		if !n.Deleted && n.Countable {
			if index <= n.Length {
				if index < n.Length {
					// insert in-between
					transaction.Doc.Store.GetItemCleanStart(transaction, utils.NewID(n.Id.Client, n.Id.Clock+index))
				}
				break
			}
			index -= n.Length
		}
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	if y._searchMarkers != nil && y._searchMarkers.Count() > 0 {
		y._searchMarkers.UpdateMarkerChanges(startIndex, len(content))
	}

	y.InsertGenericsAfter(transaction, n, content)
}

func (y *YArrayBase) InsertGenericsAfter(transaction *utils.Transaction, referenceItem *structs.Item, content []interface{}) {
	left := referenceItem
	doc := transaction.Doc
	ownClientId := doc.ClientId
	store := doc.Store

	var right *structs.Item
	if referenceItem == nil {
		right = y._start
	} else {
		if rightItem, ok := referenceItem.Right.(*structs.Item); ok {
			right = rightItem
		}
	}

	jsonContent := make([]interface{}, 0)

	packJsonContent := func() {
		if len(jsonContent) > 0 {
			var lastId *utils.ID
			if left != nil {
				lastId = left.LastId
			}

			left = structs.NewItem(
				utils.NewID(ownClientId, store.GetState(ownClientId)),
				left, lastId, right, nil, y, "",
				structs.NewContentAny(jsonContent),
			)
			left.Integrate(transaction, 0)
			jsonContent = make([]interface{}, 0)
		}
	}

	for _, c := range content {
		switch v := c.(type) {
		case []byte:
			packJsonContent()
			var lastId *utils.ID
			if left != nil {
				lastId = left.LastId
			}
			left = structs.NewItem(
				utils.NewID(ownClientId, store.GetState(ownClientId)),
				left, lastId, right, nil, y, "",
				structs.NewContentBinary(v),
			)
			left.Integrate(transaction, 0)
		case *utils.YDoc:
			packJsonContent()
			var lastId *utils.ID
			if left != nil {
				lastId = left.LastId
			}
			left = structs.NewItem(
				utils.NewID(ownClientId, store.GetState(ownClientId)),
				left, lastId, right, nil, y, "",
				structs.NewContentDoc(v),
			)
			left.Integrate(transaction, 0)
		case *AbstractType:
			packJsonContent()
			var lastId *utils.ID
			if left != nil {
				lastId = left.LastId
			}
			left = structs.NewItem(
				utils.NewID(ownClientId, store.GetState(ownClientId)),
				left, lastId, right, nil, y, "",
				structs.NewContentType(v),
			)
			left.Integrate(transaction, 0)
		default:
			jsonContent = append(jsonContent, c)
		}
	}

	packJsonContent()
}

func (y *YArrayBase) Delete(transaction *utils.Transaction, index int, length int) {
	if length == 0 {
		return
	}

	startIndex := index
	startLength := length
	var marker *ArraySearchMarker
	if y._searchMarkers != nil {
		marker = y.FindMarker(index)
	}

	n := y._start
	if marker != nil {
		n = marker.p
		index -= marker.index
	}

	// Compute the first item to be deleted
	for n != nil && index > 0 {
		if !n.Deleted && n.Countable {
			if index < n.Length {
				transaction.Doc.Store.GetItemCleanStart(transaction, utils.NewID(n.Id.Client, n.Id.Clock+index))
			}
			index -= n.Length
		}
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	// Delete all items until done
	for length > 0 && n != nil {
		if !n.Deleted {
			if length < n.Length {
				transaction.Doc.Store.GetItemCleanStart(transaction, utils.NewID(n.Id.Client, n.Id.Clock+length))
			}
			n.Delete(transaction)
			length -= n.Length
		}
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	if length > 0 {
		panic("Array length exceeded")
	}

	if y._searchMarkers != nil && y._searchMarkers.Count() > 0 {
		y._searchMarkers.UpdateMarkerChanges(startIndex, -startLength+length)
	}
}

func (y *YArrayBase) InternalSlice(start int, end int) []interface{} {
	if start < 0 {
		start += y.Length
	}

	if end < 0 {
		end += y.Length
	}

	if start < 0 {
		panic("start < 0")
	}

	if end < 0 {
		panic("end < 0")
	}

	if start > end {
		panic("start > end")
	}

	length := end - start
	cs := make([]interface{}, 0)

	n := y._start

	for n != nil && length > 0 {
		if n.Countable && !n.Deleted {
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
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	return cs
}

func (y *YArrayBase) ForEach(fun func(interface{}, int, *YArrayBase)) {
	index := 0
	n := y._start

	for n != nil {
		if n.Countable && !n.Deleted {
			c := n.Content.GetContent()
			for _, cItem := range c {
				fun(cItem, index, y)
				index++
			}
		}
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}
}

func (y *YArrayBase) ForEachSnapshot(fun func(interface{}, int, *YArrayBase), snapshot *utils.Snapshot) {
	index := 0
	n := y._start

	for n != nil {
		if n.Countable && snapshot.IsVisible(n) {
			c := n.Content.GetContent()
			for _, value := range c {
				fun(value, index, y)
				index++
			}
		}
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}
}

func (y *YArrayBase) EnumerateContent() []interface{} {
	n := y._start
	cs := make([]interface{}, 0)

	for n != nil {
		for n != nil && n.Deleted {
			if rightItem, ok := n.Right.(*structs.Item); ok {
				n = rightItem
			} else {
				n = nil
			}
		}

		// Check if we reached the end
		if n == nil {
			break
		}

		c := n.Content.GetContent()
		for _, cItem := range c {
			cs = append(cs, cItem)
		}

		// We used content of n, now iterate to next
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	return cs
}

func (y *YArrayBase) FindMarker(index int) *ArraySearchMarker {
	if y._start == nil || index == 0 || y._searchMarkers == nil || y._searchMarkers.Count() == 0 {
		return nil
	}

	var marker *ArraySearchMarker
	if y._searchMarkers.Count() > 0 {
		y._searchMarkers.mutex.RLock()
		for _, m := range y._searchMarkers.searchMarkers {
			if marker == nil || math.Abs(float64(index-m.index)) < math.Abs(float64(index-marker.index)) {
				marker = m
			}
		}
		y._searchMarkers.mutex.RUnlock()
	}

	p := y._start
	pIndex := 0

	if marker != nil {
		p = marker.p
		pIndex = marker.index
		// We used it, we might need to use it again
		marker.RefreshTimestamp()
	}

	// Iterate to right if possible
	for p.Right != nil && pIndex < index {
		if !p.Deleted && p.Countable {
			if index < pIndex+p.Length {
				break
			}
			pIndex += p.Length
		}
		if rightItem, ok := p.Right.(*structs.Item); ok {
			p = rightItem
		} else {
			break
		}
	}

	// Iterate to left if necessary (might be that pIndex > index)
	for p.Left != nil && pIndex > index {
		if leftItem, ok := p.Left.(*structs.Item); ok {
			p = leftItem
			if p != nil && !p.Deleted && p.Countable {
				pIndex -= p.Length
			}
		} else {
			break
		}
	}

	// We want to make sure that p can't be merged with left
	// Iterate to left until p can't be merged with left
	for p.Left != nil {
		if leftItem, ok := p.Left.(*structs.Item); ok {
			if leftItem.Id.Client == p.Id.Client && leftItem.Id.Clock+leftItem.Length == p.Id.Clock {
				p = leftItem
				if p != nil && !p.Deleted && p.Countable {
					pIndex -= p.Length
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	if marker != nil && math.Abs(float64(marker.index-pIndex)) < float64(y.Length)/MaxSearchMarkers {
		// Adjust existing marker
		marker.Update(p, pIndex)
		return marker
	} else {
		// Create a new marker
		return y._searchMarkers.MarkPosition(p, pIndex)
	}
}
