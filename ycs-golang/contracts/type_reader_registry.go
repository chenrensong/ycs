package contracts

// ITypeReaderRegistry represents a registry for type readers to avoid circular dependencies
type ITypeReaderRegistry interface {
	// RegisterTypeReader registers a type reader for a specific type reference ID
	RegisterTypeReader(typeRefID uint32, reader func(IUpdateDecoder) IAbstractType)
	// ReadType reads AbstractType by type reference ID
	ReadType(typeRefID uint32, decoder IUpdateDecoder) IAbstractType
}
