// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"github.com/chenrensong/ygo/contracts"
)

// RelativePosition represents a relative position in the document
type RelativePosition struct {
	typeID *contracts.StructID
	tname  string
	item   *contracts.StructID
	assoc  int
}

// NewRelativePosition creates a new RelativePosition
func NewRelativePosition(typeID *contracts.StructID, tname string, item *contracts.StructID, assoc int) *RelativePosition {
	return &RelativePosition{
		typeID: typeID,
		tname:  tname,
		item:   item,
		assoc:  assoc,
	}
}

// GetTypeID returns the type ID
func (rp *RelativePosition) GetTypeID() *contracts.StructID {
	return rp.typeID
}

// GetTname returns the type name
func (rp *RelativePosition) GetTname() string {
	return rp.tname
}

// GetItem returns the item ID
func (rp *RelativePosition) GetItem() *contracts.StructID {
	return rp.item
}

// GetAssoc returns the association
func (rp *RelativePosition) GetAssoc() int {
	return rp.assoc
}

// Write writes the relative position to an encoder
func (rp *RelativePosition) Write(encoder contracts.IUpdateEncoder) {
	if rp.item != nil {
		encoder.WriteInfo(128) // Has item
		encoder.WriteClient(rp.item.Client)
		encoder.WriteLength(int(rp.item.Clock))
	} else {
		encoder.WriteInfo(0) // No item
	}

	if rp.typeID != nil {
		encoder.WriteClient(rp.typeID.Client)
		encoder.WriteLength(int(rp.typeID.Clock))
	} else if rp.tname != "" {
		encoder.WriteString(rp.tname)
	}

	encoder.WriteLength(rp.assoc)
}

// ReadRelativePosition reads a relative position from a decoder
func ReadRelativePosition(decoder contracts.IUpdateDecoder) *RelativePosition {
	var item *contracts.StructID
	var typeID *contracts.StructID
	var tname string

	info := decoder.ReadInfo()
	if info&128 != 0 { // Has item
		client := decoder.ReadClient()
		clock := int64(decoder.ReadLength())
		item = &contracts.StructID{Client: client, Clock: clock}
	}

	hasType := decoder.ReadLength() > 0
	if hasType {
		client := decoder.ReadClient()
		clock := int64(decoder.ReadLength())
		typeID = &contracts.StructID{Client: client, Clock: clock}
	} else {
		tname = decoder.ReadString()
	}

	assoc := decoder.ReadLength()

	return NewRelativePosition(typeID, tname, item, assoc)
}

// CreateRelativePositionFromTypeIndex creates a relative position from a type and index
func CreateRelativePositionFromTypeIndex(abstractType contracts.IAbstractType, index int, assoc int) *RelativePosition {
	var item *contracts.StructID
	var typeID *contracts.StructID
	var tname string

	if abstractType.GetItem() != nil {
		typeID = &abstractType.GetItem().GetID()
	} else {
		tname = abstractType.FindRootTypeKey()
	}

	if index == 0 {
		return NewRelativePosition(typeID, tname, nil, assoc)
	}

	// Find the item at the given index
	current := abstractType.GetStart()
	currentIndex := 0

	for current != nil && currentIndex < index {
		if !current.GetDeleted() {
			currentIndex += current.GetLength()
			if currentIndex >= index {
				// Found the item
				itemID := current.GetID()
				offset := currentIndex - index
				if offset > 0 {
					itemID.Clock -= int64(offset)
				}
				item = &itemID
				break
			}
		}
		current = current.GetRight()
	}

	return NewRelativePosition(typeID, tname, item, assoc)
}

// CreateAbsolutePositionFromRelativePosition creates an absolute position from a relative position
func CreateAbsolutePositionFromRelativePosition(rpos *RelativePosition, doc contracts.IYDoc) *AbsolutePosition {
	var abstractType contracts.IAbstractType

	if rpos.typeID != nil {
		// Find type by ID
		item := doc.GetStore().GetStruct(*rpos.typeID)
		if item != nil {
			if contentType, ok := item.GetContent().(interface {
				GetType() contracts.IAbstractType
			}); ok {
				abstractType = contentType.GetType()
			}
		}
	} else if rpos.tname != "" {
		// Find type by name
		abstractType = doc.Get(rpos.tname)
	}

	if abstractType == nil {
		return NewAbsolutePosition(nil, 0, rpos.assoc)
	}

	var index int
	if rpos.item == nil {
		index = 0
	} else {
		// Find the index of the item
		current := abstractType.GetStart()
		currentIndex := 0

		for current != nil {
			if current.GetID().Client == rpos.item.Client && current.GetID().Clock <= rpos.item.Clock {
				if current.GetID().Clock == rpos.item.Clock {
					// Found exact match
					index = currentIndex
					break
				} else if current.GetID().Clock+int64(current.GetLength()) > rpos.item.Clock {
					// Item is within this struct
					offset := rpos.item.Clock - current.GetID().Clock
					index = currentIndex + int(offset)
					break
				}
			}

			if !current.GetDeleted() {
				currentIndex += current.GetLength()
			}
			current = current.GetRight()
		}
	}

	return NewAbsolutePosition(abstractType, index, rpos.assoc)
}
