package contracts

// IStructItem represents a struct item interface
type IStructItem interface {
	MergeWith(right IStructItem) bool
	Delete(transaction ITransaction)
	Integrate(transaction ITransaction, offset int)
	GetMissing(transaction ITransaction, store IStructStore) *int64
	Write(encoder IUpdateEncoder, offset int)

	GetID() StructID
	SetID(id StructID)
	GetContent() IContentEx
	SetContent(content IContentEx)
	GetCountable() bool
	SetCountable(countable bool)
	GetDeleted() bool
	GetKeep() bool
	SetKeep(keep bool)
	GetLastID() StructID
	GetLeft() IStructItem
	SetLeft(left IStructItem)
	GetLeftOrigin() *StructID
	SetLeftOrigin(leftOrigin *StructID)
	GetMarker() bool
	SetMarker(marker bool)
	GetNext() IStructItem
	GetParent() interface{}
	SetParent(parent interface{})
	GetParentSub() string
	SetParentSub(parentSub string)
	GetPrev() IStructItem
	GetRedone() *StructID
	SetRedone(redone *StructID)
	GetRight() IStructItem
	SetRight(right IStructItem)
	GetRightOrigin() *StructID
	SetRightOrigin(rightOrigin *StructID)
	GetLength() int
	SetLength(length int)

	Gc(store IStructStore, parentGCd bool)
	IsVisible(snap ISnapshot) bool
	KeepItemAndParents(value bool)
	MarkDeleted()
	SplitItem(transaction ITransaction, diff int) IStructItem
}

// IAbstractType represents an abstract type interface
type IAbstractType interface {
	GetDoc() IYDoc
	GetItem() IStructItem
	SetItem(item IStructItem)
	GetLength() int
	SetLength(length int)
	GetMap() map[string]IStructItem
	SetMap(m map[string]IStructItem)
	GetParent() IAbstractType
	GetStart() IStructItem
	SetStart(start IStructItem)

	CallDeepEventHandlerListeners(events []IYEvent, transaction ITransaction)
	CallObserver(transaction ITransaction, parentSubs map[string]struct{})
	CallTypeObservers(transaction ITransaction, evt IYEvent)
	FindRootTypeKey() string
	Integrate(doc IYDoc, item IStructItem)
	InternalClone() IAbstractType
	InternalCopy() IAbstractType
	InvokeEventHandlers(evt IYEvent, transaction ITransaction)
	Write(encoder IUpdateEncoder)
	First() IStructItem
}

// IContent represents content interface
type IContent interface {
	GetCountable() bool
	GetLength() int
	GetContent() []interface{}
	Copy() IContent
	Splice(offset int) IContent
	MergeWith(right IContent) bool
}

// IContentEx represents extended content interface
type IContentEx interface {
	IContent
	GetRef() int
	Integrate(transaction ITransaction, item IStructItem)
	Delete(transaction ITransaction)
	Gc(store IStructStore)
	Write(encoder IUpdateEncoder, offset int)
}

// This file contains the core interfaces for the YCS contracts
