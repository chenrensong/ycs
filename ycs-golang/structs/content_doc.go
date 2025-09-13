// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

import (
	"errors"
)

// ContentDoc represents a document content type
// This is the Go implementation of the C# ContentDoc class
type ContentDoc struct {
	doc  *YDoc
	opts *YDocOptions
}

// Ref is a constant reference ID for ContentDoc
type _ref int

const (
	RefContentDoc _ref = 9
)

// NewContentDoc creates a new instance of ContentDoc
func NewContentDoc(doc *YDoc) (*ContentDoc, error) {
	if doc.GetItem() != nil {
		return nil, errors.New("this document was already integrated as a sub-document. You should create a second instance instead with the same guid")
	}

	// Create a new instance with the same GUID
	opts := NewYDocOptions()
	if !doc.Gc() {
		opts.SetGc(false)
	}
	if doc.AutoLoad() {
		opts.SetAutoLoad(true)
	}
	if doc.Meta() != nil {
		opts.SetMeta(doc.Meta())
	}

	return &ContentDoc{
		doc:  doc,
		opts: opts,
	}, nil
}

// Ref returns the reference ID of the content
func (c *ContentDoc) Ref() int {
	return int(RefContentDoc)
}

// Countable returns true as ContentDoc is countable
func (c *ContentDoc) Countable() bool {
	return true
}

// Length returns the length of the content (always 1)
func (c *ContentDoc) Length() int {
	return 1
}

// GetContent returns the content as a read-only list
func (c *ContentDoc) GetContent() []interface{} {
	return []interface{}{c.doc}
}

// Copy creates a copy of this content
func (c *ContentDoc) Copy() IContent {
	// Create a new ContentDoc with the same document
	copyDoc, _ := NewContentDoc(c.doc)
	return copyDoc
}

// Splice splits the content at the given offset
// This operation is not supported for ContentDoc
func (c *ContentDoc) Splice(offset int) IContent {
	// Document content cannot be spliced
	return nil // Or return self if needed
}

// MergeWith merges this content with another content
// Returns true if the merge was successful
func (c *ContentDoc) MergeWith(right IContent) bool {
	// Document content cannot be merged
	return false
}

// Integrate integrates the content with a transaction
func (c *ContentDoc) Integrate(transaction *Transaction, item *Item) {
	// Set the item reference in the document
	c.doc.SetItem(item)
	// Add to transaction's subdocs added
	transaction.SubdocsAdded.Add(c.doc)

	// If document should load, add to loaded subdocs
	if c.doc.ShouldLoad() {
		transaction.SubdocsLoaded.Add(c.doc)
	}
}

// Delete deletes the content
func (c *ContentDoc) Delete(transaction *Transaction) {
	// Handle subdocument removal
	if transaction.SubdocsAdded.Contains(c.doc) {
		transaction.SubdocsAdded.Remove(c.doc)
	} else {
		transaction.SubdocsRemoved.Add(c.doc)
	}
}

// Gc performs garbage collection on the content
func (c *ContentDoc) Gc(store *StructStore) {
	// Do nothing (implementation as in C#)
}

// Write writes the document content to an update encoder
func (c *ContentDoc) Write(encoder IUpdateEncoder, offset int) {
	// Write the document GUID
	encoder.WriteString(c.doc.Guid())

	// Write the document options
	opts, _ := c.opts.Write(encoder, offset)
}

// Read reads a ContentDoc from the decoder
func (c *ContentDoc) Read(decoder IUpdateDecoder) (*ContentDoc, error) {
	// Read the document GUID
	guidStr, err := decoder.ReadString()
	if err != nil {
		return nil, err
	}

	// Read document options
	opts, err := YDocOptionsRead(decoder)
	if err != nil {
		return nil, err
	}

	// Set GUID in options
	opts.SetGuid(guidStr)

	// Create new document
	doc := NewYDoc(opts)

	return &ContentDoc{
		doc:  doc,
		opts: opts,
	}, nil
}
