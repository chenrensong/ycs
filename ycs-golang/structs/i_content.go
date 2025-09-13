// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

import (
	"github.com/yjs/yjs-go/encoding"
	"github.com/yjs/yjs-go/structs"
)

// Content represents content that can be stored in a Yjs document.
type Content interface {
	// Countable returns whether this content is countable.
	Countable() bool

	// Length returns the length of this content.
	Length() int

	// GetContent returns the content as a slice of objects.
	GetContent() []interface{}

	// Copy creates a deep copy of this content.
	Copy() Content

	// Splice splits the content at the given offset.
	Splice(offset int) Content

	// MergeWith attempts to merge this content with another content.
	MergeWith(right Content) bool
}

// ContentEx extends Content with additional methods for internal use.
type ContentEx interface {
	Content

	// Ref returns the content reference identifier.
	Ref() int

	// Integrate integrates this content into a document.
	Integrate(transaction *structs.Transaction, item *structs.Item)

	// Delete marks this content as deleted.
	Delete(transaction *structs.Transaction)

	// GC performs garbage collection on this content.
	GC(store *structs.StructStore)

	// Write encodes this content to an encoder.
	Write(encoder encoding.Encoder, offset int) error
}
