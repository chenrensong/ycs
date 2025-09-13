// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"fmt"

	"github.com/chenrensong/ygo/content"
	"github.com/chenrensong/ygo/contracts"
	"github.com/chenrensong/ygo/types"
)

// ContentFactory is the default implementation of content factory
type ContentFactory struct{}

// NewContentFactory creates a new ContentFactory
func NewContentFactory() contracts.IContentFactory {
	return &ContentFactory{}
}

// CreateContentType creates a ContentType
func (cf *ContentFactory) CreateContentType(abstractType contracts.IAbstractType) contracts.IContent {
	return content.NewContentType(abstractType)
}

// CreateContentDoc creates a ContentDoc
func (cf *ContentFactory) CreateContentDoc(doc interface{}) contracts.IContent {
	if yDoc, ok := doc.(*types.YDoc); ok {
		return content.NewContentDoc(yDoc)
	}
	panic(fmt.Sprintf("Expected YDoc instance, got %T", doc))
}

// CreateContentBinary creates a ContentBinary
func (cf *ContentFactory) CreateContentBinary(ba []byte) contracts.IContent {
	return content.NewContentBinary(ba)
}

// CreateContentAny creates a ContentAny
func (cf *ContentFactory) CreateContentAny(value interface{}) contracts.IContent {
	return content.NewContentAny([]interface{}{value})
}

// CreateContentFormat creates a ContentFormat
func (cf *ContentFactory) CreateContentFormat(key string, value interface{}) contracts.IContent {
	return content.NewContentFormat(key, value)
}

// CreateContentString creates a ContentString
func (cf *ContentFactory) CreateContentString(text string) contracts.IContent {
	return content.NewContentString(text)
}

// CreateContentEmbed creates a ContentEmbed
func (cf *ContentFactory) CreateContentEmbed(embed interface{}) contracts.IContent {
	return content.NewContentEmbed(embed)
}

// CreateContent creates content based on the value type
func (cf *ContentFactory) CreateContent(value interface{}) contracts.IContent {
	if value == nil {
		return cf.CreateContentAny(value)
	}

	switch v := value.(type) {
	case *types.YDoc:
		return cf.CreateContentDoc(v)
	case contracts.IAbstractType:
		return cf.CreateContentType(v)
	case []byte:
		return cf.CreateContentBinary(v)
	default:
		return cf.CreateContentAny(value)
	}
}
