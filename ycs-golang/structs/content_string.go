// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

import (
	"errors"
	"strings"
)

// ContentString represents a string content in the document
// This is the Go implementation of the C# ContentString class
type ContentString struct {
	content string
}

// Ref is a constant reference ID for ContentString
type _ref int

const (
	RefContentString _ref = 4
)

// NewContentString creates a new instance of ContentString
// value is the string content
func NewContentString(value string) (*ContentString, error) {
	return &ContentString{
		content: value,
	}, nil
}

// Countable returns true as ContentString is countable
func (c *ContentString) Countable() bool {
	return true
}

// Length returns the length of the string content
func (c *ContentString) Length() int {
	return len(c.content)
}

// GetContent returns the string content as a read-only string
func (c *ContentString) GetContent() string {
	return c.content
}

// Copy creates a copy of this string content
func (c *ContentString) Copy() IContent {
	return &ContentString{
		content: c.content,
	}, nil
}

// Splice splits the string content at the given offset
// Returns the right part of the split
func (c *ContentString) Splice(offset int) IContent {
	if offset < 0 || offset > len(c.content) {
		return nil // Or return an error
	}

	// Create a new content with the right part
	rightContent := c.content[offset:]

	// Update this content to the left part
	c.content = c.content[:offset]

	// Check for invalid surrogate pairs
	if len(c.content) > 0 && offset > 0 {
		firstCharCode := c.content[offset-1]
		if (firstCharCode >= 0xD800 && firstCharCode <= 0xDBFF) {
		// Last character of the left split is the start of a surrogate pair
		// We don't support splitting of surrogate pairs
		// Replace the invalid character with a unicode replacement character
		c.content = c.content[:offset-1] + "�" + c.content[offset-1:]
		rightContent = "�" + rightContent
	}

	return &ContentString{
		content: rightContent,
	}, nil
}

// MergeWith merges this string content with another content
// Returns true if the merge was successful
func (c *ContentString) MergeWith(right IContent) bool {
	// Assert that right is ContentString
	rightString, ok := right.(*ContentString)
	if !ok {
		return false
	}

	// Append the right content to this content
	c.content += rightString.content
	return true
}

// Integrate integrates the string content with a transaction
func (c *ContentString) Integrate(transaction *Transaction, item *Item) {
	// Do nothing (implementation as in C#)
}

// Delete deletes the string content
func (c *ContentString) Delete(transaction *Transaction) {
	// Do nothing (implementation as in C#)
}

// Gc performs garbage collection on the string content
func (c *ContentString) Gc(store *StructStore) {
	// Do nothing (implementation as in C#)
}

// Write writes the string content to an update encoder
func (c *ContentString) Write(encoder IUpdateEncoder, offset int) {
	// Write the string content
	encoder.WriteString(c.content)
}

// Read reads a ContentString from the decoder
func (c *ContentString) Read(decoder IUpdateDecoder) (*ContentString, error) {
	// Read the string content
	str, err := decoder.ReadString()
	if err != nil {
		return nil, err
	}

	return &ContentString{
		content: str,
	}, nil
}