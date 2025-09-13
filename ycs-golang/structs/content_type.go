// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

import (
	"errors"
)

// ContentType represents a content type in the document
// This is the Go implementation of the C# ContentType class
type ContentType struct {
	typeRef int
	typ     AbstractType
}

// Ref is a constant reference ID for ContentType
type _ref int

const (
	RefContentType _ref = 7
)

// NewContentType creates a new instance of ContentType
// type is the abstract type to wrap
func NewContentType(typ AbstractType) (*ContentType, error) {
	return &ContentType{
		typeRef: RefContentType,
		typ:     typ,
	}, nil
}

// Countable returns true as ContentType is countable
func (c *ContentType) Countable() bool {
	return true
}

// Length returns the length of the content (always 1)
func (c *ContentType) Length() int {
	return 1
}

// GetType returns the abstract type
func (c *ContentType) GetType() AbstractType {
	return c.typ
}

// GetContent returns the content as a list
func (c *ContentType) GetContent() []interface{} {
	return []interface{}{c.typ}
}

// Copy creates a copy of this content
func (c *ContentType) Copy() IContent {
	// Create a copy of the underlying type
	copyType := c.typ.InternalCopy()
	return &ContentType{
		typ: copyType,
	}
}

// Splice splits the content at the given offset
// This operation is not supported for ContentType
func (c *ContentType) Splice(offset int) IContent {
	// Content type cannot be spliced
	return nil // Or return an error
}

// MergeWith merges this content with another content
// Returns true if the merge was successful
func (c *ContentType) MergeWith(right IContent) bool {
	// Content type cannot be merged
	return false
}

// Integrate integrates the content with a transaction
func (c *ContentType) Integrate(transaction *Transaction, item *Item) {
	// Integrate the type with the transaction
	c.typ.Integrate(transaction, item)
}

// Delete deletes the content
func (c *ContentType) Delete(transaction *Transaction) {
	item := c.typ.GetStart()

	for item != nil {
		if !item.Deleted() {
			item.Delete(transaction)
		} else {
			// Add to mergeStructs as it needs to be merged
			transaction.MergeStructs = append(transaction.MergeStructs, item)
		}

		// Move to the next item
		item = item.GetRight().(*Item)
	}

	// Process map values
	for _, valueItem := range c.typ.GetMap() {
		if !valueItem.Deleted() {
			valueItem.Delete(transaction)
		} else {
			// Add to mergeStructs
			transaction.MergeStructs = append(transaction.MergeStructs, valueItem)
		}
	}

	// Remove from changed
	transaction.Changed.Remove(c.typ)
}

// Gc performs garbage collection on the content
func (c *ContentType) Gc(store *StructStore) {
	item := c.typ.GetStart()

	for item != nil {
		item.Gc(store, true)
		// Move to the next item
		item = item.GetRight().(*Item)
	}

	// Clear the start reference
	c.typ.SetStart(nil)

	// Process map values
	for _, valueItem := range c.typ.GetMap() {
		for valueItem != nil {
			valueItem.Gc(store, true)
			// Move to the left item
			valueItem = valueItem.GetLeft().(*Item)
		}
	}

	// Clear the map
	c.typ.GetMap().Clear()
}

// Write writes the content to an update encoder
func (c *ContentType) Write(encoder IUpdateEncoder, offset int) {
	// Write the type to the encoder
	c.typ.Write(encoder)
}

// Read reads a ContentType from the decoder
func (c *ContentType) Read(decoder IUpdateDecoder) (*ContentType, error) {
	// Read the type reference
	typeRef, err := decoder.ReadTypeRef()
	if err != nil {
		return nil, err
	}

	// Handle based on type reference
	switch typeRef {
	case YArrayRefId: // YArray type reference
		arr, err := YArrayRead(decoder)
		if err != nil {
			return nil, err
		}
		return NewContentType(arr)
	case YMapRefId: // YMap type reference
		m, err := YMapRead(decoder)
		if err != nil {
			return nil, err
		}
		return NewContentType(m)
	case YTextRefId: // YText type reference
		text, err := YTextRead(decoder)
		if err != nil {
			return nil, err
		}
		return NewContentType(text)
	default:
		return nil, errors.New("type " + string(typeRef) + " not implemented")
	}
}

// Type references for different content types
const (
	YArrayRefId = 0 // Reference ID for YArray
	YMapRefId   = 1 // Reference ID for YMap
	YTextRefId  = 2 // Reference ID for YText
)

// YArrayRead is a placeholder for the actual YArray.Read implementation
// This should be replaced with the real implementation
func YArrayRead(decoder IUpdateDecoder) (*YArray, error) {
	return nil, errors.New("YArrayRead not implemented")
}

// YMapRead is a placeholder for the actual YMap.Read implementation
// This should be replaced with the real implementation
func YMapRead(decoder IUpdateDecoder) (*YMap, error) {
	return nil, errors.New("YMapRead not implemented")
}

// YTextRead is a placeholder for the actual YText.Read implementation
// This should be replaced with the real implementation
func YTextRead(decoder IUpdateDecoder) (*YText, error) {
	return nil, errors.New("YTextRead not implemented")
}
