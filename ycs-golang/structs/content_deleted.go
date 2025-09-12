package structs

// ContentDeleted represents deleted content
type ContentDeleted struct {
	Length int
}

// NewContentDeleted creates a new ContentDeleted
func NewContentDeleted(length int) *ContentDeleted {
	return &ContentDeleted{
		Length: length,
	}
}

// Ref returns the reference type for ContentDeleted
func (c *ContentDeleted) Ref() int {
	return 1 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentDeleted) Countable() bool {
	return false
}

// Length returns the length of this content
func (c *ContentDeleted) Length() int {
	return c.Length
}

// GetContent returns the content as a list of objects
func (c *ContentDeleted) GetContent() []interface{} {
	// In Go, we return nil to represent the NotImplementedException
	return nil
}

// Copy creates a copy of this content
func (c *ContentDeleted) Copy() Content {
	return NewContentDeleted(c.Length)
}

// Splice splits this content at the specified offset
func (c *ContentDeleted) Splice(offset int) Content {
	right := NewContentDeleted(c.Length - offset)
	c.Length = offset
	return right
}

// MergeWith merges this content with the right content
func (c *ContentDeleted) MergeWith(right Content) bool {
	// In Go, we use type assertion to check the type
	if _, ok := right.(*ContentDeleted); ok {
		c.Length += right.Length()
		return true
	}
	return false
}

// Integrate integrates this content
func (c *ContentDeleted) Integrate(transaction *Transaction, item *Item) {
	// In a real implementation, you would need to add to the delete set
	// transaction.DeleteSet.Add(item.Id.Client, item.Id.Clock, c.Length)
	item.MarkDeleted()
}

// Delete deletes this content
func (c *ContentDeleted) Delete(transaction *Transaction) {
	// Do nothing
}

// Gc garbage collects this content
func (c *ContentDeleted) Gc(store *StructStore) {
	// Do nothing
}

// Write writes this content to an encoder
func (c *ContentDeleted) Write(encoder IUpdateEncoder, offset int) {
	encoder.WriteLength(c.Length - offset)
}

// Read reads ContentDeleted from a decoder
func ReadContentDeleted(decoder IUpdateDecoder) *ContentDeleted {
	length := decoder.ReadLength()
	return NewContentDeleted(length)
}