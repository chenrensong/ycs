package contracts

// ITransaction represents a transaction interface
type ITransaction interface {
	GetAfterState() map[int64]int64
	SetAfterState(afterState map[int64]int64)
	GetBeforeState() map[int64]int64
	SetBeforeState(beforeState map[int64]int64)
	GetChanged() map[IAbstractType]map[string]struct{}
	GetChangedParentTypes() map[IAbstractType][]IYEvent
	GetDeleteSet() IDeleteSet
	GetDoc() IYDoc
	GetLocal() bool
	GetMeta() map[string]interface{}
	GetOrigin() interface{}
	GetSubdocsAdded() map[IYDoc]struct{}
	GetSubdocsLoaded() map[IYDoc]struct{}
	GetSubdocsRemoved() map[IYDoc]struct{}
	GetMergeStructs() []IStructItem
	AddMergeStruct(item IStructItem)

	AddChangedTypeToTransaction(abstractType IAbstractType, parentSub string)
	GetNextID() StructID
	RedoItem(item IStructItem, redoItems map[IStructItem]struct{}) IStructItem
	WriteUpdateMessageFromTransaction(encoder IUpdateEncoder) bool
}
