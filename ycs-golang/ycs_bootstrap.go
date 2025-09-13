// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package ycs

import (
	"sync"

	"github.com/chenrensong/ygo/content"
	"github.com/chenrensong/ygo/contracts"
	"github.com/chenrensong/ygo/core"
	"github.com/chenrensong/ygo/types"
)

// YcsBootstrap Bootstrap class to initialize dependencies and break circular references
type YcsBootstrap struct {
	initialized           bool
	typeReaderRegistry    contracts.ITypeReaderRegistry
	contentReaderRegistry contracts.IContentReaderRegistry
	contentFactory        contracts.IContentFactory
	mu                    sync.Mutex
}

var instance *YcsBootstrap
var once sync.Once

// GetInstance returns the singleton instance of YcsBootstrap
func GetInstance() *YcsBootstrap {
	once.Do(func() {
		instance = &YcsBootstrap{}
	})
	return instance
}

// Initialize initializes the Ycs system with dependency injection to avoid circular references
func (b *YcsBootstrap) Initialize() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return
	}

	b.typeReaderRegistry = core.NewTypeReaderRegistry()
	b.contentReaderRegistry = core.NewContentReaderRegistry()
	b.contentFactory = core.NewContentFactory()

	// Register all type readers directly here to ensure they are registered
	b.typeReaderRegistry.RegisterTypeReader(0, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return types.ReadYArray(decoder)
	}) // YArray
	b.typeReaderRegistry.RegisterTypeReader(1, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return types.ReadYMap(decoder)
	}) // YMap
	b.typeReaderRegistry.RegisterTypeReader(2, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return types.ReadYText(decoder)
	}) // YText

	// Register all content readers
	b.contentReaderRegistry.RegisterContentReader(1, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentDeleted(decoder)
	}) // Deleted
	b.contentReaderRegistry.RegisterContentReader(2, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentJson(decoder)
	}) // JSON
	b.contentReaderRegistry.RegisterContentReader(3, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentBinary(decoder)
	}) // Binary
	b.contentReaderRegistry.RegisterContentReader(4, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentString(decoder)
	}) // String
	b.contentReaderRegistry.RegisterContentReader(5, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentEmbed(decoder)
	}) // Embed
	b.contentReaderRegistry.RegisterContentReader(6, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentFormat(decoder)
	}) // Format
	b.contentReaderRegistry.RegisterContentReader(7, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentType(decoder)
	}) // Type
	b.contentReaderRegistry.RegisterContentReader(8, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentAny(decoder)
	}) // Any
	b.contentReaderRegistry.RegisterContentReader(9, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentDoc(decoder)
	}) // Doc

	content.SetTypeReaderRegistry(b.typeReaderRegistry)
	core.SetContentReaderRegistry(b.contentReaderRegistry)
	core.SetContentFactory(b.contentFactory)
	content.SetDocFactory(func(opts contracts.YDocOptions) contracts.IYDoc {
		return types.NewYDoc(opts)
	})

	b.initialized = true
}

// GetTypeReaderRegistry returns the type reader registry
func (b *YcsBootstrap) GetTypeReaderRegistry() contracts.ITypeReaderRegistry {
	return b.typeReaderRegistry
}

// GetContentReaderRegistry returns the content reader registry
func (b *YcsBootstrap) GetContentReaderRegistry() contracts.IContentReaderRegistry {
	return b.contentReaderRegistry
}

// GetContentFactory returns the content factory
func (b *YcsBootstrap) GetContentFactory() contracts.IContentFactory {
	return b.contentFactory
}
