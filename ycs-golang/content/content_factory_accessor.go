package content

import (
	"errors"
	"ycs/contracts"
)

// ContentFactoryAccessor provides global access to ContentFactory to avoid circular dependencies
// This is initialized by YcsBootstrap
var contentFactoryAccessor *ContentFactoryAccessor

// ContentFactoryAccessor holds the global content factory instance
type ContentFactoryAccessor struct {
	factory contracts.IContentFactory
}

// NewContentFactoryAccessor creates a new ContentFactoryAccessor
func NewContentFactoryAccessor() *ContentFactoryAccessor {
	return &ContentFactoryAccessor{}
}

// SetFactory sets the ContentFactory instance (called by YcsBootstrap)
func (cfa *ContentFactoryAccessor) SetFactory(factory contracts.IContentFactory) {
	cfa.factory = factory
}

// GetFactory gets the ContentFactory instance
func (cfa *ContentFactoryAccessor) GetFactory() (contracts.IContentFactory, error) {
	if cfa.factory == nil {
		return nil, errors.New("ContentFactory not initialized. Call YcsBootstrap.Initialize() first")
	}
	return cfa.factory, nil
}

// Global functions for easy access
func init() {
	contentFactoryAccessor = NewContentFactoryAccessor()
}

// SetGlobalFactory sets the global ContentFactory instance
func SetGlobalFactory(factory contracts.IContentFactory) {
	contentFactoryAccessor.SetFactory(factory)
}

// GetGlobalFactory gets the global ContentFactory instance
func GetGlobalFactory() (contracts.IContentFactory, error) {
	return contentFactoryAccessor.GetFactory()
}

// MustGetGlobalFactory gets the global ContentFactory instance, panics if not set
func MustGetGlobalFactory() contracts.IContentFactory {
	factory, err := GetGlobalFactory()
	if err != nil {
		panic(err)
	}
	return factory
}
