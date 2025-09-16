package content

import (
	"fmt"
	"ycs/contracts"
)

const ContentTypeRef = 7

// ContentType represents type content
type ContentType struct {
	contentType contracts.IAbstractType
}

// NewContentType creates a new ContentType instance
func NewContentType(contentType contracts.IAbstractType) *ContentType {
	return &ContentType{
		contentType: contentType,
	}
}

// GetRef returns the reference ID for this content type
func (c *ContentType) GetRef() int {
	return ContentTypeRef
}

// GetCountable returns whether this content is countable
func (c *ContentType) GetCountable() bool {
	return false
}

// GetLength returns the length of this content
func (c *ContentType) GetLength() int {
	return 1
}

// GetContent returns the content as an interface slice
func (c *ContentType) GetContent() []interface{} {
	return []interface{}{c.contentType}
}

// Copy creates a copy of this content
func (c *ContentType) Copy() contracts.IContent {
	// ContentType cannot be copied directly
	return nil
}

// Splice splits this content at the given offset
func (c *ContentType) Splice(offset int) contracts.IContent {
	// ContentType cannot be split
	return nil
}

// MergeWith attempts to merge this content with another
func (c *ContentType) MergeWith(right contracts.IContent) bool {
	// ContentType cannot be merged
	return false
}

// Integrate integrates this content into a transaction
func (c *ContentType) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Implementation would go here
}

// Delete deletes this content
func (c *ContentType) Delete(transaction contracts.ITransaction) {
	// Implementation would go here
}

// Gc garbage collects this content
func (c *ContentType) Gc(store contracts.IStructStore) {
	// Implementation would go here
}

// Write writes this content to an encoder
func (c *ContentType) Write(encoder contracts.IUpdateEncoder, offset int) error {
	// Implementation would go here
	return nil
}

// Type reader registry for avoiding circular dependencies
var typeReaderRegistry = make(map[uint32]func(contracts.IUpdateDecoder) contracts.IAbstractType)

// RegisterTypeReader registers a type reader for a given type reference ID
func RegisterTypeReader(key uint32, reader func(contracts.IUpdateDecoder) contracts.IAbstractType) {
	typeReaderRegistry[key] = reader
}

// ReadContentType reads ContentType from a decoder
func ReadContentType(decoder contracts.IUpdateDecoder) *ContentType {
	if typeReaderRegistry == nil {
		panic("TypeReaderRegistry not initialized. Call RegisterTypeReader() first.")
	}

	typeRef := decoder.ReadTypeRef()
	if reader, exists := typeReaderRegistry[typeRef]; exists {
		abstractType := reader(decoder)
		return NewContentType(abstractType)
	} else {
		panic(fmt.Sprintf("No type reader registered for typeRef %d", typeRef))
	}
}
