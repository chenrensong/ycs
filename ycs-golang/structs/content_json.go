package structs

import (
	"encoding/json"
)

// ContentJson represents JSON content
type ContentJson struct {
	content []interface{}
}

// NewContentJson creates a new ContentJson from a slice of interface{}
func NewContentJson(data []interface{}) *ContentJson {
	contentCopy := make([]interface{}, len(data))
	copy(contentCopy, data)
	return &ContentJson{
		content: contentCopy,
	}
}

// NewContentJsonFromContent creates a new ContentJson from another ContentJson's content
func NewContentJsonFromContent(other []interface{}) *ContentJson {
	return &ContentJson{
		content: other,
	}
}

// Ref returns the reference type for ContentJson
func (c *ContentJson) Ref() int {
	return 2 // _ref constant from C# version
}

// Countable returns whether this content is countable
func (c *ContentJson) Countable() bool {
	return true
}

// Length returns the length of this content
func (c *ContentJson) Length() int {
	return len(c.content)
}

// GetContent returns the content as a list of objects
func (c *ContentJson) GetContent() []interface{} {
	contentCopy := make([]interface{}, len(c.content))
	copy(contentCopy, c.content)
	return contentCopy
}

// Copy creates a copy of this content
func (c *ContentJson) Copy() Content {
	contentCopy := make([]interface{}, len(c.content))
	copy(contentCopy, c.content)
	return NewContentJsonFromContent(contentCopy)
}

// Splice splits this content at the specified offset
func (c *ContentJson) Splice(offset int) Content {
	rightContent := make([]interface{}, len(c.content)-offset)
	copy(rightContent, c.content[offset:])
	
	right := NewContentJsonFromContent(rightContent)
	
	// Remove the content from the original
	c.content = c.content[:offset]
	
	return right
}

// MergeWith merges this content with the right content
func (c *ContentJson) MergeWith(right Content) bool {
	// In Go, we use type assertion to check the type
	if rightJson, ok := right.(*ContentJson); ok {
		c.content = append(c.content, rightJson.content...)
		return true
	}
	return false
}

// Integrate integrates this content
func (c *ContentJson) Integrate(transaction *Transaction, item *Item) {
	// Do nothing
}

// Delete deletes this content
func (c *ContentJson) Delete(transaction *Transaction) {
	// Do nothing
}

// Gc garbage collects this content
func (c *ContentJson) Gc(store *StructStore) {
	// Do nothing
}

// Write writes this content to an encoder
func (c *ContentJson) Write(encoder IUpdateEncoder, offset int) {
	length := len(c.content)
	encoder.WriteLength(length)
	
	for i := offset; i < length; i++ {
		// Serialize the object to JSON
		jsonBytes, err := json.Marshal(c.content[i])
		if err != nil {
			// If marshaling fails, write "undefined" as in the C# version
			encoder.WriteString("undefined")
			continue
		}
		
		jsonStr := string(jsonBytes)
		encoder.WriteString(jsonStr)
	}
}

// Read reads ContentJson from a decoder
func ReadContentJson(decoder IUpdateDecoder) *ContentJson {
	length := decoder.ReadLength()
	content := make([]interface{}, length)

	for i := 0; i < length; i++ {
		jsonStr := decoder.ReadString()
		
		// Check if the string is "undefined"
		if jsonStr == "undefined" {
			content[i] = nil
			continue
		}
		
		// Deserialize the JSON string
		var jsonObj interface{}
		err := json.Unmarshal([]byte(jsonStr), &jsonObj)
		if err != nil {
			// If unmarshaling fails, store nil
			content[i] = nil
			continue
		}
		
		content[i] = jsonObj
	}

	return NewContentJsonFromContent(content)
}