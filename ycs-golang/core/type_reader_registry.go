package core

import (
	"fmt"
	"ycs/contracts"
)

// TypeReaderRegistry is the default implementation of type reader registry
type TypeReaderRegistry struct {
	readers map[uint32]func(contracts.IUpdateDecoder) contracts.IAbstractType
}

var globalTypeReaderRegistry contracts.ITypeReaderRegistry

// NewTypeReaderRegistry creates a new TypeReaderRegistry
func NewTypeReaderRegistry() *TypeReaderRegistry {
	return &TypeReaderRegistry{
		readers: make(map[uint32]func(contracts.IUpdateDecoder) contracts.IAbstractType),
	}
}

// SetGlobalTypeReaderRegistry sets the global type reader registry
func SetGlobalTypeReaderRegistry(registry contracts.ITypeReaderRegistry) {
	globalTypeReaderRegistry = registry
}

// GetGlobalTypeReaderRegistry gets the global type reader registry
func GetGlobalTypeReaderRegistry() contracts.ITypeReaderRegistry {
	return globalTypeReaderRegistry
}

// RegisterTypeReader registers a type reader for a given type reference ID
func (tr *TypeReaderRegistry) RegisterTypeReader(typeRefID uint32, reader func(contracts.IUpdateDecoder) contracts.IAbstractType) {
	tr.readers[typeRefID] = reader
}

// ReadType reads a type using the registered reader
func (tr *TypeReaderRegistry) ReadType(typeRefID uint32, decoder contracts.IUpdateDecoder) contracts.IAbstractType {
	if reader, exists := tr.readers[typeRefID]; exists {
		return reader(decoder)
	}

	panic(fmt.Sprintf("Type %d not implemented", typeRefID))
}
