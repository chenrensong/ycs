// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

// ContentBinary represents binary content in the document
// This is the Go implementation of the C# ContentBinary class
type ContentBinary struct {
	content []byte
}

// Ref is a constant reference ID for ContentBinary
type _ref int

const (
	RefContentBinary _ref = 3
)

// NewContentBinary creates a new instance of ContentBinary
func NewContentBinary(data []byte) (*ContentBinary, error) {
	return &ContentBinary{
		content: data,
	}, nil
}

// Countable returns true as ContentBinary is countable
func (c *ContentBinary) Countable() bool {
	return true
}

// Length returns the length of the content (always 1)
func (c *ContentBinary) Length() int {
	return 1
}

// Ref returns the reference ID of the content
func (c *ContentBinary) Ref() int {
	return int(RefContentBinary)
}

// GetContent returns the content as a read-only list
func (c *ContentBinary) GetContent() []interface{} {
	return []interface{}{c.content}
}

// Copy creates a copy of this content
func (c *ContentBinary) Copy() IContent {
	return &ContentBinary{
		content: c.content,
	}
}

// Splice splits the content at the given offset
// This operation is not supported for ContentBinary
func (c *ContentBinary) Splice(offset int) IContent {
	// Binary content cannot be spliced
	return nil // Or return self if needed
}

// MergeWith merges this content with another content
// Returns true if the merge was successful
func (c *ContentBinary) MergeWith(right IContent) bool {
	// Binary content cannot be merged
	return false
}

// Integrate integrates the content with a transaction
func (c *ContentBinary) Integrate(transaction *Transaction, item *Item) {
	// Do nothing (implementation as in C#)
}

// Delete deletes the content
func (c *ContentBinary) Delete(transaction *Transaction) {
	// Do nothing (implementation as in C#)
}

// Gc performs garbage collection on the content
func (c *ContentBinary) Gc(store *StructStore) {
	// Do nothing (implementation as in C#)
}

// Write writes the binary content to an update encoder
func (c *ContentBinary) Write(encoder IUpdateEncoder, offset int) {
	// Write the binary buffer to the encoder
	encoder.WriteBuffer(c.content)
}

// Read reads a ContentBinary from the decoder
func (c *ContentBinary) Read(decoder IUpdateDecoder) (*ContentBinary, error) {
	// Read the binary buffer from the decoder
	content, err := decoder.ReadBuffer()
	if err != nil {
		return nil, err
	}

	return &ContentBinary{
		content: content,
	}, nil
}
