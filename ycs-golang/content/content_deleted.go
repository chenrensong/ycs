package content

import (
	"errors"

	"github.com/chenrensong/ygo/contracts"
)

const ContentDeletedRef = 1

type ContentDeleted struct {
	length int
}

func NewContentDeleted(length int) *ContentDeleted {
	return &ContentDeleted{
		length: length,
	}
}

func (c *ContentDeleted) GetRef() int {
	return ContentDeletedRef
}

func (c *ContentDeleted) GetCountable() bool {
	return false
}

func (c *ContentDeleted) GetLength() int {
	return c.length
}

func (c *ContentDeleted) GetContent() []interface{} {
	panic(errors.New("not implemented"))
}

func (c *ContentDeleted) Copy() contracts.IContent {
	return &ContentDeleted{length: c.length}
}

func (c *ContentDeleted) Splice(offset int) contracts.IContent {
	right := &ContentDeleted{length: c.length - offset}
	c.length = offset
	return right
}

func (c *ContentDeleted) MergeWith(right contracts.IContent) bool {
	rightDeleted, ok := right.(*ContentDeleted)
	if !ok {
		return false
	}
	c.length += rightDeleted.length
	return true
}

func (c *ContentDeleted) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	transaction.GetDeleteSet().Add(item.GetID().Client, item.GetID().Clock, int64(c.length))
	item.MarkDeleted()
}

func (c *ContentDeleted) Delete(transaction contracts.ITransaction) {
	// Do nothing
}

func (c *ContentDeleted) Gc(store contracts.IStructStore) {
	// Do nothing
}

func (c *ContentDeleted) Write(encoder contracts.IUpdateEncoder, offset int) {
	encoder.WriteLength(c.length - offset)
}

func ReadContentDeleted(decoder contracts.IUpdateDecoder) *ContentDeleted {
	length := decoder.ReadLength()
	return &ContentDeleted{length: length}
}
