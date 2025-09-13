package content

import (
	"errors"

	"github.com/chenrensong/ygo/contracts"
)

const ContentBinaryRef = 3

type ContentBinary struct {
	content []byte
}

func NewContentBinary(data []byte) *ContentBinary {
	return &ContentBinary{
		content: data,
	}
}

func (c *ContentBinary) GetRef() int {
	return ContentBinaryRef
}

func (c *ContentBinary) GetCountable() bool {
	return true
}

func (c *ContentBinary) GetLength() int {
	return 1
}

func (c *ContentBinary) GetContent() []interface{} {
	return []interface{}{c.content}
}

func (c *ContentBinary) Copy() contracts.IContent {
	return &ContentBinary{content: c.content}
}

func (c *ContentBinary) Splice(offset int) contracts.IContent {
	panic(errors.New("not implemented"))
}

func (c *ContentBinary) MergeWith(right contracts.IContent) bool {
	return false
}

func (c *ContentBinary) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Do nothing
}

func (c *ContentBinary) Delete(transaction contracts.ITransaction) {
	// Do nothing
}

func (c *ContentBinary) Gc(store contracts.IStructStore) {
	// Do nothing
}

func (c *ContentBinary) Write(encoder contracts.IUpdateEncoder, offset int) {
	encoder.WriteBuffer(c.content)
}

func ReadContentBinary(decoder contracts.IUpdateDecoder) *ContentBinary {
	content := decoder.ReadBuffer()
	return &ContentBinary{content: content}
}
