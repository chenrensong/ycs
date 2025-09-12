package structs

// ContentDoc represents document content
type ContentDoc struct {
	Doc  *YDoc
	Opts *YDocOptions
}

// NewContentDoc creates a new ContentDoc
func NewContentDoc(doc *YDoc) *ContentDoc {
	// In the C# version, there's a check to ensure the document hasn't been integrated yet
	// We'll skip that for now in this Go implementation
	
	opts := NewYDocOptions()
	
	// Copy options from the document
	if !doc.Gc {
		opts.Gc = false
	}
	
	if doc.AutoLoad {
		opts.AutoLoad = true
	}
	
	if doc.Meta != nil {
		opts.Meta = doc.Meta
	}
	
	return &ContentDoc{
		Doc:  doc,
		Opts: opts,
	}
}

// Ref returns the reference type for ContentDoc
func (c *ContentDoc) Ref() int {
	return 9 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentDoc) Countable() bool {
	return true
}

// Length returns the length of this content
func (c *ContentDoc) Length() int {
	return 1
}

// GetContent returns the content as a list of objects
func (c *ContentDoc) GetContent() []interface{} {
	return []interface{}{c.Doc}
}

// Copy creates a copy of this content
func (c *ContentDoc) Copy() Content {
	return NewContentDoc(c.Doc)
}

// Splice splits this content at the specified offset
func (c *ContentDoc) Splice(offset int) Content {
	// In the C# version, this throws NotImplementedException
	// We'll panic in Go to indicate this is not implemented
	panic("splice not implemented for ContentDoc")
}

// MergeWith merges this content with the right content
func (c *ContentDoc) MergeWith(right Content) bool {
	// In the C# version, this always returns false
	return false
}

// Integrate integrates this content
func (c *ContentDoc) Integrate(transaction *Transaction, item *Item) {
	// This needs to be reflected in doc.destroy as well.
	c.Doc.Item = item
	transaction.SubdocsAdded = append(transaction.SubdocsAdded, c.Doc)
	
	if c.Doc.ShouldLoad {
		transaction.SubdocsLoaded = append(transaction.SubdocsLoaded, c.Doc)
	}
}

// Delete deletes this content
func (c *ContentDoc) Delete(transaction *Transaction) {
	// In a real implementation, you would need to check if the document is in SubdocsAdded
	// and remove it or add it to SubdocsRemoved
	
	// This is a simplified implementation
	found := false
	for i, doc := range transaction.SubdocsAdded {
		if doc == c.Doc {
			// Remove from SubdocsAdded
			transaction.SubdocsAdded = append(transaction.SubdocsAdded[:i], transaction.SubdocsAdded[i+1:]...)
			found = true
			break
		}
	}
	
	if !found {
		// Add to SubdocsRemoved
		transaction.SubdocsRemoved = append(transaction.SubdocsRemoved, c.Doc)
	}
}

// Gc garbage collects this content
func (c *ContentDoc) Gc(store *StructStore) {
	// Do nothing
}

// Write writes this content to an encoder
func (c *ContentDoc) Write(encoder IUpdateEncoder, offset int) {
	// 32 digits separated by hyphens, no braces.
	encoder.WriteString(c.Doc.Guid)
	c.Opts.Write(encoder, offset)
}

// Read reads ContentDoc from a decoder
func ReadContentDoc(decoder IUpdateDecoder) *ContentDoc {
	guidStr := decoder.ReadString()
	
	opts := ReadYDocOptions(decoder)
	opts.Guid = guidStr
	
	// In a real implementation, you would need to create a new YDoc with the options
	// doc := NewYDoc(opts)
	// return NewContentDoc(doc)
	
	return nil // Placeholder
}

// Placeholder types that would need to be implemented elsewhere
type YDoc struct {
	Guid      string
	Gc        bool
	AutoLoad  bool
	ShouldLoad bool
	Meta      interface{}
	Item      *Item
}

type YDocOptions struct {
	Guid     string
	Gc       bool
	AutoLoad bool
	Meta     interface{}
}

func NewYDocOptions() *YDocOptions {
	return &YDocOptions{
		Gc:       true,
		AutoLoad: false,
	}
}

func ReadYDocOptions(decoder IUpdateDecoder) *YDocOptions {
	// Placeholder implementation
	return NewYDocOptions()
}

func (o *YDocOptions) Write(encoder IUpdateEncoder, offset int) {
	// Placeholder implementation
}