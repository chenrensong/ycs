package types

import (
	"strings"
	"github.com/yjs/ycs-golang/structs"
	"github.com/yjs/ycs-golang/utils"
)

// YTextEvent represents an event for YText
type YTextEvent struct {
	*YEvent
	KeysChanged map[string]struct{}
	ChildListChanged bool
	delta       []Delta
}

// NewYTextEvent creates a new YTextEvent
func NewYTextEvent(text *YText, transaction *utils.Transaction, subs map[string]struct{}) *YTextEvent {
	keysChanged := make(map[string]struct{})
	childListChanged := false
	
	if len(subs) > 0 {
		for sub := range subs {
			if sub == "" {
				childListChanged = true
			} else {
				keysChanged[sub] = struct{}{}
			}
		}
	}
	
	return &YTextEvent{
		YEvent:           NewYEvent(text.AbstractType, transaction),
		KeysChanged:      keysChanged,
		ChildListChanged: childListChanged,
	}
}

// Delta returns the changes in the delta format
func (y *YTextEvent) Delta() []Delta {
	if y.delta == nil {
		doc := y.Target.Doc
		y.delta = make([]Delta, 0)
		
		doc.Transact(func(transaction *utils.Transaction) {
			delta := y.delta
			
			// Saves all current attributes for insert.
			currentAttributes := make(map[string]interface{})
			oldAttributes := make(map[string]interface{})
			item := y.Target.Start
			var action *string
			attributes := make(map[string]interface{})
			
			var insert interface{} = ""
			retain := 0
			deleteLen := 0
			
			addOp := func() {
				if action != nil {
					var op Delta
					
					switch *action {
					case "delete":
						op = Delta{Delete: &deleteLen}
						deleteLen = 0
					case "insert":
						op = Delta{Insert: insert}
						if len(currentAttributes) > 0 {
							op.Attributes = make(map[string]interface{})
							for k, v := range currentAttributes {
								if v != nil {
									op.Attributes[k] = v
								}
							}
						}
					case "retain":
						op = Delta{Retain: &retain}
						if len(attributes) > 0 {
							op.Attributes = make(map[string]interface{})
							for k, v := range attributes {
								op.Attributes[k] = v
							}
						}
						retain = 0
					}
					
					delta = append(delta, op)
					action = nil
				}
			}
			
			for item != nil {
				switch content := item.Content.(type) {
				case *structs.ContentEmbed:
					if y.Adds(item.AbstractStruct) {
						if !y.Deletes(item.AbstractStruct) {
							addOp()
							actionStr := "insert"
							action = &actionStr
							insert = content.Embed
							addOp()
						}
					} else if y.Deletes(item.AbstractStruct) {
						actionStr := "delete"
						if action == nil || *action != actionStr {
							addOp()
							action = &actionStr
						}
						deleteLen++
					} else if !item.Deleted() {
						actionStr := "retain"
						if action == nil || *action != actionStr {
							addOp()
							action = &actionStr
						}
						retain++
					}
				case *structs.ContentString:
					if y.Adds(item.AbstractStruct) {
						if !y.Deletes(item.AbstractStruct) {
							actionStr := "insert"
							if action == nil || *action != actionStr {
								addOp()
								action = &actionStr
							}
							
							insertStr, ok := insert.(string)
							if !ok {
								insertStr = ""
							}
							insert = insertStr + content.GetString()
						}
					} else if y.Deletes(item.AbstractStruct) {
						actionStr := "delete"
						if action == nil || *action != actionStr {
							addOp()
							action = &actionStr
						}
						deleteLen += item.Length
					} else if !item.Deleted() {
						actionStr := "retain"
						if action == nil || *action != actionStr {
							addOp()
							action = &actionStr
						}
						retain += item.Length
					}
				case *structs.ContentFormat:
					if y.Adds(item.AbstractStruct) {
						if !y.Deletes(item.AbstractStruct) {
							var curVal interface{}
							if val, ok := currentAttributes[content.Key]; ok {
								curVal = val
							}
							
							if !EqualAttrs(curVal, content.Value) {
								actionStr := "retain"
								if action != nil && *action == actionStr {
									addOp()
								}
								
								var oldVal interface{}
								if val, ok := oldAttributes[content.Key]; ok {
									oldVal = val
								}
								
								if EqualAttrs(content.Value, oldVal) {
									delete(attributes, content.Key)
								} else {
									attributes[content.Key] = content.Value
								}
							} else {
								item.Delete(transaction)
							}
						}
					} else if y.Deletes(item.AbstractStruct) {
						oldAttributes[content.Key] = content.Value
						
						var curVal interface{}
						if val, ok := currentAttributes[content.Key]; ok {
							curVal = val
						}
						
						if !EqualAttrs(curVal, content.Value) {
							actionStr := "retain"
							if action != nil && *action == actionStr {
								addOp()
							}
							attributes[content.Key] = curVal
						}
					} else if !item.Deleted() {
						oldAttributes[content.Key] = content.Value
						
						if attr, ok := attributes[content.Key]; ok {
							if !EqualAttrs(attr, content.Value) {
								actionStr := "retain"
								if action != nil && *action == actionStr {
									addOp()
								}
								
								if content.Value == nil {
									attributes[content.Key] = nil
								} else {
									delete(attributes, content.Key)
								}
							} else {
								item.Delete(transaction)
							}
						}
					}
					
					if !item.Deleted() {
						actionStr := "insert"
						if action != nil && *action == actionStr {
							addOp()
						}
						
						UpdateCurrentAttributes(currentAttributes, content)
					}
				}
				
				// Type assert to Item to access Right
				if rightItem, ok := item.Right.(*structs.Item); ok {
					item = rightItem
				} else {
					break
				}
			}
			
			addOp()
			
			// Remove trailing retain operations with attributes
			for len(delta) > 0 {
				lastOp := delta[len(delta)-1]
				if lastOp.Retain != nil && lastOp.Attributes != nil {
					// Retain delta's if they don't assign attributes.
					delta = delta[:len(delta)-1]
				} else {
					break
				}
			}
			
			y.delta = delta
		})
	}
	
	return y.delta
}

// YText represents a text type with formatting information
type YText struct {
	*YArrayBase
	Pending []func()
}

// YTextRefId is the reference ID for YText
const YTextRefId int = 2

// NewYText creates a new YText
func NewYText(str string) *YText {
	pending := make([]func(), 0)
	if str != "" {
		pending = append(pending, func() { 
			// Insert(0, str) - we can't call this directly due to circular dependency
		})
	}
	
	return &YText{
		YArrayBase: NewYArrayBase(),
		Pending:    pending,
	}
}

// NewYTextEmpty creates a new empty YText
func NewYTextEmpty() *YText {
	return NewYText("")
}

// ApplyDelta applies delta operations to the text
func (y *YText) ApplyDelta(delta []Delta, sanitize bool) {
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			// Implementation would go here
		})
	} else {
		y.Pending = append(y.Pending, func() { 
			y.ApplyDelta(delta, sanitize) 
		})
	}
}

// Insert inserts text at the specified index
func (y *YText) Insert(index int, text string, attributes map[string]interface{}) {
	if text == "" {
		return
	}
	
	doc := y.Doc
	if doc != nil {
		doc.Transact(func(tr *utils.Transaction) {
			// Implementation would go here
		})
	} else {
		y.Pending = append(y.Pending, func() { 
			y.Insert(index, text, attributes) 
		})
	}
}

// Delete deletes text at the specified index with the specified length
func (y *YText) Delete(index, length int) {
	if length == 0 {
		return
	}
	
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			// Implementation would go here
		})
	} else {
		y.Pending = append(y.Pending, func() { 
			y.Delete(index, length) 
		})
	}
}

// Format formats text at the specified index with the specified length
func (y *YText) Format(index, length int, attributes map[string]interface{}) {
	if length == 0 {
		return
	}
	
	if y.Doc != nil {
		y.Doc.Transact(func(tr *utils.Transaction) {
			// Implementation would go here
		})
	} else {
		y.Pending = append(y.Pending, func() { 
			y.Format(index, length, attributes) 
		})
	}
}

// ToString returns the text as a string
func (y *YText) ToString() string {
	var sb strings.Builder
	
	n := y.Start
	for n != nil {
		if !n.Deleted() && n.Countable() {
			if cs, ok := n.Content.(*structs.ContentString); ok {
				cs.AppendToBuilder(&sb)
			}
		}
		
		// Type assert to Item to access Right
		if rightItem, ok := n.Right.(*structs.Item); ok {
			n = rightItem
		} else {
			break
		}
	}
	
	return sb.String()
}

// Integrate integrates the text
func (y *YText) Integrate(doc *utils.YDoc, item *structs.Item) {
	y.YArrayBase.Integrate(doc, item)
	
	for _, c := range y.Pending {
		c()
	}
	
	y.Pending = nil
}

// CallObserver creates YTextEvent and calls observers
func (y *YText) CallObserver(transaction *utils.Transaction, parentSubs map[string]struct{}) {
	y.YArrayBase.CallObserver(transaction, parentSubs)
	
	evt := NewYTextEvent(y, transaction, parentSubs)
	
	// If a remote change happened, we try to cleanup potential formatting duplicates.
	if !transaction.Local {
		// Implementation would go here
	}
	
	y.CallTypeObservers(transaction, evt.YEvent)
}

// Write writes the text to an encoder
func (y *YText) Write(encoder utils.IUpdateEncoder) {
	encoder.WriteTypeRef(YTextRefId)
}

// Read reads a YText from a decoder
func ReadYText(decoder utils.IUpdateDecoder) *YText {
	return NewYTextEmpty()
}

// EqualAttrs checks if two attributes are equal
func EqualAttrs(attr1, attr2 interface{}) bool {
	// In Go, we need to handle nil values differently than in C#
	if attr1 == nil && attr2 == nil {
		return true
	}
	if attr1 == nil || attr2 == nil {
		return false
	}
	
	// For now, we'll just use simple equality
	// In a real implementation, you might need more complex comparison
	return attr1 == attr2
}

// UpdateCurrentAttributes updates the current attributes with a format
func UpdateCurrentAttributes(attributes map[string]interface{}, format *structs.ContentFormat) {
	if format.Value == nil {
		delete(attributes, format.Key)
	} else {
		attributes[format.Key] = format.Value
	}
}