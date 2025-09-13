// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

// ContentFormat represents a content format in the document
// This is the Go implementation of the C# ContentFormat class
type ContentFormat struct {
	key   string
	value interface{}
}

// Ref is a constant reference ID for ContentFormat
type _ref int

const (
	RefContentFormat _ref = 6
)

// NewContentFormat creates a new instance of ContentFormat
func NewContentFormat(key string, value interface{}) (*ContentFormat, error) {
	return &ContentFormat{
		key:   key,
		value: value,
	}, nil
}

// Countable returns false as ContentFormat is not countable
func (c *ContentFormat) Countable() bool {
	return false
}

// Length returns the length of the content (always 1)
func (c *ContentFormat) Length() int {
	return 1
}

// Ref returns the reference ID of the content
func (c *ContentFormat) Ref() int {
	return int(RefContentFormat)
}

// GetContent returns the content as a read-only list
// This operation is not supported for format content
func (c *ContentFormat) GetContent() []interface{} {
	// Format content has no actual content
	return nil
}

// Copy creates a copy of this content
func (c *ContentFormat) Copy() IContent {
	return &ContentFormat{
		key:   c.key,
		value: c.value,
	}
}

// Splice splits the content at the given offset
// This operation is not supported for format content
func (c *ContentFormat) Splice(offset int) IContent {
	// Format content cannot be spliced
	return nil
}

// MergeWith merges this content with another content
// Returns true if the merge was successful
func (c *ContentFormat) MergeWith(right IContent) bool {
	// Format content cannot be merged
	return false
}

// Integrate integrates the format content with a transaction
func (c *ContentFormat) Integrate(transaction *Transaction, item *Item) {
	// Search markers are currently unsupported for rich text documents
	// Clear search markers from parent YArrayBase
	if yArrayBase, ok := item.Parent().(*YArrayBase); ok {
		yArrayBase.ClearSearchMarkers()
	}
}

// Delete deletes the format content
func (c *ContentFormat) Delete(transaction *Transaction) {
	// Do nothing (implementation as in C#)
}

// Gc performs garbage collection on the format content
func (c *ContentFormat) Gc(store *StructStore) {
	// Do nothing (implementation as in C#)
}

// Write writes the format content to an update encoder
func (c *ContentFormat) Write(encoder IUpdateEncoder, offset int) {
	// Write the key and JSON value to the encoder
	encoder.WriteKey(c.key)
	encoder.WriteJson(c.value)
}

// Read reads a ContentFormat from the decoder
func (c *ContentFormat) Read(decoder IUpdateDecoder) (*ContentFormat, error) {
	// Read the key and JSON value from the decoder
	key := decoder.ReadKey()
	value, err := decoder.ReadJson()
	if err != nil {
		return nil, err
	}

	return &ContentFormat{
		key:   key,
		value: value,
	}, nil
}
