package content

import (
	"errors"

	"github.com/chenrensong/ygo/contracts"
)

const ContentDocRef = 9

var docFactory func(*contracts.YDocOptions) contracts.IYDoc

type ContentDoc struct {
	doc  contracts.IYDoc
	opts *contracts.YDocOptions
}

func NewContentDoc(doc contracts.IYDoc) (*ContentDoc, error) {
	if doc.GetItem() != nil {
		return nil, errors.New("this document was already integrated as a sub-document. You should create a second instance instead with the same guid")
	}

	opts := &contracts.YDocOptions{}

	if !doc.GetGc() {
		opts.Gc = false
	}

	if doc.GetAutoLoad() {
		opts.AutoLoad = true
	}

	if doc.GetMeta() != nil {
		opts.Meta = doc.GetMeta()
	}

	return &ContentDoc{
		doc:  doc,
		opts: opts,
	}, nil
}

func (c *ContentDoc) GetRef() int {
	return ContentDocRef
}

func (c *ContentDoc) GetCountable() bool {
	return true
}

func (c *ContentDoc) GetLength() int {
	return 1
}

func (c *ContentDoc) GetDoc() contracts.IYDoc {
	return c.doc
}

func (c *ContentDoc) GetOpts() *contracts.YDocOptions {
	return c.opts
}

func (c *ContentDoc) GetContent() []interface{} {
	return []interface{}{c.doc}
}

func (c *ContentDoc) Copy() contracts.IContent {
	result, _ := NewContentDoc(c.doc)
	return result
}

func (c *ContentDoc) Splice(offset int) contracts.IContent {
	panic(errors.New("not implemented"))
}

func (c *ContentDoc) MergeWith(right contracts.IContent) bool {
	return false
}

func (c *ContentDoc) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// This needs to be reflected in doc.destroy as well.
	c.doc.SetItem(item)
	transaction.GetSubdocsAdded().Add(c.doc)

	if c.doc.GetShouldLoad() {
		transaction.GetSubdocsLoaded().Add(c.doc)
	}
}

func (c *ContentDoc) Delete(transaction contracts.ITransaction) {
	if transaction.GetSubdocsAdded().Contains(c.doc) {
		transaction.GetSubdocsAdded().Remove(c.doc)
	} else {
		transaction.GetSubdocsRemoved().Add(c.doc)
	}
}

func (c *ContentDoc) Gc(store contracts.IStructStore) {
	// Do nothing
}

func (c *ContentDoc) Write(encoder contracts.IUpdateEncoder, offset int) {
	// 32 digits separated by hyphens, no braces.
	encoder.WriteString(c.doc.GetGuid())
	c.opts.Write(encoder, offset)
}

func SetDocFactory(factory func(*contracts.YDocOptions) contracts.IYDoc) {
	docFactory = factory
}

func ReadContentDoc(decoder contracts.IUpdateDecoder) (*ContentDoc, error) {
	guidStr := decoder.ReadString()

	opts := contracts.ReadYDocOptions(decoder)
	opts.Guid = guidStr

	if docFactory == nil {
		return nil, errors.New("DocFactory not initialized. Call SetDocFactory() first")
	}

	doc := docFactory(opts)
	return &ContentDoc{
		doc:  doc,
		opts: opts,
	}, nil
}
