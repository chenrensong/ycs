package contracts

// YEventArgs represents event arguments for Y events
type YEventArgs struct {
	Event       IYEvent
	Transaction ITransaction
}

// NewYEventArgs creates a new YEventArgs instance
func NewYEventArgs(event IYEvent, transaction ITransaction) *YEventArgs {
	return &YEventArgs{
		Event:       event,
		Transaction: transaction,
	}
}

// YDeepEventArgs represents deep event arguments for Y events
type YDeepEventArgs struct {
	Events      []IYEvent
	Transaction ITransaction
}

// NewYDeepEventArgs creates a new YDeepEventArgs instance
func NewYDeepEventArgs(events []IYEvent, transaction ITransaction) *YDeepEventArgs {
	return &YDeepEventArgs{
		Events:      events,
		Transaction: transaction,
	}
}

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
