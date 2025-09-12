package structs

// Content interface represents the content of an item
type Content interface {
	Countable() bool
	Length() int
	GetContent() []interface{}
	Copy() Content
	Splice(offset int) Content
	MergeWith(right Content) bool
}

// ContentEx interface extends Content with additional methods
type ContentEx interface {
	Content
	Ref() int
	Integrate(transaction *Transaction, item *Item)
	Delete(transaction *Transaction)
	Gc(store *StructStore)
	Write(encoder IUpdateEncoder, offset int)
}