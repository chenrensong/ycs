package structs

// ContentEmbed represents embedded content
type ContentEmbed struct {
	Embed interface{}
}

// NewContentEmbed creates a new ContentEmbed
func NewContentEmbed(embed interface{}) *ContentEmbed {
	return &ContentEmbed{
		Embed: embed,
	}
}

// Ref returns the reference type for ContentEmbed
func (c *ContentEmbed) Ref() int {
	return 5 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentEmbed) Countable() bool {
	return true
}

// Length returns the length of this content
func (c *ContentEmbed) Length() int {
	return 1
}

// GetContent returns the content as a list of objects
func (c *ContentEmbed) GetContent() []interface{} {
	return []interface{}{c.Embed}
}

// Copy creates a copy of this content
func (c *ContentEmbed) Copy() Content {
	return NewContentEmbed(c.Embed)
}

// Splice splits this content at the specified offset
func (c *ContentEmbed) Splice(offset int) Content {
	// In the C# version, this throws NotImplementedException
	// We'll panic in Go to indicate this is not implemented
	panic("splice not implemented for ContentEmbed")
}

// MergeWith merges this content with the right content
func (c *ContentEmbed) MergeWith(right Content) bool {
	// In the C# version, this always returns false
	return false
}

// Integrate integrates this content
func (c *ContentEmbed) Integrate(transaction *Transaction, item *Item) {
	// Do nothing
}

// Delete deletes this content
func (c *ContentEmbed) Delete(transaction *Transaction) {
	// Do nothing
}

// Gc garbage collects this content
func (c *ContentEmbed) Gc(store *StructStore) {
	// Do nothing
}

// Write writes this content to an encoder
func (c *ContentEmbed) Write(encoder IUpdateEncoder, offset int) {
	encoder.WriteJson(c.Embed)
}

// Read reads ContentEmbed from a decoder
func ReadContentEmbed(decoder IUpdateDecoder) *ContentEmbed {
	content := decoder.ReadJson()
	return NewContentEmbed(content)
}