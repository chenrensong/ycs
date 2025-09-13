// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"errors"

	structs "github.com/chenrensong/ygo/structs"
	types "github.com/chenrensong/ygo/types"
)

var (
	ErrInvalidPosition = errors.New("invalid position parameters")
)

// AbsolutePosition represents an absolute position in a Yjs document
type AbsolutePosition struct {
	Type  *types.AbstractType
	Index int
	Assoc int
}

// NewAbsolutePosition creates a new AbsolutePosition instance
func NewAbsolutePosition(t *types.AbstractType, index int, assoc ...int) (*AbsolutePosition, error) {
	a := 0
	if len(assoc) > 0 {
		a = assoc[0]
	}

	if t == nil {
		return nil, ErrInvalidPosition
	}

	return &AbsolutePosition{
		Type:  t,
		Index: index,
		Assoc: a,
	}, nil
}

// TryCreateFromRelativePosition attempts to create an AbsolutePosition from a RelativePosition
func TryCreateFromRelativePosition(rpos *RelativePosition, doc *YDoc) (*AbsolutePosition, error) {
	if rpos == nil || doc == nil {
		return nil, ErrInvalidPosition
	}

	store := doc.Store
	rightID := rpos.Item
	typeID := rpos.TypeID
	tName := rpos.TName
	assoc := rpos.Assoc
	var index int
	var t *types.AbstractType

	if rightID != nil {
		if store.GetState(rightID.Client) <= rightID.Clock {
			return nil, nil // Position doesn't exist yet
		}

		res := store.FollowRedone(*rightID)
		right, ok := res.Item.(*structs.Item)
		if !ok {
			return nil, nil
		}

		t, ok = right.Parent.(*types.AbstractType)
		if !ok {
			return nil, ErrInvalidPosition
		}

		if t.Item == nil || !t.Item.Deleted {
			// Adjust position based on the left association
			if right.Deleted || !right.Countable {
				index = 0
			} else {
				index = res.Diff
				if assoc >= 0 {
					index += 0
				} else {
					index += 1
				}
			}

			n := right.Left
			for n != nil {
				item, ok := n.(*structs.Item)
				if ok && !item.Deleted && item.Countable {
					index += item.Length
				}
				n = item.Left
			}
		}
	} else {
		if tName != "" {
			t = doc.Get(tName)
		} else if typeID != nil {
			if store.GetState(typeID.Client) <= typeID.Clock {
				// Type does not exist yet
				return nil, nil
			}

			item, ok := store.FollowRedone(*typeID).Item.(*structs.Item)
			if !ok {
				return nil, nil
			}

			if content, ok := item.Content.(*types.ContentType); ok {
				t = content.Type
			} else {
				// Struct is garbage collected
				return nil, nil
			}
		} else {
			return nil, ErrInvalidPosition
		}

		if assoc >= 0 {
			index = t.Length
		} else {
			index = 0
		}
	}

	return NewAbsolutePosition(t, index, assoc)
}
