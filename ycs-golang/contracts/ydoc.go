package contracts

import "io"

// EventHandler represents various event handler function types
type BeforeObserverCallsHandler func(ITransaction)
type BeforeTransactionHandler func(ITransaction)
type AfterTransactionHandler func(ITransaction)
type AfterTransactionCleanupHandler func(ITransaction)
type BeforeAllTransactionsHandler func()
type AfterAllTransactionsHandler func([]ITransaction)
type UpdateV2Handler func([]byte, interface{}, ITransaction)
type DestroyedHandler func()
type SubdocsChangedHandler func(map[IYDoc]struct{}, map[IYDoc]struct{}, map[IYDoc]struct{})

// IYDoc represents a Y document interface
type IYDoc interface {
	GetAutoLoad() bool
	GetClientID() int
	SetClientID(clientID int)
	GetGc() bool
	GetGcFilter() func(IStructItem) bool
	GetGuid() string
	GetItem() IStructItem
	SetItem(item IStructItem)
	GetMeta() map[string]string
	GetOpts() *YDocOptions
	GetShare() map[string]IAbstractType
	SetShare(share map[string]IAbstractType)
	GetShouldLoad() bool
	SetShouldLoad(shouldLoad bool)
	GetStore() IStructStore
	SetStore(store IStructStore)
	GetSubdocs() map[IYDoc]struct{}
	GetTransaction() ITransaction
	SetTransaction(transaction ITransaction)
	GetTransactionCleanups() []ITransaction
	SetTransactionCleanups(transactionCleanups []ITransaction)

	ApplyUpdateV2(update []byte, transactionOrigin interface{}, local ...bool)         // local defaults to false
	ApplyUpdateV2Stream(input io.Reader, transactionOrigin interface{}, local ...bool) // local defaults to false
	CloneOptionsWithNewGuid() *YDocOptions
	CreateSnapshot() ISnapshot
	Destroy()
	EncodeStateAsUpdateV2(encodedTargetStateVector ...[]byte) []byte // optional parameter
	EncodeStateVectorV2() []byte
	FindRootTypeKey(abstractType IAbstractType) string
	Get(name string, typeConstructor func() IAbstractType) IAbstractType // Generic equivalent
	GetArray(name ...string) IYArray                                     // name defaults to ""
	GetMap(name ...string) IYMap                                         // name defaults to ""
	GetSubdocGuids() []string
	GetText(name ...string) IYText // name defaults to ""
	InvokeAfterAllTransactions(transactions []ITransaction)
	InvokeBeforeAllTransactions()
	InvokeDestroyed()
	InvokeOnAfterTransaction(transaction ITransaction)
	InvokeOnAfterTransactionCleanup(transaction ITransaction)
	InvokeOnBeforeObserverCalls(transaction ITransaction)
	InvokeOnBeforeTransaction(transaction ITransaction)
	InvokeSubdocsChanged(loaded map[IYDoc]struct{}, added map[IYDoc]struct{}, removed map[IYDoc]struct{})
	InvokeUpdateV2(transaction ITransaction)
	Load()
	Transact(fun func(ITransaction), origin interface{}, local ...bool) // local defaults to true
	WriteStateAsUpdate(encoder IUpdateEncoder, targetStateVector map[int64]int64) error
	WriteStateVector(encoder IDSEncoder) error

	// Event handlers - Go doesn't have events like C#, so we use function fields
	OnBeforeObserverCalls(handler BeforeObserverCallsHandler)
	OnBeforeTransaction(handler BeforeTransactionHandler)
	OnAfterTransaction(handler AfterTransactionHandler)
	OnAfterTransactionCleanup(handler AfterTransactionCleanupHandler)
	OnBeforeAllTransactions(handler BeforeAllTransactionsHandler)
	OnAfterAllTransactions(handler AfterAllTransactionsHandler)
	OnUpdateV2(handler UpdateV2Handler)
	OnDestroyed(handler DestroyedHandler)
	OnSubdocsChanged(handler SubdocsChangedHandler)
}
