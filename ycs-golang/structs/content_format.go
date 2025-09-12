package structs

// ContentFormat represents format content
type ContentFormat struct {
	Key   string
	Value interface{}
}

// NewContentFormat creates a new ContentFormat
func NewContentFormat(key string, value interface{}) *ContentFormat {
	return &ContentFormat{
		Key:   key,
		Value: value,
	}
}

// Ref returns the reference type for ContentFormat
func (c *ContentFormat) Ref() int {
	return 6 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentFormat) Countable() bool {
	return false
}

// Length returns the length of this content
func (c *ContentFormat) Length() int {
	return 1
}

// GetContent returns the content as a list of objects
func (c *ContentFormat) GetContent() []interface{} {
	// In the C# version, this throws NotImplementedException
	// We'll return nil in Go to indicate this is not implemented
	return nil
}

// Copy creates a copy of this content
func (c *ContentFormat) Copy() Content {
	return NewContentFormat(c.Key, c.Value)
}

// Splice splits this content at the specified offset
func (c *ContentFormat) Splice(offset int) Content {
	// In the C# version, this throws NotImplementedException
	// We'll panic in Go to indicate this is not implemented
	panic("splice not implemented for ContentFormat")
}

// MergeWith merges this content with the right content
func (c *ContentFormat) MergeWith(right Content) bool {
	// In the C# version, this always returns false
	return false
}

// Integrate integrates this content
func (c *ContentFormat) Integrate(transaction *Transaction, item *Item) {
	// Search markers are currently unsupported for rich text documents.
	// In Go, we would need to check if the parent is a YArrayBase and clear search markers
	// (item.Parent as YArrayBase)?.ClearSearchMarkers()
}

// Delete deletes this content
func (c *ContentFormat) Delete(transaction *Transaction) {
	// Do nothing
}

// Gc garbage collects this content
func (c *ContentFormat) Gc(store *StructStore) {
	// Do nothing
}

// Write writes this content to an encoder
func (c *ContentFormat) Write(encoder IUpdateEncoder, offset int) {
	encoder.WriteKey(c.Key)
	encoder.WriteJson(c.Value)
}

// Read reads ContentFormat from a decoder
func ReadContentFormat(decoder IUpdateDecoder) *ContentFormat {
	key := decoder.ReadKey()
	value := decoder.ReadJson()
	return NewContentFormat(key, value)
}