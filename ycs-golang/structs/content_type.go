package structs

// ContentType represents content with a type
type ContentType struct {
	Type AbstractType
}

// NewContentType creates a new ContentType
func NewContentType(typ AbstractType) *ContentType {
	return &ContentType{
		Type: typ,
	}
}

// Ref returns the reference type for ContentType
func (c *ContentType) Ref() int {
	return 7 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentType) Countable() bool {
	return true
}

// Length returns the length of this content
func (c *ContentType) Length() int {
	return 1
}

// GetContent returns the content as a list of objects
func (c *ContentType) GetContent() []interface{} {
	return []interface{}{c.Type}
}

// Copy creates a copy of this content
func (c *ContentType) Copy() Content {
	return NewContentType(c.Type.InternalCopy())
}

// Splice splits this content at the specified offset
func (c *ContentType) Splice(offset int) Content {
	// In the C# version, this throws NotImplementedException
	// We'll panic in Go to indicate this is not implemented
	panic("splice not implemented for ContentType")
}

// MergeWith merges this content with the right content
func (c *ContentType) MergeWith(right Content) bool {
	// In the C# version, this always returns false
	return false
}

// Integrate integrates this content
func (c *ContentType) Integrate(transaction *Transaction, item *Item) {
	c.Type.Integrate(transaction.Doc, item)
}

// Delete deletes this content
func (c *ContentType) Delete(transaction *Transaction) {
	// This is a simplified implementation
	// A full implementation would need to handle all the logic from the C# version
	
	item := c.Type.Start()
	for item != nil {
		// In a real implementation, you would need to check if the item is deleted
		// and handle accordingly
		
		// Move to the next item
		if item.Right != nil {
			// Type assert to Item to access Right
			if rightItem, ok := item.Right.(*Item); ok {
				item = rightItem
			} else {
				break
			}
		} else {
			break
		}
	}
	
	// In a real implementation, you would also need to handle the map values
	// and transaction changes
}

// Gc garbage collects this content
func (c *ContentType) Gc(store *StructStore) {
	// This is a simplified implementation
	// A full implementation would need to handle all the logic from the C# version
	
	// In a real implementation, you would need to garbage collect all items
	// and clear the start and map
}

// Write writes this content to an encoder
func (c *ContentType) Write(encoder IUpdateEncoder, offset int) {
	c.Type.Write(encoder)
}

// Read reads ContentType from a decoder
func ReadContentType(decoder IUpdateDecoder) *ContentType {
	typeRef := decoder.ReadTypeRef()
	
	// In a real implementation, you would need to handle the different type references
	// and create the appropriate type
	
	switch typeRef {
	case 0: // YArrayRefId placeholder
		// arr := ReadYArray(decoder)
		// return NewContentType(arr)
	case 1: // YMapRefId placeholder
		// map := ReadYMap(decoder)
		// return NewContentType(map)
	case 2: // YTextRefId placeholder
		// text := ReadYText(decoder)
		// return NewContentType(text)
	default:
		// In the C# version, this throws NotImplementedException
		// We'll panic in Go to indicate this is not implemented
		panic("type not implemented")
	}
	
	return nil
}