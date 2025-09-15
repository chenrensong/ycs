package content

import (
	"ycs/contracts"
)

const ContentFormatRef = 6

// ContentFormat represents formatted content
type ContentFormat struct {
	key   string
	value interface{}
}

// NewContentFormat creates a new ContentFormat instance
func NewContentFormat(key string, value interface{}) *ContentFormat {
	return &ContentFormat{
		key:   key,
		value: value,
	}
}

// GetRef returns the reference ID for this content type
func (c *ContentFormat) GetRef() int {
	return ContentFormatRef
}

// GetCountable returns whether this content is countable
func (c *ContentFormat) GetCountable() bool {
	return false
}

// GetLength returns the length of this content
func (c *ContentFormat) GetLength() int {
	return 1
}

// GetContent returns the content as an interface slice
func (c *ContentFormat) GetContent() []interface{} {
	return []interface{}{nil}
}

// Copy creates a copy of this content
func (c *ContentFormat) Copy() contracts.IContent {
	return &ContentFormat{
		key:   c.key,
		value: c.value,
	}
}

// Splice splits this content at the given offset
func (c *ContentFormat) Splice(offset int) contracts.IContent {
	// ContentFormat cannot be split
	return nil
}

// MergeWith attempts to merge this content with another
func (c *ContentFormat) MergeWith(right contracts.IContent) bool {
	// ContentFormat cannot be merged
	return false
}

// Integrate integrates this content into a transaction
func (c *ContentFormat) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Implementation would go here
}

// Delete deletes this content
func (c *ContentFormat) Delete(transaction contracts.ITransaction) {
	// Implementation would go here
}

// Gc garbage collects this content
func (c *ContentFormat) Gc(store contracts.IStructStore) {
	// Implementation would go here
}

// Write writes this content to an encoder
func (c *ContentFormat) Write(encoder contracts.IUpdateEncoder, offset int) error {
	encoder.WriteKey(c.key)
	encoder.WriteAny(c.value)
	return nil
}

// ReadContentFormat reads ContentFormat from a decoder
func ReadContentFormat(decoder contracts.IUpdateDecoder) *ContentFormat {
	key := decoder.ReadKey()
	value := decoder.ReadAny()
	return &ContentFormat{key: key, value: value}
}
