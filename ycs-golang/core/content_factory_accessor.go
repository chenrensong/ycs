// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"github.com/chenrensong/ygo/contracts"
)

// ContentFactoryAccessor provides access to the content factory
type ContentFactoryAccessor struct{}

var factory contracts.IContentFactory

// SetFactory sets the content factory
func SetFactory(contentFactory contracts.IContentFactory) {
	factory = contentFactory
}

// GetFactory gets the content factory
func GetFactory() contracts.IContentFactory {
	return factory
}
