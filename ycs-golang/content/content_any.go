package content

import (
	"github.com/chenrensong/ygo/contracts"
)

const ContentAnyRef = 8

type ContentAny struct {
	content []interface{}
}

func NewContentAny(content []interface{}) *ContentAny {
	contentCopy := make([]interface{}, len(content))
	copy(contentCopy, content)
	return &ContentAny{
		content: contentCopy,
	}
}

func (c *ContentAny) GetRef() int {
	return ContentAnyRef
}

func (c *ContentAny) GetCountable() bool {
	return true
}

func (c *ContentAny) GetLength() int {
	return len(c.content)
}

func (c *ContentAny) GetContent() []interface{} {
	result := make([]interface{}, len(c.content))
	copy(result, c.content)
	return result
}

func (c *ContentAny) Copy() contracts.IContent {
	contentCopy := make([]interface{}, len(c.content))
	copy(contentCopy, c.content)
	return &ContentAny{content: contentCopy}
}

func (c *ContentAny) Splice(offset int) contracts.IContent {
	right := &ContentAny{
		content: make([]interface{}, len(c.content)-offset),
	}
	copy(right.content, c.content[offset:])
	c.content = c.content[:offset]
	return right
}

func (c *ContentAny) MergeWith(right contracts.IContent) bool {
	rightAny, ok := right.(*ContentAny)
	if !ok {
		return false
	}
	c.content = append(c.content, rightAny.content...)
	return true
}

func (c *ContentAny) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Do nothing
}

func (c *ContentAny) Delete(transaction contracts.ITransaction) {
	// Do nothing
}

func (c *ContentAny) Gc(store contracts.IStructStore) {
	// Do nothing
}

func (c *ContentAny) Write(encoder contracts.IUpdateEncoder, offset int) {
	length := len(c.content)
	encoder.WriteLength(length - offset)

	for i := offset; i < length; i++ {
		encoder.WriteAny(c.content[i])
	}
}

func ReadContentAny(decoder contracts.IUpdateDecoder) *ContentAny {
	length := decoder.ReadLength()
	content := make([]interface{}, length)

	for i := 0; i < length; i++ {
		content[i] = decoder.ReadAny()
	}

	return &ContentAny{content: content}
}
