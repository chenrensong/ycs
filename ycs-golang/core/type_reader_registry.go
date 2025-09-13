// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"github.com/chenrensong/ygo/contracts"
)

// TypeReaderRegistry implements ITypeReaderRegistry
type TypeReaderRegistry struct {
	readers map[uint32]func(contracts.IUpdateDecoder) contracts.IAbstractType
}

// NewTypeReaderRegistry creates a new TypeReaderRegistry
func NewTypeReaderRegistry() contracts.ITypeReaderRegistry {
	return &TypeReaderRegistry{
		readers: make(map[uint32]func(contracts.IUpdateDecoder) contracts.IAbstractType),
	}
}

// RegisterTypeReader registers a type reader for a specific type
func (tr *TypeReaderRegistry) RegisterTypeReader(typeId uint32, reader func(contracts.IUpdateDecoder) contracts.IAbstractType) {
	tr.readers[typeId] = reader
}

// ReadType reads a type using the appropriate reader
func (tr *TypeReaderRegistry) ReadType(typeId uint32, decoder contracts.IUpdateDecoder) contracts.IAbstractType {
	if reader, exists := tr.readers[typeId]; exists {
		return reader(decoder)
	}
	return nil
}
