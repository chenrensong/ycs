package types

import (
	"strings"

	"github.com/chenrensong/ygo/utils"
)

const YTextRefId = 2

// Delta represents a delta operation for text changes
type Delta struct {
	Insert     interface{}            `json:"insert,omitempty"`
	Delete     *int                   `json:"delete,omitempty"`
	Retain     *int                   `json:"retain,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// YTextChangeType represents the type of text change
type YTextChangeType int

const (
	YTextChangeTypeAdded YTextChangeType = iota
	YTextChangeTypeRemoved
)

const ChangeKey = "ychange"

// YTextChangeAttributes represents change attributes
type YTextChangeAttributes struct {
	Type  YTextChangeType
	User  uint64
	State YTextChangeType
}

// YTextEvent represents an event for YText changes
type YTextEvent struct {
	*utils.YEvent
	KeysChanged      map[string]struct{}
	ChildListChanged bool
	delta            []Delta
}

func NewYTextEvent(text *YText, transaction *utils.Transaction, subs map[string]struct{}) *YTextEvent {
	event := &YTextEvent{
		YEvent:           utils.NewYEvent(&text.YArrayBase.AbstractType, transaction),
		KeysChanged:      make(map[string]struct{}),
		ChildListChanged: false,
	}

	if len(subs) > 0 {
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

// ItemTextListPosition represents a position in the text list
type ItemTextListPosition struct {
	Left              *structs.Item
	Right             *structs.Item
	Index             int
	CurrentAttributes map[string]interface{}
}

func NewItemTextListPosition(left *structs.Item, right *structs.Item, index int, currentAttributes map[string]interface{}) *ItemTextListPosition {
	return &ItemTextListPosition{
		Left:              left,
		Right:             right,
		Index:             index,
		CurrentAttributes: currentAttributes,
	}
}

// YText represents a text type with formatting information
type YText struct {
	*YArrayBase
	_prelimContent []interface{}
	_pending       []func()
}

func NewYText() *YText {
	return &YText{
		YArrayBase: NewYArrayBase(),
		_pending:   make([]func(), 0),
	}
}

func (y *YText) Write(encoder utils.IUpdateEncoder) {
	encoder.WriteTypeRef(YTextRefId)
}

func ReadYText(decoder utils.IUpdateEncoder) *YText {
	return NewYText()
}

func (y *YText) Insert(index int, text string, attributes ...map[string]interface{}) {
	if text == "" {
		return
	}

	var attrs map[string]interface{}
	if len(attributes) > 0 {
		attrs = attributes[0]
	}

	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			pos := y.FindPosition(tr, index)
			if attrs == nil {
				attrs = make(map[string]interface{})
				for k, v := range pos.CurrentAttributes {
					attrs[k] = v
				}
			}
			y.InsertText(tr, pos, text, attrs)
		})
	} else {
		y._pending = append(y._pending, func() {
			y.Insert(index, text, attrs)
		})
	}
}

func (y *YText) InsertEmbed(index int, embed interface{}, attributes ...map[string]interface{}) {
	var attrs map[string]interface{}
	if len(attributes) > 0 {
		attrs = attributes[0]
	} else {
		attrs = make(map[string]interface{})
	}

	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			pos := y.FindPosition(tr, index)
			y.InsertText(tr, pos, embed, attrs)
		})
	} else {
		y._pending = append(y._pending, func() {
			y.InsertEmbed(index, embed, attrs)
		})
	}
}

func (y *YText) Delete(start, length int) {
	if length == 0 {
		return
	}

	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			pos := y.FindPosition(tr, start)
			y.DeleteText(tr, pos, length)
		})
	} else {
		y._pending = append(y._pending, func() {
			y.Delete(start, length)
		})
	}
}

func (y *YText) Format(start, length int, attributes map[string]interface{}) {
	if length == 0 {
		return
	}

	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			pos := y.FindPosition(tr, start)
			if pos.Right == nil {
				return
			}
			y.FormatText(tr, pos, length, attributes)
		})
	} else {
		y._pending = append(y._pending, func() {
			y.Format(start, length, attributes)
		})
	}
}

func (y *YText) Get() string {
	var sb strings.Builder

	n := y._start
	for n != nil {
		if !n.Deleted && n.Countable {
			if cs, ok := n.Content.(*structs.ContentString); ok {
				sb.WriteString(cs.GetString())
			}
		}
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	return sb.String()
}

func (y *YText) ToString() string {
	return y.Get()
}

func (y *YText) Clone() *YText {
	return y.InternalClone().(*YText)
}

func (y *YText) InternalCopy() *AbstractType {
	text := NewYText()
	return &text.YArrayBase.AbstractType
}

func (y *YText) InternalClone() *AbstractType {
	text := NewYText()
	content := y.Get()
	if content != "" {
		text.Insert(0, content)
	}
	return &text.YArrayBase.AbstractType
}

func (y *YText) Integrate(doc *utils.YDoc, item *structs.Item) {
	y.YArrayBase.Integrate(doc, item)

	// Execute pending operations
	for _, pending := range y._pending {
		pending()
	}
	y._pending = nil
}

func (y *YText) CallObserver(transaction *utils.Transaction, parentSubs map[string]struct{}) {
	y.YArrayBase.CallObserver(transaction, parentSubs)
	y.CallTypeObservers(transaction, NewYTextEvent(y, transaction, parentSubs).YEvent)
}

func (y *YText) FindPosition(transaction *utils.Transaction, index int) *ItemTextListPosition {
	currentAttributes := make(map[string]interface{})

	if y._start == nil {
		return NewItemTextListPosition(nil, nil, 0, currentAttributes)
	}

	var left *structs.Item
	right := y._start
	currentIndex := 0

	for right != nil && currentIndex < index {
		if !right.Deleted && right.Countable {
			if currentIndex+right.Length <= index {
				currentIndex += right.Length
				left = right
				if rightItem, ok := right.Right.(*structs.Item); ok {
					right = rightItem
				} else {
					right = nil
				}
			} else {
				// Need to split the item
				if transaction != nil {
					splitPos := index - currentIndex
					transaction.Doc.Store.GetItemCleanStart(transaction, utils.NewID(right.Id.Client, right.Id.Clock+splitPos))
				}
				break
			}
		} else {
			// Handle format items
			if cf, ok := right.Content.(*structs.ContentFormat); ok && !right.Deleted {
				UpdateCurrentAttributes(currentAttributes, cf)
			}
			left = right
			if rightItem, ok := right.Right.(*structs.Item); ok {
				right = rightItem
			} else {
				right = nil
			}
		}
	}

	return NewItemTextListPosition(left, right, index, currentAttributes)
}

func (y *YText) InsertText(transaction *utils.Transaction, pos *ItemTextListPosition, content interface{}, attributes map[string]interface{}) {
	// Insert format items for attributes that differ from current
	for key, value := range attributes {
		if currentValue, exists := pos.CurrentAttributes[key]; !exists || !EqualAttrs(currentValue, value) {
			// Insert format item
			var lastId *utils.ID
			if pos.Left != nil {
				lastId = pos.Left.LastId
			}

			formatItem := structs.NewItem(
				utils.NewID(transaction.Doc.ClientId, transaction.Doc.Store.GetState(transaction.Doc.ClientId)),
				pos.Left, lastId, pos.Right, nil, y, "",
				structs.NewContentFormat(key, value),
			)
			formatItem.Integrate(transaction, 0)
			pos.Left = formatItem
		}
	}

	// Insert the actual content
	var contentStruct structs.IContent
	if text, ok := content.(string); ok {
		contentStruct = structs.NewContentString(text)
	} else {
		contentStruct = structs.NewContentEmbed(content)
	}

	var lastId *utils.ID
	if pos.Left != nil {
		lastId = pos.Left.LastId
	}

	item := structs.NewItem(
		utils.NewID(transaction.Doc.ClientId, transaction.Doc.Store.GetState(transaction.Doc.ClientId)),
		pos.Left, lastId, pos.Right, nil, y, "",
		contentStruct,
	)
	item.Integrate(transaction, 0)
}

func (y *YText) DeleteText(transaction *utils.Transaction, pos *ItemTextListPosition, length int) {
	remaining := length
	current := pos.Right

	for current != nil && remaining > 0 {
		if !current.Deleted && current.Countable {
			if current.Length <= remaining {
				current.Delete(transaction)
				remaining -= current.Length
			} else {
				// Split and delete part of the item
				transaction.Doc.Store.GetItemCleanStart(transaction, utils.NewID(current.Id.Client, current.Id.Clock+remaining))
				current.Delete(transaction)
				remaining = 0
			}
		}
		if rightItem, ok := current.Right.(*structs.Item); ok {
			current = rightItem
		} else {
			current = nil
		}
	}
}

func (y *YText) FormatText(transaction *utils.Transaction, pos *ItemTextListPosition, length int, attributes map[string]interface{}) {
	remaining := length
	current := pos.Right

	for current != nil && remaining > 0 {
		if !current.Deleted && current.Countable {
			// Apply formatting attributes
			for key, value := range attributes {
				var lastId *utils.ID
				if current != nil {
					lastId = current.LastId
				}

				formatItem := structs.NewItem(
					utils.NewID(transaction.Doc.ClientId, transaction.Doc.Store.GetState(transaction.Doc.ClientId)),
					current, lastId, current.Right, nil, y, "",
					structs.NewContentFormat(key, value),
				)
				formatItem.Integrate(transaction, 0)
			}

			if current.Length <= remaining {
				remaining -= current.Length
			} else {
				remaining = 0
			}
		}
		if rightItem, ok := current.Right.(*structs.Item); ok {
			current = rightItem
		} else {
			current = nil
		}
	}
}

func (y *YText) ToDelta() []Delta {
	deltas := make([]Delta, 0)
	currentAttributes := make(map[string]interface{})
	var str strings.Builder

	packStr := func() {
		if str.Len() > 0 {
			delta := Delta{Insert: str.String()}
			if len(currentAttributes) > 0 {
				delta.Attributes = make(map[string]interface{})
				for k, v := range currentAttributes {
					if v != nil {
						delta.Attributes[k] = v
					}
				}
			}
			deltas = append(deltas, delta)
			str.Reset()
		}
	}

	n := y._start
	for n != nil {
		if !n.Deleted {
			switch content := n.Content.(type) {
			case *structs.ContentString:
				str.WriteString(content.GetString())
			case *structs.ContentEmbed:
				packStr()
				delta := Delta{Insert: content.Embed}
				if len(currentAttributes) > 0 {
					delta.Attributes = make(map[string]interface{})
					for k, v := range currentAttributes {
						delta.Attributes[k] = v
					}
				}
				deltas = append(deltas, delta)
			case *structs.ContentFormat:
				packStr()
				UpdateCurrentAttributes(currentAttributes, content)
			}
		}
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			n = nil
		}
	}

	packStr()
	return deltas
}

func (y *YText) ApplyDelta(delta []Delta, sanitize ...bool) {
	shouldSanitize := false
	if len(sanitize) > 0 {
		shouldSanitize = sanitize[0]
	}

	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			curPos := NewItemTextListPosition(nil, y._start, 0, make(map[string]interface{}))

			for i, op := range delta {
				if op.Insert != nil {
					insertStr, isString := op.Insert.(string)
					var ins interface{} = op.Insert

					if !shouldSanitize && isString && i == len(delta)-1 && curPos.Right == nil && strings.HasSuffix(insertStr, "\n") {
						ins = insertStr[:len(insertStr)-1]
					}

					if !isString || len(ins.(string)) > 0 {
						attrs := op.Attributes
						if attrs == nil {
							attrs = make(map[string]interface{})
						}
						y.InsertText(tr, curPos, ins, attrs)
					}
				} else if op.Retain != nil {
					attrs := op.Attributes
					if attrs == nil {
						attrs = make(map[string]interface{})
					}
					y.FormatText(tr, curPos, *op.Retain, attrs)
				} else if op.Delete != nil {
					y.DeleteText(tr, curPos, *op.Delete)
				}
			}
		})
	} else {
		y._pending = append(y._pending, func() {
			y.ApplyDelta(delta, shouldSanitize)
		})
	}
}

// Helper functions
func UpdateCurrentAttributes(currentAttributes map[string]interface{}, contentFormat *structs.ContentFormat) {
	if contentFormat.Value == nil {
		delete(currentAttributes, contentFormat.Key)
	} else {
		currentAttributes[contentFormat.Key] = contentFormat.Value
	}
}

func EqualAttrs(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a == b
}

func (y *YText) SetAttribute(name string, value interface{}) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			y.TypeMapSet(tr, name, value)
		})
	} else {
		y._pending = append(y._pending, func() {
			y.SetAttribute(name, value)
		})
	}
}

func (y *YText) RemoveAttribute(name string) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			y.TypeMapDelete(tr, name)
		})
	} else {
		y._pending = append(y._pending, func() {
			y.RemoveAttribute(name)
		})
	}
}

func (y *YText) Observe(observer func()) {
	// Implementation for observing changes to the text
}

func (y *YText) Unobserve(observer func()) {
	// Implementation for stopping observation of changes
}
