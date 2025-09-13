package utils

// RelativePosition represents a relative position in a document
type RelativePosition struct {
	Type   string
	Item   *ID
	Offset int
}

// NewRelativePosition creates a new relative position
func NewRelativePosition(type_ string, item *ID, offset int) *RelativePosition {
	return &RelativePosition{
		Type:   type_,
		Item:   item,
		Offset: offset,
	}
}

// CreateRelativePositionFromAbsolutePosition creates a relative position from an absolute position
func CreateRelativePositionFromAbsolutePosition(absPos *AbsolutePosition) *RelativePosition {
	if absPos == nil || absPos.Type == nil {
		return nil
	}

	// Get the type key
	typeKey := ""
	if absPos.Type.GetDoc() != nil {
		typeKey = absPos.Type.GetDoc().FindRootTypeKey(absPos.Type)
	}

	// Find the item at the given index
	var itemId *ID
	offset := absPos.Index

	// This is a simplified implementation
	// In a real implementation, we would traverse the linked list to find the exact item

	return NewRelativePosition(typeKey, itemId, offset)
}

// Equal checks if two relative positions are equal
func (rp *RelativePosition) Equal(other *RelativePosition) bool {
	if other == nil {
		return false
	}

	return rp.Type == other.Type &&
		rp.Item.Equals(other.Item) &&
		rp.Offset == other.Offset
}

// Compare compares two relative positions
func (rp *RelativePosition) Compare(other *RelativePosition) int {
	if other == nil {
		return 1
	}

	// Compare by type first
	if rp.Type < other.Type {
		return -1
	} else if rp.Type > other.Type {
		return 1
	}

	// Compare by item
	if rp.Item != nil && other.Item != nil {
		itemCmp := rp.Item.Compare(other.Item)
		if itemCmp != 0 {
			return itemCmp
		}
	} else if rp.Item == nil && other.Item != nil {
		return -1
	} else if rp.Item != nil && other.Item == nil {
		return 1
	}

	// Compare by offset
	if rp.Offset < other.Offset {
		return -1
	} else if rp.Offset > other.Offset {
		return 1
	}

	return 0
}
