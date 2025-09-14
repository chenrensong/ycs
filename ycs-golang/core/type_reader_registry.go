package core

import (
	"errors"
	"fmt"
	"ycs/contracts"
)

// TypeReaderRegistry is the default implementation of type reader registry
type TypeReaderRegistry struct {
	readers map[uint32]func(contracts.IUpdateDecoder) contracts.IAbstractType
}

// NewTypeReaderRegistry creates a new TypeReaderRegistry
func NewTypeReaderRegistry() *TypeReaderRegistry {
	return &TypeReaderRegistry{
		readers: make(map[uint32]func(contracts.IUpdateDecoder) contracts.IAbstractType),
	}
}

// RegisterTypeReader registers a type reader for a given type reference ID
func (tr *TypeReaderRegistry) RegisterTypeReader(typeRefID uint32, reader func(contracts.IUpdateDecoder) contracts.IAbstractType) {
	tr.readers[typeRefID] = reader
}

// ReadType reads a type using the registered reader
func (tr *TypeReaderRegistry) ReadType(typeRefID uint32, decoder contracts.IUpdateDecoder) (contracts.IAbstractType, error) {
	if reader, exists := tr.readers[typeRefID]; exists {
		return reader(decoder), nil
	}

	return nil, errors.New(fmt.Sprintf("Type %d not implemented", typeRefID))
}
