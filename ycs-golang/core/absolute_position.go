package core

import (
	"fmt"
	"ycs/contracts"
)

// AbsolutePosition represents an absolute position in the document
type AbsolutePosition struct {
	Type  contracts.IAbstractType
	Index int
	Assoc int
}

// NewAbsolutePosition creates a new AbsolutePosition
func NewAbsolutePosition(typ contracts.IAbstractType, index int, assoc int) *AbsolutePosition {
	return &AbsolutePosition{
		Type:  typ,
		Index: index,
		Assoc: assoc,
	}
}

// TryCreateFromRelativePosition tries to create an AbsolutePosition from a RelativePosition
func TryCreateFromRelativePosition(rpos *RelativePosition, doc contracts.IYDoc) *AbsolutePosition {
	store := doc.GetStore()
	rightId := rpos.Item
	typeId := rpos.TypeId
	tName := rpos.TName
	assoc := rpos.Assoc
	index := 0
	var typ contracts.IAbstractType

	if rightId != nil {
		if store.GetState(rightId.Client) <= rightId.Clock {
			return nil
		}

		item, diff := store.FollowRedone(*rightId)
		right, ok := item.(contracts.IStructItem)
		if !ok {
			return nil
		}

		typ, ok = right.GetParent().(contracts.IAbstractType)
		if !ok {
			return nil
		}

		if typ.GetItem() == nil || !typ.GetItem().GetDeleted() {
			// Adjust position based on the left association, if necessary.
			if right.GetDeleted() || !right.GetCountable() {
				index = 0
			} else {
				if assoc >= 0 {
					index = diff
				} else {
					index = diff + 1
				}
			}

			n, ok := right.GetLeft().(contracts.IStructItem)
			for ok && n != nil {
				if !n.GetDeleted() && n.GetCountable() {
					index += n.GetLength()
				}
				n, ok = n.GetLeft().(contracts.IStructItem)
			}
		}
	} else {
		if tName != "" {
			typ = doc.Get(tName, func() contracts.IAbstractType {
				return nil // This should not happen in normal cases
			})
		} else if typeId != nil {
			if store.GetState(typeId.Client) <= typeId.Clock {
				// Type does not exist yet.
				return nil
			}

			item, _ := store.FollowRedone(*typeId)
			structItem, ok := item.(contracts.IStructItem)
			if !ok {
				return nil
			}

			// Check if content is ContentType
			content := structItem.GetContent()
			if contentType, ok := content.(interface {
				GetType() contracts.IAbstractType
			}); ok {
				typ = contentType.GetType()
			} else {
				// Struct is garbage collected.
				return nil
			}
		} else {
			panic(fmt.Errorf("invalid relative position"))
		}

		if assoc >= 0 {
			index = typ.GetLength()
		} else {
			index = 0
		}
	}

	return NewAbsolutePosition(typ, index, assoc)
}
