package structs

// ContentBinary represents binary content
type ContentBinary struct {
	content []byte
}

// NewContentBinary creates a new ContentBinary from a byte slice
func NewContentBinary(data []byte) *ContentBinary {
	return &ContentBinary{
		content: data,
	}
}

// Ref returns the reference type for ContentBinary
func (c *ContentBinary) Ref() int {
	return 3 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentBinary) Countable() bool {
	return true
}

// Length returns the length of this content
func (c *ContentBinary) Length() int {
	return 1
}

// GetContent returns the content as a list of objects
func (c *ContentBinary) GetContent() []interface{} {
	return []interface{}{c.content}
}

// Copy creates a copy of this content
func (c *ContentBinary) Copy() Content {
	contentCopy := make([]byte, len(c.content))
	copy(contentCopy, c.content)
	return NewContentBinary(contentCopy)
}

// Splice splits this content at the specified offset
func (c *ContentBinary) Splice(offset int) Content {
	// In the C# version, this throws NotImplementedException
	// We'll panic in Go to indicate this is not implemented
	panic("splice not implemented for ContentBinary")
}

// MergeWith merges this content with the right content
func (c *ContentBinary) MergeWith(right Content) bool {
	// In the C# version, this always returns false
	return false
}

// Integrate integrates this content
func (c *ContentBinary) Integrate(transaction *Transaction, item *Item) {
	// Do nothing
}

// Delete deletes this content
func (c *ContentBinary) Delete(transaction *Transaction) {
	// Do nothing
}

// Gc garbage collects this content
func (c *ContentBinary) Gc(store *StructStore) {
	// Do nothing
}

// Write writes this content to an encoder
func (c *ContentBinary) Write(encoder IUpdateEncoder, offset int) {
	encoder.WriteBuffer(c.content)
}

// Read reads ContentBinary from a decoder
func ReadContentBinary(decoder IUpdateDecoder) *ContentBinary {
	content := decoder.ReadBuffer()
	return NewContentBinary(content)
}