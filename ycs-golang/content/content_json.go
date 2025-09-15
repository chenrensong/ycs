package content

import (
	"ycs/contracts"
)

const ContentJsonRef = 2

// ContentJson represents JSON content
type ContentJson struct {
	content []interface{}
}

// NewContentJson creates a new ContentJson instance
func NewContentJson(content []interface{}) *ContentJson {
	contentCopy := make([]interface{}, len(content))
	copy(contentCopy, content)
	return &ContentJson{
		content: contentCopy,
	}
}

// GetRef returns the reference ID for this content type
func (c *ContentJson) GetRef() int {
	return ContentJsonRef
}

// GetCountable returns whether this content is countable
func (c *ContentJson) GetCountable() bool {
	return true
}

// GetLength returns the length of this content
func (c *ContentJson) GetLength() int {
	return len(c.content)
}

// GetContent returns the content as an interface slice
func (c *ContentJson) GetContent() []interface{} {
	result := make([]interface{}, len(c.content))
	copy(result, c.content)
	return result
}

// Copy creates a copy of this content
func (c *ContentJson) Copy() contracts.IContent {
	contentCopy := make([]interface{}, len(c.content))
	copy(contentCopy, c.content)
	return &ContentJson{content: contentCopy}
}

// Splice splits this content at the given offset
func (c *ContentJson) Splice(offset int) contracts.IContent {
	right := &ContentJson{
		content: make([]interface{}, len(c.content)-offset),
	}
	copy(right.content, c.content[offset:])
	c.content = c.content[:offset]
	return right
}

// MergeWith attempts to merge this content with another
func (c *ContentJson) MergeWith(right contracts.IContent) bool {
	rightJson, ok := right.(*ContentJson)
	if !ok {
		return false
	}
	c.content = append(c.content, rightJson.content...)
	return true
}

// Integrate integrates this content into a transaction
func (c *ContentJson) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Implementation would go here
}

// Delete deletes this content
func (c *ContentJson) Delete(transaction contracts.ITransaction) {
	// Implementation would go here
}

// Gc garbage collects this content
func (c *ContentJson) Gc(store contracts.IStructStore) {
	// Implementation would go here
}

// Write writes this content to an encoder
func (c *ContentJson) Write(encoder contracts.IUpdateEncoder, offset int) error {
	encoder.WriteJSON(c.content)
	return nil
}

// ReadContentJson reads ContentJson from a decoder
func ReadContentJson(decoder contracts.IUpdateDecoder) *ContentJson {
	length := decoder.ReadLength()
	content := make([]interface{}, length)

	for i := 0; i < length; i++ {
		content[i] = decoder.ReadAny()
	}

	return &ContentJson{content: content}
}
