package content

import (
	"ycs/contracts"
)

const ContentStringRef = 4

// ContentString represents string content
type ContentString struct {
	content string
}

// NewContentString creates a new ContentString instance
func NewContentString(content string) *ContentString {
	return &ContentString{
		content: content,
	}
}

// GetRef returns the reference ID for this content type
func (c *ContentString) GetRef() int {
	return ContentStringRef
}

// GetCountable returns whether this content is countable
func (c *ContentString) GetCountable() bool {
	return true
}

// GetLength returns the length of this content
func (c *ContentString) GetLength() int {
	return len(c.content)
}

// GetContent returns the content as an interface slice
func (c *ContentString) GetContent() []interface{} {
	runes := []rune(c.content)
	result := make([]interface{}, len(runes))
	for i, r := range runes {
		result[i] = string(r)
	}
	return result
}

// Copy creates a copy of this content
func (c *ContentString) Copy() contracts.IContent {
	return &ContentString{
		content: c.content,
	}
}

// Splice splits this content at the given offset
func (c *ContentString) Splice(offset int) contracts.IContent {
	runes := []rune(c.content)
	right := &ContentString{
		content: string(runes[offset:]),
	}
	c.content = string(runes[:offset])
	return right
}

// MergeWith attempts to merge this content with another
func (c *ContentString) MergeWith(right contracts.IContent) bool {
	rightString, ok := right.(*ContentString)
	if !ok {
		return false
	}
	c.content += rightString.content
	return true
}

// Integrate integrates this content into a transaction
func (c *ContentString) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Implementation would go here
}

// Delete deletes this content
func (c *ContentString) Delete(transaction contracts.ITransaction) {
	// Implementation would go here
}

// Gc garbage collects this content
func (c *ContentString) Gc(store contracts.IStructStore) {
	// Implementation would go here
}

// Write writes this content to an encoder
func (c *ContentString) Write(encoder contracts.IUpdateEncoder, offset int) error {
	runes := []rune(c.content)
	encoder.WriteString(string(runes[offset:]))
	return nil
}

// ReadContentString reads ContentString from a decoder
func ReadContentString(decoder contracts.IUpdateDecoder) *ContentString {
	str := decoder.ReadString()
	return &ContentString{content: str}
}
