package structs

import (
	"strings"
)

// ContentString represents string content
type ContentString struct {
	content []rune
}

// NewContentString creates a new ContentString from a string
func NewContentString(value string) *ContentString {
	runes := []rune(value)
	return &ContentString{
		content: runes,
	}
}

// NewContentStringFromRunes creates a new ContentString from a slice of runes
func NewContentStringFromRunes(content []rune) *ContentString {
	return &ContentString{
		content: content,
	}
}

// Ref returns the reference type for ContentString
func (c *ContentString) Ref() int {
	return 4 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentString) Countable() bool {
	return true
}

// Length returns the length of this content
func (c *ContentString) Length() int {
	return len(c.content)
}

// AppendToBuilder appends the content to a strings.Builder
func (c *ContentString) AppendToBuilder(sb *strings.Builder) {
	for _, r := range c.content {
		sb.WriteRune(r)
	}
}

// GetString returns the content as a string
func (c *ContentString) GetString() string {
	return string(c.content)
}

// GetContent returns the content as a list of objects
func (c *ContentString) GetContent() []interface{} {
	result := make([]interface{}, len(c.content))
	for i, r := range c.content {
		result[i] = r
	}
	return result
}

// Copy creates a copy of this content
func (c *ContentString) Copy() Content {
	contentCopy := make([]rune, len(c.content))
	copy(contentCopy, c.content)
	return NewContentStringFromRunes(contentCopy)
}

// Splice splits this content at the specified offset
func (c *ContentString) Splice(offset int) Content {
	rightContent := make([]rune, len(c.content)-offset)
	copy(rightContent, c.content[offset:])
	
	right := NewContentStringFromRunes(rightContent)
	
	// Remove the content from the original
	c.content = c.content[:offset]
	
	// Prevent encoding invalid documents because of splitting of surrogate pairs.
	if offset > 0 {
		lastChar := c.content[offset-1]
		if lastChar >= 0xD800 && lastChar <= 0xDBFF {
			// Last character of the left split is the start of a surrogate utf16/ucs2 pair.
			// We don't support splitting of surrogate pairs because this may lead to invalid documents.
			// Replace the invalid character with a unicode replacement character U+FFFD.
			c.content[offset-1] = '\uFFFD'
			
			// Replace right as well.
			if len(right.content) > 0 {
				right.content[0] = '\uFFFD'
			}
		}
	}
	
	return right
}

// MergeWith merges this content with the right content
func (c *ContentString) MergeWith(right Content) bool {
	// In Go, we use type assertion to check the type
	if rightString, ok := right.(*ContentString); ok {
		c.content = append(c.content, rightString.content...)
		return true
	}
	return false
}

// Integrate integrates this content
func (c *ContentString) Integrate(transaction *Transaction, item *Item) {
	// Do nothing
}

// Delete deletes this content
func (c *ContentString) Delete(transaction *Transaction) {
	// Do nothing
}

// Gc garbage collects this content
func (c *ContentString) Gc(store *StructStore) {
	// Do nothing
}

// Write writes this content to an encoder
func (c *ContentString) Write(encoder IUpdateEncoder, offset int) {
	str := string(c.content[offset:])
	encoder.WriteString(str)
}

// Read reads ContentString from a decoder
func ReadContentString(decoder IUpdateDecoder) *ContentString {
	str := decoder.ReadString()
	return NewContentString(str)
}