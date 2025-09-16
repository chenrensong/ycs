package content

import (
	"ycs/contracts"
)

const ContentDocRef = 9

// ContentDoc represents document content
type ContentDoc struct {
	opts *contracts.YDocOptions
}

// NewContentDoc creates a new ContentDoc instance
func NewContentDoc(opts *contracts.YDocOptions) *ContentDoc {
	return &ContentDoc{
		opts: opts,
	}
}

// GetRef returns the reference ID for this content type
func (c *ContentDoc) GetRef() int {
	return ContentDocRef
}

// GetCountable returns whether this content is countable
func (c *ContentDoc) GetCountable() bool {
	return false
}

// GetLength returns the length of this content
func (c *ContentDoc) GetLength() int {
	return 1
}

// GetContent returns the content as an interface slice
func (c *ContentDoc) GetContent() []interface{} {
	return []interface{}{nil}
}

// Copy creates a copy of this content
func (c *ContentDoc) Copy() contracts.IContent {
	return &ContentDoc{
		opts: c.opts,
	}
}

// Splice splits this content at the given offset
func (c *ContentDoc) Splice(offset int) contracts.IContent {
	// ContentDoc cannot be split
	return nil
}

// MergeWith attempts to merge this content with another
func (c *ContentDoc) MergeWith(right contracts.IContent) bool {
	// ContentDoc cannot be merged
	return false
}

// Integrate integrates this content into a transaction
func (c *ContentDoc) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Implementation would go here
}

// Delete deletes this content
func (c *ContentDoc) Delete(transaction contracts.ITransaction) {
	// Implementation would go here
}

// Gc garbage collects this content
func (c *ContentDoc) Gc(store contracts.IStructStore) {
	// Implementation would go here
}

// Write writes this content to an encoder
func (c *ContentDoc) Write(encoder contracts.IUpdateEncoder, offset int) error {
	// Implementation would go here
	return nil
}

// Document factory for avoiding circular dependencies
var docFactory func(*contracts.YDocOptions) contracts.IYDoc

// SetDocFactory sets the document factory function
func SetDocFactory(factory func(*contracts.YDocOptions) contracts.IYDoc) {
	docFactory = factory
}

// ReadContentDoc reads ContentDoc from a decoder
func ReadContentDoc(decoder contracts.IUpdateDecoder) *ContentDoc {
	// Implementation would go here
	return nil
}
