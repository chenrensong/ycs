// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

// ContentDeleted represents deleted content in the document
// This is the Go implementation of the C# ContentDeleted class
type ContentDeleted struct {
	length int
}

// Ref is a constant reference ID for ContentDeleted
type _ref int

const (
	RefContentDeleted _ref = 1
)

// NewContentDeleted creates a new instance of ContentDeleted
func NewContentDeleted(length int) (*ContentDeleted, error) {
	return &ContentDeleted{
		length: length,
	}, nil
}

// Countable returns false as ContentDeleted is not countable
func (c *ContentDeleted) Countable() bool {
	return false
}

// Length returns the length of the deleted content
func (c *ContentDeleted) Length() int {
	return c.length
}

// Ref returns the reference ID of the content
func (c *ContentDeleted) Ref() int {
	return int(RefContentDeleted)
}

// GetContent returns the content as a read-only list
// This operation is not supported for deleted content
func (c *ContentDeleted) GetContent() []interface{} {
	// Deleted content has no actual content
	return nil
}

// Copy creates a copy of this deleted content
func (c *ContentDeleted) Copy() IContent {
	return &ContentDeleted{
		length: c.length,
	}
}

// Splice splits the deleted content at the given offset
func (c *ContentDeleted) Splice(offset int) IContent {
	right, _ := NewContentDeleted(c.length - offset)
	c.length = offset
	return right
}

// MergeWith merges this deleted content with another content
// Returns true if the merge was successful
func (c *ContentDeleted) MergeWith(right IContent) bool {
	// We expect to merge with another ContentDeleted
	rightContentDeleted, ok := right.(*ContentDeleted)
	if !ok {
		return false
	}

	// Merge the lengths
	c.length += rightContentDeleted.length
	return true
}

// Integrate integrates the deleted content with a transaction
func (c *ContentDeleted) Integrate(transaction *Transaction, item *Item) {
	// Add to delete set and mark item as deleted
	transaction.DeleteSet.Add(item.Id.Client, item.Id.Clock, c.length)
	item.MarkDeleted()
}

// Delete deletes the content
func (c *ContentDeleted) Delete(transaction *Transaction) {
	// Do nothing (implementation as in C#)
}

// Gc performs garbage collection on the deleted content
func (c *ContentDeleted) Gc(store *StructStore) {
	// Do nothing (implementation as in C#)
}

// Write writes the deleted content to an update encoder
func (c *ContentDeleted) Write(encoder IUpdateEncoder, offset int) {
	// Write the length to the encoder
	encoder.WriteLength(c.length - offset)
}

// Read reads a ContentDeleted from the decoder
func (c *ContentDeleted) Read(decoder IUpdateDecoder) (*ContentDeleted, error) {
	// Read the length from the decoder
	length, err := decoder.ReadLength()
	if err != nil {
		return nil, err
	}

	return &ContentDeleted{
		length: length,
	}, nil
}
