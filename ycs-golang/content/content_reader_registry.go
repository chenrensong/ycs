package content

import (
	"fmt"
	"ycs/contracts"
)

// ContentReaderRegistry represents the default implementation of content reader registry
type ContentReaderRegistry struct {
	readers map[int]func(contracts.IUpdateDecoder) contracts.IContent
}

// NewContentReaderRegistry creates a new ContentReaderRegistry
func NewContentReaderRegistry() *ContentReaderRegistry {
	return &ContentReaderRegistry{
		readers: make(map[int]func(contracts.IUpdateDecoder) contracts.IContent),
	}
}

// RegisterContentReader registers a content reader for a specific content reference
func (crr *ContentReaderRegistry) RegisterContentReader(contentRef int, reader func(contracts.IUpdateDecoder) contracts.IContent) {
	crr.readers[contentRef] = reader
}

// ReadContent reads content using the registered reader for the given content reference
func (crr *ContentReaderRegistry) ReadContent(contentRef int, decoder contracts.IUpdateDecoder) contracts.IContent {
	if reader, exists := crr.readers[contentRef]; exists {
		return reader(decoder)
	}

	panic(fmt.Sprintf("Content type not recognized: %d", contentRef))
}
