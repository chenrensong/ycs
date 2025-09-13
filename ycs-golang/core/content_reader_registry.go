// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"github.com/chenrensong/ygo/contracts"
)

// ContentReaderRegistry implements IContentReaderRegistry
type ContentReaderRegistry struct {
	readers map[int]func(contracts.IUpdateDecoder) contracts.IContent
}

// NewContentReaderRegistry creates a new ContentReaderRegistry
func NewContentReaderRegistry() contracts.IContentReaderRegistry {
	return &ContentReaderRegistry{
		readers: make(map[int]func(contracts.IUpdateDecoder) contracts.IContent),
	}
}

// RegisterContentReader registers a content reader for a specific type
func (cr *ContentReaderRegistry) RegisterContentReader(contentType int, reader func(contracts.IUpdateDecoder) contracts.IContent) {
	cr.readers[contentType] = reader
}

// ReadContent reads content using the appropriate reader
func (cr *ContentReaderRegistry) ReadContent(contentType int, decoder contracts.IUpdateDecoder) contracts.IContent {
	if reader, exists := cr.readers[contentType]; exists {
		return reader(decoder)
	}
	return nil
}
