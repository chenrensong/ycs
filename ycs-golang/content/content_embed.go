package content

import (
	"ycs/contracts"
)

const ContentEmbedRef = 3

// ContentEmbed represents embedded content
type ContentEmbed struct {
	embed interface{}
}

// NewContentEmbed creates a new ContentEmbed instance
func NewContentEmbed(embed interface{}) *ContentEmbed {
	return &ContentEmbed{
		embed: embed,
	}
}

// GetRef returns the reference ID for this content type
func (c *ContentEmbed) GetRef() int {
	return ContentEmbedRef
}

// GetCountable returns whether this content is countable
func (c *ContentEmbed) GetCountable() bool {
	return true
}

// GetLength returns the length of this content
func (c *ContentEmbed) GetLength() int {
	return 1
}

// GetContent returns the content as an interface slice
func (c *ContentEmbed) GetContent() []interface{} {
	return []interface{}{c.embed}
}

// Copy creates a copy of this content
func (c *ContentEmbed) Copy() contracts.IContent {
	return &ContentEmbed{
		embed: c.embed,
	}
}

// Splice splits this content at the given offset
func (c *ContentEmbed) Splice(offset int) contracts.IContent {
	// ContentEmbed cannot be split
	return nil
}

// MergeWith attempts to merge this content with another
func (c *ContentEmbed) MergeWith(right contracts.IContent) bool {
	// ContentEmbed cannot be merged
	return false
}

// Integrate integrates this content into a transaction
func (c *ContentEmbed) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Implementation would go here
}

// Delete deletes this content
func (c *ContentEmbed) Delete(transaction contracts.ITransaction) {
	// Implementation would go here
}

// Gc garbage collects this content
func (c *ContentEmbed) Gc(store contracts.IStructStore) {
	// Implementation would go here
}

// Write writes this content to an encoder
func (c *ContentEmbed) Write(encoder contracts.IUpdateEncoder, offset int) error {
	encoder.WriteEmbed(c.embed)
	return nil
}

// ReadContentEmbed reads ContentEmbed from a decoder
func ReadContentEmbed(decoder contracts.IUpdateDecoder) *ContentEmbed {
	embed := decoder.ReadEmbed()
	return &ContentEmbed{embed: embed}
}
