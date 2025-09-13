// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

import (
	"encoding/json"
)

// ContentJson represents JSON content in the document
// This is the Go implementation of the C# ContentJson class
type ContentJson struct {
	content []interface{}
}

// Ref is a constant reference ID for ContentJson
type _ref int

const (
	RefContentJson _ref = 2
)

// NewContentJson creates a new instance of ContentJson
// data is the JSON content
func NewContentJson(data []interface{}) (*ContentJson, error) {
	return &ContentJson{
		content: data,
	}, nil
}

// NewContentJsonFromInterfaceSlice creates a new instance of ContentJson from a slice of interfaces
func NewContentJsonFromInterfaceSlice(data []interface{}) (*ContentJson, error) {
	return NewContentJson(data)
}

// Ref returns the reference ID of the content
func (c *ContentJson) Ref() int {
	return int(RefContentJson)
}

// Countable returns true as ContentJson is countable
func (c *ContentJson) Countable() bool {
	return true
}

// Length returns the length of the JSON content
func (c *ContentJson) Length() int {
	if c.content == nil {
		return 0
	}
	return len(c.content)
}

// GetContent returns the JSON content as a read-only list
func (c *ContentJson) GetContent() []interface{} {
	if c.content == nil {
		return nil
	}
	return c.content
}

// Copy creates a copy of this JSON content
func (c *ContentJson) Copy() IContent {
	newContent := make([]interface{}, len(c.content))
	copy(newContent, c.content)
	return &ContentJson{
		content: newContent,
	}
}

// Splice splits the JSON content at the given offset
// Returns the right part of the split
func (c *ContentJson) Splice(offset int) IContent {
	if offset < 0 || offset >= len(c.content) {
		return nil // Or return an error
	}

	// Create a new content with the right part
	rightContent := c.content[offset:]

	// Remove the right part from this content
	c.content = c.content[:offset]

	return &ContentJson{
		content: rightContent,
	}
}

// MergeWith merges this JSON content with another content
// Returns true if the merge was successful
func (c *ContentJson) MergeWith(right IContent) bool {
	// Assert that right is ContentJson
	rightJson, ok := right.(*ContentJson)
	if !ok {
		return false
	}

	// Add the right content to this content
	c.content = append(c.content, rightJson.content...)
	return true
}

// Integrate integrates the JSON content with a transaction
func (c *ContentJson) Integrate(transaction *Transaction, item *Item) {
	// Do nothing (implementation as in C#)
}

// Delete deletes the JSON content
func (c *ContentJson) Delete(transaction *Transaction) {
	// Do nothing (implementation as in C#)
}

// Gc performs garbage collection on the JSON content
func (c *ContentJson) Gc(store *StructStore) {
	// Do nothing (implementation as in C#)
}

// Write writes the JSON content to an update encoder
func (c *ContentJson) Write(encoder IUpdateEncoder, offset int) {
	length := len(c.content)
	encoder.WriteLength(length)

	for i := offset; i < length; i++ {
		// Serialize the JSON content to a string
		jsonStr, err := json.Marshal(c.content[i])
		if err != nil {
			// Handle the error
			return // Or panic if needed
		}
		encoder.WriteString(string(jsonStr))
	}
}

// Read reads a ContentJson from the decoder
func (c *ContentJson) Read(decoder IUpdateDecoder) (*ContentJson, error) {
	// Read the length
	length, err := decoder.ReadLength()
	if err != nil {
		return nil, err
	}

	content := make([]interface{}, length)

	for i := 0; i < length; i++ {
		// Read the JSON string
		jsonStr, err := decoder.ReadString()
		if err != nil {
			return nil, err
		}

		// Handle undefined values
		if jsonStr == "undefined" {
			content[i] = nil
		} else {
			// Deserialize the JSON string
			var jsonObj interface{}
			err := json.Unmarshal([]byte(jsonStr), &jsonObj)
			if err != nil {
				return nil, err
			}
			content[i] = jsonObj
		}
	}

	return &ContentJson{
		content: content,
	}, nil
}
