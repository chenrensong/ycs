package main

import (
	"sync"
	"ycs/content"
	"ycs/contracts"
	"ycs/core"
)

// YcsBootstrap provides bootstrap functionality to initialize dependencies and break circular references
type YcsBootstrap struct {
	initialized bool
	mutex       sync.Mutex
}

var bootstrap = &YcsBootstrap{}

// Initialize initializes the Ycs system with dependency injection to avoid circular references
func Initialize() {
	bootstrap.mutex.Lock()
	defer bootstrap.mutex.Unlock()

	if bootstrap.initialized {
		return
	}

	// Register type readers directly in ContentType to match C# implementation
	content.RegisterTypeReader(0, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return core.ReadYArray(decoder)
	}) // YArray
	content.RegisterTypeReader(1, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return core.ReadYMap(decoder)
	}) // YMap
	content.RegisterTypeReader(2, func(decoder contracts.IUpdateDecoder) contracts.IAbstractType {
		return core.ReadYText(decoder)
	}) // YText

	// Set document factory
	content.SetDocFactory(func(opts *contracts.YDocOptions) contracts.IYDoc {
		return core.NewYDoc(*opts)
	})

	bootstrap.initialized = true
}

// IsInitialized returns whether the bootstrap has been initialized
func IsInitialized() bool {
	bootstrap.mutex.Lock()
	defer bootstrap.mutex.Unlock()
	return bootstrap.initialized
}
