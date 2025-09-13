package contracts

// IYEvent represents a Y event interface
type IYEvent interface {
	GetChanges() *ChangesCollection
	GetCurrentTarget() IAbstractType
	SetCurrentTarget(target IAbstractType)
	GetPath() []interface{}
	GetTarget() IAbstractType
	SetTarget(target IAbstractType)
	GetTransaction() ITransaction
	SetTransaction(transaction ITransaction)
}
