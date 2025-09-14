package core

import (
	"ycs/contracts"
)

// StructID represents a struct ID in core package
type StructID struct {
	Client int64
	Clock  int64
}

// NewStructID creates a new StructID
func NewStructID(client, clock int64) StructID {
	return StructID{
		Client: client,
		Clock:  clock,
	}
}

// ToContractsStructID converts to contracts.StructID
func (s StructID) ToContractsStructID() contracts.StructID {
	return contracts.StructID{
		Client: s.Client,
		Clock:  s.Clock,
	}
}

// FromContractsStructID converts from contracts.StructID
func FromContractsStructID(s contracts.StructID) StructID {
	return StructID{
		Client: s.Client,
		Clock:  s.Clock,
	}
}

// ToPointer returns a pointer to StructID
func (s StructID) ToPointer() *StructID {
	return &s
}
