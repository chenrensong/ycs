// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"github.com/chenrensong/ygo/contracts"
)

// AbsolutePosition represents an absolute position that can be transformed back to a relative position
type AbsolutePosition struct {
	abstractType contracts.IAbstractType
	index        int
	assoc        int
}

// NewAbsolutePosition creates a new AbsolutePosition
func NewAbsolutePosition(abstractType contracts.IAbstractType, index int, assoc int) *AbsolutePosition {
	return &AbsolutePosition{
		abstractType: abstractType,
		index:        index,
		assoc:        assoc,
	}
}

// GetType returns the abstract type
func (ap *AbsolutePosition) GetType() contracts.IAbstractType {
	return ap.abstractType
}

// GetIndex returns the index
func (ap *AbsolutePosition) GetIndex() int {
	return ap.index
}

// GetAssoc returns the association
func (ap *AbsolutePosition) GetAssoc() int {
	return ap.assoc
}

// ToRelativePosition converts this absolute position to a relative position
func (ap *AbsolutePosition) ToRelativePosition() *RelativePosition {
	if ap.abstractType == nil {
		return NewRelativePosition(nil, "", nil, ap.assoc)
	}

	return CreateRelativePositionFromTypeIndex(ap.abstractType, ap.index, ap.assoc)
}

// CreateAbsolutePositionFromTypeIndex creates an absolute position from a type and index
func CreateAbsolutePositionFromTypeIndex(abstractType contracts.IAbstractType, index int, assoc int) *AbsolutePosition {
	return NewAbsolutePosition(abstractType, index, assoc)
}

// CompareAbsolutePositions compares two absolute positions
func CompareAbsolutePositions(a, b *AbsolutePosition) int {
	if a.abstractType != b.abstractType {
		// Different types, cannot compare directly
		return 0
	}

	if a.index < b.index {
		return -1
	} else if a.index > b.index {
		return 1
	}

	// Same index, compare association
	if a.assoc < b.assoc {
		return -1
	} else if a.assoc > b.assoc {
		return 1
	}

	return 0
}
