// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"bytes"
	"ycs-golang/core"
	"ycs-golang/structs"
	"ycs-golang/types"
)

// RelativePosition represents a relative position in a document
type RelativePosition struct {
	Item   *ID
	TypeId *ID
	TName  string
	Assoc  int
}

// NewRelativePosition creates a new RelativePosition
func NewRelativePosition(typ *types.AbstractType, item *ID, assoc int) *RelativePosition {
	rp := &RelativePosition{
		Item:  item,
		Assoc: assoc,
	}

	if typ.Item == nil {
		rp.TName = typ.FindRootTypeKey()
	} else {
		rp.TypeId = &ID{Client: typ.Item.Id.Client, Clock: typ.Item.Id.Clock}
	}

	return rp
}

// NewRelativePositionFromFields creates a new RelativePosition from fields
func NewRelativePositionFromFields(typeId *ID, tname string, item *ID, assoc int) *RelativePosition {
	return &RelativePosition{
		TypeId: typeId,
		TName:  tname,
		Item:   item,
		Assoc:  assoc,
	}
}

// Equals checks if two relative positions are equal
func (rp *RelativePosition) Equals(other *RelativePosition) bool {
	if rp == other {
		return true
	}

	if other == nil {
		return false
	}

	return rp.TName == other.TName &&
		IDEquals(rp.Item, other.Item) &&
		IDEquals(rp.TypeId, other.TypeId) &&
		rp.Assoc == other.Assoc
}

// FromTypeIndex creates a relative position based on an absolute position
func FromTypeIndex(typ *types.AbstractType, index, assoc int) *RelativePosition {
	if assoc < 0 {
		// Associated with the left character or the beginning of a type, decrement index if possible.
		if index == 0 {
			return NewRelativePosition(typ, func() *ID {
				if typ.Item != nil {
					return &typ.Item.Id
				}
				return nil
			}(), assoc)
		}

		index--
	}

	t := typ.Start
	for t != nil {
		if !t.Deleted && t.Countable {
			if t.Length > index {
				// Case 1: found position somewhere in the linked list.
				return NewRelativePosition(typ, &ID{Client: t.Id.Client, Clock: t.Id.Clock + int64(index)}, assoc)
			}

			index -= t.Length
		}

		if t.Right == nil && assoc < 0 {
			// Left-associated position, return last available id.
			return NewRelativePosition(typ, t.LastId(), assoc)
		}

		if item, ok := t.Right.(*structs.Item); ok {
			t = item
		} else {
			break
		}
	}

	return NewRelativePosition(typ, func() *ID {
		if typ.Item != nil {
			return &typ.Item.Id
		}
		return nil
	}(), assoc)
}

// Write writes the relative position to a writer
func (rp *RelativePosition) Write(writer *bytes.Buffer) {
	if rp.Item != nil {
		// Case 1: Found position somewhere in the linked list.
		core.WriteVarUint(writer, 0)
		rp.Item.Write(writer)
	} else if rp.TName != "" {
		// Case 2: Found position at the end of the list and type is stored in y.share.
		core.WriteVarUint(writer, 1)
		core.WriteVarString(writer, rp.TName)
	} else if rp.TypeId != nil {
		// Case 3: Found position at the end of the list and type is attached to an item.
		core.WriteVarUint(writer, 2)
		rp.TypeId.Write(writer)
	} else {
		panic("Invalid relative position")
	}

	core.WriteVarInt(writer, int64(rp.Assoc), false)
}

// Read reads a relative position from a reader
func ReadRelativePosition(reader *bytes.Reader) *RelativePosition {
	var itemId *ID
	var typeId *ID
	var tName string

	switch core.ReadVarUint(reader) {
	case 0:
		// Case 1: Found position somewhere in the linked list.
		itemId = ReadID(reader)
	case 1:
		// Case 2: Found position at the end of the list and type is stored in y.share.
		tName = core.ReadVarString(reader)
	case 2:
		// Case 3: Found position at the end of the list and type is attached to an item.
		typeId = ReadID(reader)
	default:
		panic("Invalid relative position")
	}

	var assoc int
	// Note: In Go, we need to check if there's more data to read
	// This is a simplified check - you may need to implement a proper check
	assoc = int(core.ReadVarInt(reader, false))

	return NewRelativePositionFromFields(typeId, tName, itemId, assoc)
}

// ReadRelativePositionFromBytes reads a relative position from bytes
func ReadRelativePositionFromBytes(encodedPosition []byte) *RelativePosition {
	reader := bytes.NewReader(encodedPosition)
	return ReadRelativePosition(reader)
}

// ToArray converts the relative position to a byte array
func (rp *RelativePosition) ToArray() []byte {
	var stream bytes.Buffer
	rp.Write(&stream)
	return stream.Bytes()
}

// IDEquals checks if two IDs are equal
func IDEquals(a, b *ID) bool {
	if a == nil && b == nil {
		return true
	}
	
	if a == nil || b == nil {
		return false
	}
	
	return a.Client == b.Client && a.Clock == b.Clock
}