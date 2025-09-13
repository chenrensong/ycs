package types

import (
	"github.com/chenrensong/ygo/structs"
	"github.com/chenrensong/ygo/utils"
)

const YArrayRefId = 0

type YArrayEvent struct {
	*utils.YEvent
}

func NewYArrayEvent(arr *YArray, transaction *utils.Transaction) *YArrayEvent {
	return &YArrayEvent{
		YEvent: utils.NewYEvent(arr, transaction),
	}
}

type YArray struct {
	*YArrayBase
	_prelimContent []interface{}
}

func NewYArray(prelimContent ...[]interface{}) *YArray {
	arr := &YArray{
		YArrayBase:     NewYArrayBase(),
		_prelimContent: make([]interface{}, 0),
	}

	if len(prelimContent) > 0 && prelimContent[0] != nil {
		arr._prelimContent = prelimContent[0]
	}

	return arr
}

func (y *YArray) GetLength() int {
	if y._prelimContent != nil {
		return len(y._prelimContent)
	}
	return y.YArrayBase.Length
}

func (y *YArray) Clone() *YArray {
	return y.InternalClone().(*YArray)
}

func (y *YArray) InternalCopy() *AbstractType {
	arr := NewYArray()
	return &arr.YArrayBase.AbstractType
}

func (y *YArray) InternalClone() *AbstractType {
	arr := NewYArray()

	for _, item := range y.EnumerateList() {
		if at, ok := item.(*AbstractType); ok {
			arr.Add([]interface{}{at.InternalClone()})
		} else {
			arr.Add([]interface{}{item})
		}
	}

	return &arr.YArrayBase.AbstractType
}

func (y *YArray) Write(encoder utils.IUpdateEncoder) {
	encoder.WriteTypeRef(YArrayRefId)
}

func ReadYArray(decoder utils.IUpdateDecoder) *YArray {
	return NewYArray()
}

func (y *YArray) Integrate(doc *utils.YDoc, item *structs.Item) {
	y.YArrayBase.Integrate(doc, item)
	if y._prelimContent != nil && len(y._prelimContent) > 0 {
		y.Insert(0, y._prelimContent)
		y._prelimContent = nil
	}
}

func (y *YArray) CallObserver(transaction *utils.Transaction, parentSubs map[string]struct{}) {
	y.YArrayBase.CallObserver(transaction, parentSubs)
	y.CallTypeObservers(transaction, NewYArrayEvent(y, transaction).YEvent)
}

func (y *YArray) Insert(index int, content []interface{}) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			y.InsertGenerics(tr, index, content)
		})
	} else {
		if y._prelimContent == nil {
			y._prelimContent = make([]interface{}, 0)
		}

		// Insert content at specified index in _prelimContent
		if index > len(y._prelimContent) {
			index = len(y._prelimContent)
		}

		// Create a new slice with the content inserted
		newContent := make([]interface{}, len(y._prelimContent)+len(content))
		copy(newContent, y._prelimContent[:index])
		copy(newContent[index:], content)
		copy(newContent[index+len(content):], y._prelimContent[index:])
		y._prelimContent = newContent
	}
}

func (y *YArray) Add(content []interface{}) {
	y.Insert(y.GetLength(), content)
}

func (y *YArray) Unshift(content []interface{}) {
	y.Insert(0, content)
}

func (y *YArray) Delete(index int, length ...int) {
	l := 1
	if len(length) > 0 {
		l = length[0]
	}

	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			y.YArrayBase.Delete(tr, index, l)
		})
	} else if y._prelimContent != nil {
		// Remove content from _prelimContent
		start := index
		end := index + l
		if start < 0 {
			start = 0
		}
		if end > len(y._prelimContent) {
			end = len(y._prelimContent)
		}
		if start < end {
			y._prelimContent = append(y._prelimContent[:start], y._prelimContent[end:]...)
		}
	}
}

func (y *YArray) Slice(start ...int) []interface{} {
	startIndex := 0
	if len(start) > 0 {
		startIndex = start[0]
	}
	return y.InternalSlice(startIndex, y.GetLength())
}

func (y *YArray) SliceWithEnd(start int, end int) []interface{} {
	return y.InternalSlice(start, end)
}

func (y *YArray) Get(index int) interface{} {
	if y._prelimContent != nil {
		if index >= 0 && index < len(y._prelimContent) {
			return y._prelimContent[index]
		}
		return nil
	}

	marker := y.FindMarker(index)
	n := y._start

	if marker != nil {
		n = marker.p
		index -= marker.index
	}

	for n != nil {
		if !n.Deleted && n.Countable {
			if index < n.Length {
				content := n.Content.GetContent()
				if len(content) > index {
					return content[index]
				}
			}
			index -= n.Length
		}
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	return nil
}

func (y *YArray) ToArray() []interface{} {
	cs := make([]interface{}, 0)
	for _, item := range y.EnumerateList() {
		cs = append(cs, item)
	}
	return cs
}

func (y *YArray) EnumerateList() []interface{} {
	if y._prelimContent != nil {
		return y._prelimContent
	}

	n := y._start
	result := make([]interface{}, 0)

	for n != nil {
		if n.Countable && !n.Deleted {
			c := n.Content.GetContent()
			for _, item := range c {
				result = append(result, item)
			}
		}

		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	return result
}
