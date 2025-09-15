package main

import (
	"sync"
	"ycs/content"
	"ycs/contracts"
	"ycs/core"
)

// YcsBootstrap provides bootstrap functionality to initialize dependencies and break circular references
type YcsBootstrap struct {
	initialized           bool
	typeReaderRegistry    contracts.ITypeReaderRegistry
	contentReaderRegistry contracts.IContentReaderRegistry
	contentFactory        contracts.IContentFactory
	mutex                 sync.Mutex
}

var bootstrap = &YcsBootstrap{}

// Initialize initializes the Ycs system with dependency injection to avoid circular references
func Initialize() {
	bootstrap.mutex.Lock()
	defer bootstrap.mutex.Unlock()

	if bootstrap.initialized {
		return
	}

	bootstrap.typeReaderRegistry = core.NewTypeReaderRegistry()
	bootstrap.contentReaderRegistry = content.NewContentReaderRegistry()
	bootstrap.contentFactory = content.NewContentFactory()

	// Register all type readers directly here to ensure they are registered
	bootstrap.typeReaderRegistry.RegisterTypeReader(0, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return core.ReadYArray(decoder)
	}) // YArray
	bootstrap.typeReaderRegistry.RegisterTypeReader(1, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return core.ReadYMap(decoder)
	}) // YMap
	bootstrap.typeReaderRegistry.RegisterTypeReader(2, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return core.ReadYText(decoder)
	}) // YText

	// Register all content readers
	bootstrap.contentReaderRegistry.RegisterContentReader(1, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentDeleted(decoder)
	}) // Deleted
	bootstrap.contentReaderRegistry.RegisterContentReader(2, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentJson(decoder)
	}) // JSON
	bootstrap.contentReaderRegistry.RegisterContentReader(3, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentBinary(decoder)
	}) // Binary
	bootstrap.contentReaderRegistry.RegisterContentReader(4, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentString(decoder)
	}) // String
	bootstrap.contentReaderRegistry.RegisterContentReader(5, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentEmbed(decoder)
	}) // Embed
	bootstrap.contentReaderRegistry.RegisterContentReader(6, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentFormat(decoder)
	}) // Format
	bootstrap.contentReaderRegistry.RegisterContentReader(7, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentType(decoder)
	}) // Type
	bootstrap.contentReaderRegistry.RegisterContentReader(8, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentAny(decoder)
	}) // Any
	bootstrap.contentReaderRegistry.RegisterContentReader(9, func(decoder contracts.IUpdateDecoder) contracts.IContent {
		return content.ReadContentDoc(decoder)
	}) // Doc

	// Set global registries and factories using the accessor pattern
	content.SetGlobalFactory(bootstrap.contentFactory)
	core.SetGlobalContentReaderRegistry(bootstrap.contentReaderRegistry)
	core.SetGlobalTypeReaderRegistry(bootstrap.typeReaderRegistry)

	bootstrap.initialized = true
}

// IsInitialized returns whether the bootstrap has been initialized
func IsInitialized() bool {
	bootstrap.mutex.Lock()
	defer bootstrap.mutex.Unlock()
	return bootstrap.initialized
}

// GetTypeReaderRegistry returns the type reader registry (for internal use)
func GetTypeReaderRegistry() contracts.ITypeReaderRegistry {
	return bootstrap.typeReaderRegistry
}

// GetContentReaderRegistry returns the content reader registry (for internal use)
func GetContentReaderRegistry() contracts.IContentReaderRegistry {
	return bootstrap.contentReaderRegistry
}

// GetContentFactory returns the content factory (for internal use)
func GetContentFactory() contracts.IContentFactory {
	return bootstrap.contentFactory
}
