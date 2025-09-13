// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

// ContentEmbed represents embedded content in the document
// This is the Go implementation of the C# ContentEmbed class
type ContentEmbed struct {
	Embed interface{}
}

// Ref is a constant reference ID for ContentEmbed
type _ref int

const (
	RefContentEmbed _ref = 5
)

// NewContentEmbed creates a new instance of ContentEmbed
func NewContentEmbed(embed interface{}) (*ContentEmbed, error) {
	return &ContentEmbed{
		Embed: embed,
	}, nil
}

// Countable returns true as ContentEmbed is countable
func (c *ContentEmbed) Countable() bool {
	return true
}

// Length returns the length of the content (always 1)
func (c *ContentEmbed) Length() int {
	return 1
}

// Ref returns the reference ID of the content
func (c *ContentEmbed) Ref() int {
	return int(RefContentEmbed)
}

// GetContent returns the content as a read-only list
func (c *ContentEmbed) GetContent() []interface{} {
	return []interface{}{c.Embed}
}

// Copy creates a copy of this content
func (c *ContentEmbed) Copy() IContent {
	return &ContentEmbed{
		Embed: c.Embed,
	}
}

// Splice splits the content at the given offset
// This operation is not supported for embedded content
func (c *ContentEmbed) Splice(offset int) IContent {
	// Embedded content cannot be spliced
	return nil // Or return self if needed
}

// MergeWith merges this content with another content
// Returns true if the merge was successful
func (c *ContentEmbed) MergeWith(right IContent) bool {
	// Embedded content cannot be merged
	return false
}

// Integrate integrates the embedded content with a transaction
func (c *ContentEmbed) Integrate(transaction *Transaction, item *Item) {
	// Do nothing (implementation as in C#)
}

// Delete deletes the embedded content
func (c *ContentEmbed) Delete(transaction *Transaction) {
	// Do nothing (implementation as in C#)
}

// Gc performs garbage collection on the embedded content
func (c *ContentEmbed) Gc(store *StructStore) {
	// Do nothing (implementation as in C#)
}

// Write writes the embedded content to an update encoder
func (c *ContentEmbed) Write(encoder IUpdateEncoder, offset int) {
	// Write the embedded content as JSON
	encoder.WriteJson(c.Embed)
}

// Read reads a ContentEmbed from the decoder
func (c *ContentEmbed) Read(decoder IUpdateDecoder) (*ContentEmbed, error) {
	// Read the embedded content from the decoder
	content, err := decoder.ReadJson()
	if err != nil {
		return nil, err
	}

	return &ContentEmbed{
		Embed: content,
	}, nil
}
