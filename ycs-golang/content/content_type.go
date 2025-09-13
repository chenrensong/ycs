package content

import (
	"errors"

	"github.com/chenrensong/ygo/contracts"
)

const ContentTypeRef = 7

var typeReaderRegistry contracts.ITypeReaderRegistry

type ContentType struct {
	abstractType contracts.IAbstractType
}

func NewContentType(abstractType contracts.IAbstractType) *ContentType {
	return &ContentType{
		abstractType: abstractType,
	}
}

func SetTypeReaderRegistry(registry contracts.ITypeReaderRegistry) {
	typeReaderRegistry = registry
}

func (c *ContentType) GetRef() int {
	return ContentTypeRef
}

func (c *ContentType) GetCountable() bool {
	return true
}

func (c *ContentType) GetLength() int {
	return 1
}

func (c *ContentType) GetType() contracts.IAbstractType {
	return c.abstractType
}

func (c *ContentType) GetContent() []interface{} {
	return []interface{}{c.abstractType}
}

func (c *ContentType) Copy() contracts.IContent {
	return &ContentType{abstractType: c.abstractType.InternalCopy()}
}

func (c *ContentType) Splice(offset int) contracts.IContent {
	panic(errors.New("not implemented"))
}

func (c *ContentType) MergeWith(right contracts.IContent) bool {
	return false
}

func (c *ContentType) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	c.abstractType.Integrate(transaction.GetDoc(), item)
}

func (c *ContentType) Delete(transaction contracts.ITransaction) {
	item := c.abstractType.GetStart()

	for item != nil {
		if !item.GetDeleted() {
			item.Delete(transaction)
		} else {
			// This will be gc'd later and we want to merge it if possible.
			// We try to merge all deleted items each transaction,
			// but we have no knowledge about that this needs to merged
			// since it is not in transaction. Hence we add it to transaction._mergeStructs.
			transaction.GetMergeStructs().Add(item)
		}

		item = item.GetRight()
	}

	for _, valueItem := range c.abstractType.GetMap() {
		if !valueItem.GetDeleted() {
			valueItem.Delete(transaction)
		} else {
			// Same as above.
			transaction.GetMergeStructs().Add(valueItem)
		}
	}

	transaction.GetChanged().Remove(c.abstractType)
}

func (c *ContentType) Gc(store contracts.IStructStore) {
	item := c.abstractType.GetStart()
	for item != nil {
		item.Gc(store, true)
		item = item.GetRight()
	}

	c.abstractType.SetStart(nil)

	for _, valueItem := range c.abstractType.GetMap() {
		currentItem := valueItem
		for currentItem != nil {
			currentItem.Gc(store, true)
			currentItem = currentItem.GetLeft()
		}
	}

	// Clear the map
	newMap := make(map[string]contracts.IStructItem)
	c.abstractType.SetMap(newMap)
}

func (c *ContentType) Write(encoder contracts.IUpdateEncoder, offset int) {
	c.abstractType.Write(encoder)
}

func ReadContentType(decoder contracts.IUpdateDecoder) (*ContentType, error) {
	if typeReaderRegistry == nil {
		return nil, errors.New("TypeReaderRegistry not initialized. Call SetTypeReaderRegistry() first")
	}

	typeRef := decoder.ReadTypeRef()
	abstractType := typeReaderRegistry.ReadType(typeRef, decoder)
	return &ContentType{abstractType: abstractType}, nil
}
