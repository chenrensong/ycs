package content

import (
	"fmt"
	"ycs/contracts"
)

// ContentFactory represents the default implementation of content factory
type ContentFactory struct{}

// NewContentFactory creates a new ContentFactory
func NewContentFactory() *ContentFactory {
	return &ContentFactory{}
}

// CreateContentType creates content for a type
func (cf *ContentFactory) CreateContentType(abstractType contracts.IAbstractType) contracts.IContentEx {
	return NewContentType(abstractType)
}

// CreateContentDoc creates content for a document
func (cf *ContentFactory) CreateContentDoc(doc interface{}) contracts.IContentEx {
	if yDoc, ok := doc.(contracts.IYDoc); ok {
		return NewContentDoc(yDoc.GetOpts())
	}
	panic(fmt.Sprintf("Expected YDoc instance, got %T", doc))
}

// CreateContentBinary creates content for binary data
func (cf *ContentFactory) CreateContentBinary(ba []byte) contracts.IContentEx {
	return NewContentBinary(ba)
}

// CreateContentAny creates content for any value
func (cf *ContentFactory) CreateContentAny(value interface{}) contracts.IContentEx {
	return NewContentAny([]interface{}{value})
}

// CreateContentFormat creates content for format
func (cf *ContentFactory) CreateContentFormat(key string, value interface{}) contracts.IContentEx {
	return NewContentFormat(key, value)
}

// CreateContentString creates content for string
func (cf *ContentFactory) CreateContentString(text string) contracts.IContentEx {
	return NewContentString(text)
}

// CreateContentEmbed creates content for embed
func (cf *ContentFactory) CreateContentEmbed(embed interface{}) contracts.IContentEx {
	return NewContentEmbed(embed)
}

// CreateContent creates appropriate content based on value type
func (cf *ContentFactory) CreateContent(value interface{}) contracts.IContentEx {
	if value == nil {
		return cf.CreateContentAny(value)
	}

	switch v := value.(type) {
	case contracts.IYDoc:
		return cf.CreateContentDoc(v)
	case contracts.IAbstractType:
		return cf.CreateContentType(v)
	case []byte:
		return cf.CreateContentBinary(v)
	default:
		return cf.CreateContentAny(value)
	}
}
