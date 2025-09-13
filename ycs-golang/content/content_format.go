package content

import (
	"errors"

	"github.com/chenrensong/ygo/contracts"
)

const ContentFormatRef = 6

type ContentFormat struct {
	key   string
	value interface{}
}

func NewContentFormat(key string, value interface{}) *ContentFormat {
	return &ContentFormat{
		key:   key,
		value: value,
	}
}

func (c *ContentFormat) GetRef() int {
	return ContentFormatRef
}

func (c *ContentFormat) GetCountable() bool {
	return false
}

func (c *ContentFormat) GetLength() int {
	return 1
}

func (c *ContentFormat) GetKey() string {
	return c.key
}

func (c *ContentFormat) GetValue() interface{} {
	return c.value
}

func (c *ContentFormat) Copy() contracts.IContent {
	return &ContentFormat{key: c.key, value: c.value}
}

func (c *ContentFormat) GetContent() []interface{} {
	panic(errors.New("not implemented"))
}

func (c *ContentFormat) Splice(offset int) contracts.IContent {
	panic(errors.New("not implemented"))
}

func (c *ContentFormat) MergeWith(right contracts.IContent) bool {
	return false
}

func (c *ContentFormat) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Search markers are currently unsupported for rich text documents.
	// Check if parent implements array-like functionality and clear search markers if needed
	if arrayBase, ok := item.GetParent().(contracts.IYArrayBase); ok {
		arrayBase.ClearSearchMarkers()
	}
}

func (c *ContentFormat) Delete(transaction contracts.ITransaction) {
	// Do nothing
}

func (c *ContentFormat) Gc(store contracts.IStructStore) {
	// Do nothing
}

func (c *ContentFormat) Write(encoder contracts.IUpdateEncoder, offset int) {
	encoder.WriteKey(c.key)
	encoder.WriteJSON(c.value)
}

func ReadContentFormat(decoder contracts.IUpdateDecoder) *ContentFormat {
	key := decoder.ReadKey()
	value := decoder.ReadJSON()
	return &ContentFormat{key: key, value: value}
}
