// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"ycs-golang/structs"
	"ycs-golang/types"
)

// AbsolutePosition represents an absolute position in a document
type AbsolutePosition struct {
	Type  *types.AbstractType
	Index int
	Assoc int
}

// NewAbsolutePosition creates a new AbsolutePosition
func NewAbsolutePosition(typ *types.AbstractType, index, assoc int) *AbsolutePosition {
	return &AbsolutePosition{
		Type:  typ,
		Index: index,
		Assoc: assoc,
	}
}

// TryCreateFromRelativePosition tries to create an absolute position from a relative position
func TryCreateFromRelativePosition(rpos *RelativePosition, doc *YDoc) *AbsolutePosition {
	store := doc.Store
	rightId := rpos.Item
	typeId := rpos.TypeId
	tName := rpos.TName
	assoc := rpos.Assoc
	index := 0
	var typ *types.AbstractType

	if rightId != nil {
		if store.GetState(rightId.Client) <= rightId.Clock {
			return nil
		}

		res, diff := store.FollowRedone(*rightId)
		right, ok := res.(*structs.Item)
		if !ok || right == nil {
			return nil
		}

		typ, ok = right.Parent.(*types.AbstractType)
		if !ok || typ == nil {
			return nil
		}

		if typ.Item == nil || !typ.Item.Deleted {
			// Adjust position based on the left association, if necessary.
			if right.Deleted || !right.Countable {
				index = 0
			} else {
				index = diff
				if assoc >= 0 {
					index += 0
				} else {
					index += 1
				}
			}
			
			n := right.Left
			for n != nil {
				if nItem, ok := n.(*structs.Item); ok {
					if !nItem.Deleted && nItem.Countable {
						index += nItem.Length
					}
					n = nItem.Left
				} else {
					break
				}
			}
		}
	} else {
		if tName != "" {
			// Note: This requires implementing the Get method for YDoc
			// typ = doc.Get(tName).(*types.AbstractType)
		} else if typeId != nil {
			if store.GetState(typeId.Client) <= typeId.Clock {
				// Type does not exist yet.
				return nil
			}

			res, _ := store.FollowRedone(*typeId)
			item, ok := res.(*structs.Item)
			if !ok || item == nil {
				// Struct is garbage collected.
				return nil
			}

			if contentType, ok := item.Content.(*structs.ContentType); ok {
				typ = contentType.Type
			} else {
				// Struct is garbage collected.
				return nil
			}
		} else {
			panic("Invalid relative position")
		}

		if assoc >= 0 {
			index = typ.Length
		} else {
			index = 0
		}
	}

	return NewAbsolutePosition(typ, index, assoc)
}