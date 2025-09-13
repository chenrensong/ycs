package content

import (
	"errors"

	"github.com/chenrensong/ygo/contracts"
)

const ContentEmbedRef = 5

type ContentEmbed struct {
	embed interface{}
}

func NewContentEmbed(embed interface{}) *ContentEmbed {
	return &ContentEmbed{
		embed: embed,
	}
}

func (c *ContentEmbed) GetRef() int {
	return ContentEmbedRef
}

func (c *ContentEmbed) GetCountable() bool {
	return true
}

func (c *ContentEmbed) GetLength() int {
	return 1
}

func (c *ContentEmbed) GetEmbed() interface{} {
	return c.embed
}

func (c *ContentEmbed) GetContent() []interface{} {
	return []interface{}{c.embed}
}

func (c *ContentEmbed) Copy() contracts.IContent {
	return &ContentEmbed{embed: c.embed}
}

func (c *ContentEmbed) Splice(offset int) contracts.IContent {
	panic(errors.New("not implemented"))
}

func (c *ContentEmbed) MergeWith(right contracts.IContent) bool {
	return false
}

func (c *ContentEmbed) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Do nothing
}

func (c *ContentEmbed) Delete(transaction contracts.ITransaction) {
	// Do nothing
}

func (c *ContentEmbed) Gc(store contracts.IStructStore) {
	// Do nothing
}

func (c *ContentEmbed) Write(encoder contracts.IUpdateEncoder, offset int) {
	encoder.WriteJSON(c.embed)
}

func ReadContentEmbed(decoder contracts.IUpdateDecoder) *ContentEmbed {
	content := decoder.ReadJSON()
	return &ContentEmbed{embed: content}
}
