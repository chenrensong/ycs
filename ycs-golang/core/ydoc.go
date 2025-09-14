package core

import (
	"bytes"
	"crypto/rand"
	"io"
	"math/big"
	"sync"
	"ycs/contracts"
)

// YDoc represents a Yjs instance that handles the state of shared data
type YDoc struct {
	opts                contracts.YDocOptions
	shouldLoad          bool
	transactionCleanups []contracts.ITransaction
	transaction         contracts.ITransaction
	subdocs             map[contracts.IYDoc]struct{}
	item                contracts.IStructItem
	share               map[string]contracts.IAbstractType
	clientID            int64
	store               contracts.IStructStore
	mutex               sync.RWMutex

	// Event handlers
	beforeObserverCalls     func(contracts.ITransaction)
	beforeTransaction       func(contracts.ITransaction)
	afterTransaction        func(contracts.ITransaction)
	afterTransactionCleanup func(contracts.ITransaction)
	beforeAllTransactions   func()
	afterAllTransactions    func([]contracts.ITransaction)
	updateV2                func([]byte, interface{}, contracts.ITransaction)
	destroyed               func()
	subdocsChanged          func(map[contracts.IYDoc]struct{}, map[contracts.IYDoc]struct{}, map[contracts.IYDoc]struct{})
}

// NewYDoc creates a new YDoc instance
func NewYDoc(opts contracts.YDocOptions) *YDoc {
	if opts.Guid == "" {
		opts.Guid = generateGUID()
	}

	doc := &YDoc{
		opts:                opts,
		transactionCleanups: make([]contracts.ITransaction, 0),
		subdocs:             make(map[contracts.IYDoc]struct{}),
		share:               make(map[string]contracts.IAbstractType),
		clientID:            generateNewClientID(),
		store:               NewStructStore(),
		shouldLoad:          opts.AutoLoad,
	}

	return doc
}

// generateNewClientID generates a new random client ID
func generateNewClientID() int64 {
	max := big.NewInt(int64(^uint(0) >> 1)) // max int64
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback to a simple method if crypto/rand fails
		return 12345 // This should be improved in production
	}
	return n.Int64()
}

// generateGUID generates a simple GUID
func generateGUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "default-guid"
	}
	return string(b)
}

// GetOpts returns the document options
func (ydoc *YDoc) GetOpts() contracts.YDocOptions {
	return ydoc.opts
}

// GetGuid returns the document GUID
func (ydoc *YDoc) GetGuid() string {
	return ydoc.opts.Guid
}

// GetGc returns whether garbage collection is enabled
func (ydoc *YDoc) GetGc() bool {
	return ydoc.opts.Gc
}

// GetGcFilter returns the garbage collection filter
func (ydoc *YDoc) GetGcFilter() func(contracts.IStructItem) bool {
	return ydoc.opts.GcFilter
}

// GetAutoLoad returns whether auto-load is enabled
func (ydoc *YDoc) GetAutoLoad() bool {
	return ydoc.opts.AutoLoad
}

// GetMeta returns the metadata
func (ydoc *YDoc) GetMeta() map[string]string {
	return ydoc.opts.Meta
}

// GetShouldLoad returns whether the document should load
func (ydoc *YDoc) GetShouldLoad() bool {
	ydoc.mutex.RLock()
	defer ydoc.mutex.RUnlock()
	return ydoc.shouldLoad
}

// SetShouldLoad sets whether the document should load
func (ydoc *YDoc) SetShouldLoad(shouldLoad bool) {
	ydoc.mutex.Lock()
	defer ydoc.mutex.Unlock()
	ydoc.shouldLoad = shouldLoad
}

// GetTransactionCleanups returns the transaction cleanups
func (ydoc *YDoc) GetTransactionCleanups() []contracts.ITransaction {
	ydoc.mutex.RLock()
	defer ydoc.mutex.RUnlock()
	return ydoc.transactionCleanups
}

// GetTransaction returns the current transaction
func (ydoc *YDoc) GetTransaction() contracts.ITransaction {
	ydoc.mutex.RLock()
	defer ydoc.mutex.RUnlock()
	return ydoc.transaction
}

// SetTransaction sets the current transaction
func (ydoc *YDoc) SetTransaction(transaction contracts.ITransaction) {
	ydoc.mutex.Lock()
	defer ydoc.mutex.Unlock()
	ydoc.transaction = transaction
}

// GetSubdocs returns the subdocuments
func (ydoc *YDoc) GetSubdocs() map[contracts.IYDoc]struct{} {
	ydoc.mutex.RLock()
	defer ydoc.mutex.RUnlock()
	return ydoc.subdocs
}

// GetItem returns the item
func (ydoc *YDoc) GetItem() contracts.IStructItem {
	ydoc.mutex.RLock()
	defer ydoc.mutex.RUnlock()
	return ydoc.item
}

// SetItem sets the item
func (ydoc *YDoc) SetItem(item contracts.IStructItem) {
	ydoc.mutex.Lock()
	defer ydoc.mutex.Unlock()
	ydoc.item = item
}

// GetShare returns the shared types
func (ydoc *YDoc) GetShare() map[string]contracts.IAbstractType {
	ydoc.mutex.RLock()
	defer ydoc.mutex.RUnlock()
	return ydoc.share
}

// GetClientID returns the client ID
func (ydoc *YDoc) GetClientID() int64 {
	ydoc.mutex.RLock()
	defer ydoc.mutex.RUnlock()
	return ydoc.clientID
}

// SetClientID sets the client ID
func (ydoc *YDoc) SetClientID(clientID int64) {
	ydoc.mutex.Lock()
	defer ydoc.mutex.Unlock()
	ydoc.clientID = clientID
}

// GetStore returns the struct store
func (ydoc *YDoc) GetStore() contracts.IStructStore {
	return ydoc.store
}

// Load notifies the parent document that you request to load data into this subdocument
func (ydoc *YDoc) Load() {
	item := ydoc.GetItem()
	if item != nil && !ydoc.GetShouldLoad() {
		parent := item.GetParent().(contracts.IAbstractType)
		parent.GetDoc().Transact(func(tr contracts.ITransaction) {
			tr.GetSubdocsLoaded()[ydoc] = struct{}{}
		}, nil, true)
	}
	ydoc.SetShouldLoad(true)
}

// CreateSnapshot creates a snapshot of the current document state
func (ydoc *YDoc) CreateSnapshot() contracts.ISnapshot {
	return NewSnapshot(NewDeleteSet(ydoc.store), ydoc.store.GetStateVector())
}

// GetSubdocGuids returns the GUIDs of all subdocuments
func (ydoc *YDoc) GetSubdocGuids() []string {
	var guids []string
	for subdoc := range ydoc.GetSubdocs() {
		guids = append(guids, subdoc.GetGuid())
	}
	return guids
}

// Destroy destroys the document and all subdocuments
func (ydoc *YDoc) Destroy() {
	for subdoc := range ydoc.GetSubdocs() {
		subdoc.Destroy()
	}

	item := ydoc.GetItem()
	if item != nil {
		ydoc.SetItem(nil)
		// Handle content doc logic here
		// This is simplified compared to the C# version
		if ydoc.destroyed != nil {
			ydoc.destroyed()
		}
	}
}

// Transact bundles changes in a transaction
func (ydoc *YDoc) Transact(fun func(contracts.ITransaction), origin interface{}, local bool) {
	initialCall := false
	if ydoc.GetTransaction() == nil {
		initialCall = true
		transaction := NewTransaction(ydoc, origin, local)
		ydoc.SetTransaction(transaction)
		ydoc.mutex.Lock()
		ydoc.transactionCleanups = append(ydoc.transactionCleanups, transaction)
		isFirst := len(ydoc.transactionCleanups) == 1
		ydoc.mutex.Unlock()

		if isFirst && ydoc.beforeAllTransactions != nil {
			ydoc.beforeAllTransactions()
		}

		if ydoc.beforeTransaction != nil {
			ydoc.beforeTransaction(transaction)
		}
	}

	defer func() {
		if initialCall && len(ydoc.GetTransactionCleanups()) > 0 && ydoc.GetTransactionCleanups()[0] == ydoc.GetTransaction() {
			CleanupTransactions(ydoc.GetTransactionCleanups(), 0)
		}
	}()

	fun(ydoc.GetTransaction())
}

// GetArray returns or creates a YArray with the given name
func (ydoc *YDoc) GetArray(name string) contracts.IYArray {
	return ydoc.Get(name).(contracts.IYArray)
}

// GetMap returns or creates a YMap with the given name
func (ydoc *YDoc) GetMap(name string) contracts.IYMap {
	return ydoc.Get(name).(contracts.IYMap)
}

// GetText returns or creates a YText with the given name
func (ydoc *YDoc) GetText(name string) contracts.IYText {
	return ydoc.Get(name).(contracts.IYText)
}

// Get returns or creates a shared type with the given name
func (ydoc *YDoc) Get(name string) contracts.IAbstractType {
	ydoc.mutex.Lock()
	defer ydoc.mutex.Unlock()

	if existingType, exists := ydoc.share[name]; exists {
		return existingType
	}

	// This is a simplified version - in practice, you'd need type factories
	var newType contracts.IAbstractType
	switch name {
	default:
		// Create a generic abstract type - this needs proper implementation
		newType = NewAbstractType()
	}

	newType.Integrate(ydoc, nil)
	ydoc.share[name] = newType
	return newType
}

// ApplyUpdateV2 applies an update to the document
func (ydoc *YDoc) ApplyUpdateV2(input io.Reader, transactionOrigin interface{}, local bool) error {
	return ydoc.Transact(func(tr contracts.ITransaction) {
		decoder := NewUpdateDecoderV2(input)
		err := ReadStructs(decoder, tr, ydoc.store)
		if err != nil {
			// Handle error - in Go we'd typically return it
			panic(err)
		}
		ydoc.store.ReadAndApplyDeleteSet(decoder, tr)
	}, transactionOrigin, local)
}

// ApplyUpdateV2Bytes applies an update from byte slice
func (ydoc *YDoc) ApplyUpdateV2Bytes(update []byte, transactionOrigin interface{}, local bool) error {
	return ydoc.ApplyUpdateV2(bytes.NewReader(update), transactionOrigin, local)
}

// EncodeStateAsUpdateV2 encodes the document state as an update
func (ydoc *YDoc) EncodeStateAsUpdateV2(encodedTargetStateVector []byte) ([]byte, error) {
	encoder := NewUpdateEncoderV2()
	defer encoder.Close()

	var targetStateVector map[int64]int64
	if encodedTargetStateVector == nil {
		targetStateVector = make(map[int64]int64)
	} else {
		var err error
		targetStateVector, err = DecodeStateVector(bytes.NewReader(encodedTargetStateVector))
		if err != nil {
			return nil, err
		}
	}

	ydoc.WriteStateAsUpdate(encoder, targetStateVector)
	return encoder.ToArray(), nil
}

// EncodeStateVectorV2 encodes the state vector
func (ydoc *YDoc) EncodeStateVectorV2() ([]byte, error) {
	encoder := NewDSEncoderV2()
	defer encoder.Close()
	ydoc.WriteStateVector(encoder)
	return encoder.ToArray(), nil
}

// WriteStateAsUpdate writes the document state as an update
func (ydoc *YDoc) WriteStateAsUpdate(encoder contracts.IUpdateEncoder, targetStateVector map[int64]int64) error {
	return WriteClientsStructs(encoder, ydoc.store, targetStateVector)
}

// WriteStateVector writes the state vector
func (ydoc *YDoc) WriteStateVector(encoder contracts.IDSEncoder) error {
	return WriteStateVector(encoder, ydoc.store.GetStateVector())
}

// FindRootTypeKey finds the root type key for a given type
func (ydoc *YDoc) FindRootTypeKey(targetType contracts.IAbstractType) string {
	for key, yType := range ydoc.GetShare() {
		if yType == targetType {
			return key
		}
	}
	panic("type not found in share")
}

// Event handler methods
func (ydoc *YDoc) InvokeSubdocsChanged(loaded, added, removed map[contracts.IYDoc]struct{}) {
	if ydoc.subdocsChanged != nil {
		ydoc.subdocsChanged(loaded, added, removed)
	}
}

func (ydoc *YDoc) InvokeOnBeforeObserverCalls(transaction contracts.ITransaction) {
	if ydoc.beforeObserverCalls != nil {
		ydoc.beforeObserverCalls(transaction)
	}
}

func (ydoc *YDoc) InvokeAfterAllTransactions(transactions []contracts.ITransaction) {
	if ydoc.afterAllTransactions != nil {
		ydoc.afterAllTransactions(transactions)
	}
}

func (ydoc *YDoc) InvokeOnBeforeTransaction(transaction contracts.ITransaction) {
	if ydoc.beforeTransaction != nil {
		ydoc.beforeTransaction(transaction)
	}
}

func (ydoc *YDoc) InvokeOnAfterTransaction(transaction contracts.ITransaction) {
	if ydoc.afterTransaction != nil {
		ydoc.afterTransaction(transaction)
	}
}

func (ydoc *YDoc) InvokeOnAfterTransactionCleanup(transaction contracts.ITransaction) {
	if ydoc.afterTransactionCleanup != nil {
		ydoc.afterTransactionCleanup(transaction)
	}
}

func (ydoc *YDoc) InvokeBeforeAllTransactions() {
	if ydoc.beforeAllTransactions != nil {
		ydoc.beforeAllTransactions()
	}
}

func (ydoc *YDoc) InvokeDestroyed() {
	if ydoc.destroyed != nil {
		ydoc.destroyed()
	}
}

func (ydoc *YDoc) InvokeUpdateV2(transaction contracts.ITransaction) {
	if ydoc.updateV2 != nil {
		encoder := NewUpdateEncoderV2()
		hasContent := transaction.WriteUpdateMessageFromTransaction(encoder)
		if hasContent {
			ydoc.updateV2(encoder.ToArray(), transaction.GetOrigin(), transaction)
		}
	}
}

// CloneOptionsWithNewGuid creates a copy of options with a new GUID
func (ydoc *YDoc) CloneOptionsWithNewGuid() contracts.YDocOptions {
	newOpts := ydoc.opts
	newOpts.Guid = generateGUID()
	return newOpts
}
