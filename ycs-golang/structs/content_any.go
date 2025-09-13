// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

// ContentAny represents any type of content in the document
// This is the Go implementation of the C# ContentAny class
type ContentAny struct {
	content []interface{}
}

// Ref is a constant reference ID for ContentAny
type _ref int

const (
	RefContentAny _ref = 8
)

// NewContentAny creates a new instance of ContentAny
func NewContentAny(content []interface{}) (*ContentAny, error) {
	// Create a copy of the content to avoid reference issues
	copyContent := make([]interface{}, len(content))
	copy(copyContent, content)

	return &ContentAny{
		content: copyContent,
	}, nil
}

// Countable returns true as ContentAny is countable
func (c *ContentAny) Countable() bool {
	return true
}

// Length returns the length of the content
func (c *ContentAny) Length() int {
	return len(c.content)
}

// Ref returns the reference ID of the content
func (c *ContentAny) Ref() int {
	return int(RefContentAny)
}

// GetContent returns the content as a read-only list
func (c *ContentAny) GetContent() []interface{} {
	// Return a copy to maintain read-only semantics
	copyContent := make([]interface{}, len(c.content))
	copy(copyContent, c.content)
	return copyContent
}

// Copy creates a copy of this content
func (c *ContentAny) Copy() IContent {
	copyContent, _ := NewContentAny(c.content)
	return copyContent
}

// Splice splits the content at the given offset
func (c *ContentAny) Splice(offset int) IContent {
	// Create a new content with the right part
	rightContent := c.content[offset:]
	right, _ := NewContentAny(rightContent)

	// Remove the right part from this content
	c.content = c.content[:offset]

	return right
}

// MergeWith merges this content with another content
// Returns true if the merge was successful
func (c *ContentAny) MergeWith(right IContent) bool {
	// We expect to merge with another ContentAny
	rightContentAny, ok := right.(*ContentAny)
	if !ok {
		return false
	}

	// Append the content from right
	c.content = append(c.content, rightContentAny.content...)
	return true
}

// Integrate integrates the content with a transaction
func (c *ContentAny) Integrate(transaction *Transaction, item *Item) {
	// Do nothing (implementation as in C#)
}

// Delete deletes the content
func (c *ContentAny) Delete(transaction *Transaction) {
	// Do nothing (implementation as in C#)
}

// Gc performs garbage collection on the content
func (c *ContentAny) Gc(store *StructStore) {
	// Do nothing (implementation as in C#)
}

// Write writes the content to an update encoder
func (c *ContentAny) Write(encoder IUpdateEncoder, offset int) {
	length := len(c.content)
	// Write the length to the encoder
	encoder.WriteLength(length - offset)

	// Write each item in the content
	for i := offset; i < length; i++ {
		encoder.WriteAny(c.content[i])
	}
}

// Read reads a ContentAny from the decoder
func (c *ContentAny) Read(decoder IUpdateDecoder) (*ContentAny, error) {
	length, err := decoder.ReadLength()
	if err != nil {
		return nil, err
	}

	cs := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		item, err := decoder.ReadAny()
		if err != nil {
			return nil, err
		}
		cs = append(cs, item)
	}

	return &ContentAny{
		content: cs,
	}, nil
}
