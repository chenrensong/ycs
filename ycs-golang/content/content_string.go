package content

import (
	"strings"

	"github.com/chenrensong/ygo/contracts"
)

const ContentStringRef = 4

type ContentString struct {
	content []rune
}

func NewContentString(value string) *ContentString {
	return &ContentString{
		content: []rune(value),
	}
}

func NewContentStringFromRunes(content []rune) *ContentString {
	runes := make([]rune, len(content))
	copy(runes, content)
	return &ContentString{
		content: runes,
	}
}

func (c *ContentString) GetRef() int {
	return ContentStringRef
}

func (c *ContentString) GetCountable() bool {
	return true
}

func (c *ContentString) GetLength() int {
	return len(c.content)
}

func (c *ContentString) AppendToBuilder(sb *strings.Builder) {
	for _, r := range c.content {
		sb.WriteRune(r)
	}
}

func (c *ContentString) GetString() string {
	return string(c.content)
}

func (c *ContentString) GetContent() []interface{} {
	result := make([]interface{}, len(c.content))
	for i, r := range c.content {
		result[i] = r
	}
	return result
}

func (c *ContentString) Copy() contracts.IContent {
	runes := make([]rune, len(c.content))
	copy(runes, c.content)
	return &ContentString{content: runes}
}

func (c *ContentString) Splice(offset int) contracts.IContent {
	right := &ContentString{
		content: make([]rune, len(c.content)-offset),
	}
	copy(right.content, c.content[offset:])
	c.content = c.content[:offset]

	// Prevent encoding invalid documents because of splitting of surrogate pairs.
	if offset > 0 {
		firstCharCode := c.content[offset-1]
		if firstCharCode >= 0xD800 && firstCharCode <= 0xDBFF {
			// Last character of the left split is the start of a surrogate utf16/ucs2 pair.
			// We don't support splitting of surrogate pairs because this may lead to invalid documents.
			// Replace the invalid character with a unicode replacement character U+FFFD.
			c.content[offset-1] = '\uFFFD'

			// Replace right as well.
			right.content[0] = '\uFFFD'
		}
	}

	return right
}

func (c *ContentString) MergeWith(right contracts.IContent) bool {
	rightString, ok := right.(*ContentString)
	if !ok {
		return false
	}
	c.content = append(c.content, rightString.content...)
	return true
}

func (c *ContentString) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Do nothing
}

func (c *ContentString) Delete(transaction contracts.ITransaction) {
	// Do nothing
}

func (c *ContentString) Gc(store contracts.IStructStore) {
	// Do nothing
}

func (c *ContentString) Write(encoder contracts.IUpdateEncoder, offset int) {
	str := string(c.content[offset:])
	encoder.WriteString(str)
}

func ReadContentString(decoder contracts.IUpdateDecoder) *ContentString {
	str := decoder.ReadString()
	return &ContentString{content: []rune(str)}
}
