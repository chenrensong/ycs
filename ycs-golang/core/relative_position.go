package core

import (
	"ycs/contracts"
)

// RelativePosition represents a relative position in the document
// A relative position is based on the Y.js model and is not affected by document changes.
// E.g. if you place a relative position before a certain character, it will always point to this character.
// If you place a relative position at the end of a type, it will always point to the end of the type.
type RelativePosition struct {
	Item   *contracts.StructID
	TypeId *contracts.StructID
	TName  string
	// A relative position is associated to a specific character.
	// By default, the value is >= 0, the relative position is associated to the character
	// after the meant position.
	// If the value is < 0, then the relative position is associated with the character
	// before the meant position.
	Assoc int
}

// NewRelativePosition creates a new RelativePosition
func NewRelativePosition(typ contracts.IAbstractType, item *contracts.StructID, assoc int) *RelativePosition {
	rpos := &RelativePosition{
		Item:  item,
		Assoc: assoc,
	}

	if typ.GetItem() == nil {
		rpos.TName = typ.FindRootTypeKey()
	} else {
		typeId := contracts.StructID{
			Client: typ.GetItem().GetID().Client,
			Clock:  typ.GetItem().GetID().Clock,
		}
		rpos.TypeId = &typeId
	}

	return rpos
}

// NewRelativePositionFromComponents creates a new RelativePosition from components
func NewRelativePositionFromComponents(typeId *contracts.StructID, tname string, item *contracts.StructID, assoc int) *RelativePosition {
	return &RelativePosition{
		TypeId: typeId,
		TName:  tname,
		Item:   item,
		Assoc:  assoc,
	}
}

// Equals compares two RelativePosition instances
func (rp *RelativePosition) Equals(other *RelativePosition) bool {
	if rp == other {
		return true
	}

	if other == nil {
		return false
	}

	return rp.TName == other.TName &&
		contracts.EqualsPtr(rp.Item, other.Item) &&
		contracts.EqualsPtr(rp.TypeId, other.TypeId) &&
		rp.Assoc == other.Assoc
}
