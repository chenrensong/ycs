package structs

// ContentAny represents any type of content
type ContentAny struct {
	content []interface{}
}

// NewContentAny creates a new ContentAny from a slice of interface{}
func NewContentAny(content []interface{}) *ContentAny {
	return &ContentAny{
		content: content,
	}
}

// Ref returns the reference type for ContentAny
func (c *ContentAny) Ref() int {
	return 8 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentAny) Countable() bool {
	return true
}

// Length returns the length of this content
func (c *ContentAny) Length() int {
	return len(c.content)
}

// GetContent returns the content as a list of objects
func (c *ContentAny) GetContent() []interface{} {
	// Create a copy of the content slice
	contentCopy := make([]interface{}, len(c.content))
	copy(contentCopy, c.content)
	return contentCopy
}

// Copy creates a copy of this content
func (c *ContentAny) Copy() Content {
	contentCopy := make([]interface{}, len(c.content))
	copy(contentCopy, c.content)
	return NewContentAny(contentCopy)
}

// Splice splits this content at the specified offset
func (c *ContentAny) Splice(offset int) Content {
	rightContent := make([]interface{}, len(c.content)-offset)
	copy(rightContent, c.content[offset:])
	
	right := NewContentAny(rightContent)
	
	// Remove the content from the original
	c.content = c.content[:offset]
	
	return right
}

// MergeWith merges this content with the right content
func (c *ContentAny) MergeWith(right Content) bool {
	// In Go, we use type assertion to check the type
	if rightAny, ok := right.(*ContentAny); ok {
		c.content = append(c.content, rightAny.content...)
		return true
	}
	return false
}

// Integrate integrates this content
func (c *ContentAny) Integrate(transaction *Transaction, item *Item) {
	// Do nothing
}

// Delete deletes this content
func (c *ContentAny) Delete(transaction *Transaction) {
	// Do nothing
}

// Gc garbage collects this content
func (c *ContentAny) Gc(store *StructStore) {
	// Do nothing
}

// Write writes this content to an encoder
func (c *ContentAny) Write(encoder IUpdateEncoder, offset int) {
	length := len(c.content)
	encoder.WriteLength(length - offset)

	for i := offset; i < length; i++ {
		c := c.content[i]
		encoder.WriteAny(c)
	}
}

// Read reads ContentAny from a decoder
func ReadContentAny(decoder IUpdateDecoder) *ContentAny {
	length := decoder.ReadLength()
	cs := make([]interface{}, length)

	for i := 0; i < length; i++ {
		c := decoder.ReadAny()
		cs[i] = c
	}

	return NewContentAny(cs)
}