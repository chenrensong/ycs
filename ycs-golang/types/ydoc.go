// ------------------------------------------------------------------------------
//  Copyright (c) Microsoft Corporation.  All rights reserved.
// ------------------------------------------------------------------------------

package types

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	"github.com/chenrensong/ygo/contracts"
	"github.com/google/uuid"
)

// YDoc represents a Yjs instance that handles the state of shared data
type YDoc struct {
	opts                *contracts.YDocOptions
	shouldLoad          bool
	transactionCleanups []contracts.ITransaction
	transaction         contracts.ITransaction
	subdocs             map[contracts.IYDoc]bool
	item                contracts.IStructItem
	share               map[string]contracts.IAbstractType
	clientID            int64
	store               contracts.IStructStore

	// Event handlers
	beforeObserverCalls     []func(contracts.ITransaction)
	beforeTransaction       []func(contracts.ITransaction)
	afterTransaction        []func(contracts.ITransaction)
	afterTransactionCleanup []func(contracts.ITransaction)
	beforeAllTransactions   []func()
	afterAllTransactions    []func([]contracts.ITransaction)
	updateV2                []func([]byte, interface{}, contracts.ITransaction)
	destroyed               []func()
	subdocsChanged          []func(map[contracts.IYDoc]bool, map[contracts.IYDoc]bool, map[contracts.IYDoc]bool)
}

// GenerateNewClientID generates a new random client ID
func GenerateNewClientID() int64 {
	max := big.NewInt(int64(^uint(0) >> 1)) // max int64
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	return n.Int64()
}

// NewYDoc creates a new YDoc instance
func NewYDoc(opts *contracts.YDocOptions) *YDoc {
	if opts == nil {
		opts = &contracts.YDocOptions{
			Guid:     "default-guid",
			Gc:       true,
			GcFilter: nil,
			AutoLoad: false,
			Meta:     make(map[string]string),
		}
	}

	doc := &YDoc{
		opts:                    opts,
		transactionCleanups:     make([]contracts.ITransaction, 0),
		clientID:                GenerateNewClientID(),
		share:                   make(map[string]contracts.IAbstractType),
		store:                   contracts.NewStructStore(),
		subdocs:                 make(map[contracts.IYDoc]bool),
		shouldLoad:              opts.AutoLoad,
		beforeObserverCalls:     make([]func(contracts.ITransaction), 0),
		beforeTransaction:       make([]func(contracts.ITransaction), 0),
		afterTransaction:        make([]func(contracts.ITransaction), 0),
		afterTransactionCleanup: make([]func(contracts.ITransaction), 0),
		beforeAllTransactions:   make([]func(), 0),
		afterAllTransactions:    make([]func([]contracts.ITransaction), 0),
		updateV2:                make([]func([]byte, interface{}, contracts.ITransaction), 0),
		destroyed:               make([]func(), 0),
		subdocsChanged:          make([]func(map[contracts.IYDoc]bool, map[contracts.IYDoc]bool, map[contracts.IYDoc]bool), 0),
	}

	return doc
}

// Opts returns the document options
func (yd *YDoc) Opts() *contracts.YDocOptions {
	return yd.opts
}

// Guid returns the document GUID
func (yd *YDoc) Guid() string {
	return yd.opts.Guid
}

// Gc returns whether garbage collection is enabled
func (yd *YDoc) Gc() bool {
	return yd.opts.Gc
}

// GcFilter returns the garbage collection filter function
func (yd *YDoc) GcFilter() func(contracts.IStructItem) bool {
	return yd.opts.GcFilter
}

// AutoLoad returns whether auto-loading is enabled
func (yd *YDoc) AutoLoad() bool {
	return yd.opts.AutoLoad
}

// Meta returns the document metadata
func (yd *YDoc) Meta() map[string]string {
	return yd.opts.Meta
}

// ShouldLoad returns whether the document should be loaded
func (yd *YDoc) ShouldLoad() bool {
	return yd.shouldLoad
}

// SetShouldLoad sets whether the document should be loaded
func (yd *YDoc) SetShouldLoad(shouldLoad bool) {
	yd.shouldLoad = shouldLoad
}

// TransactionCleanups returns the transaction cleanups list
func (yd *YDoc) TransactionCleanups() []contracts.ITransaction {
	return yd.transactionCleanups
}

// Transaction returns the current transaction
func (yd *YDoc) Transaction() contracts.ITransaction {
	return yd.transaction
}

// SetTransaction sets the current transaction
func (yd *YDoc) SetTransaction(transaction contracts.ITransaction) {
	yd.transaction = transaction
}

// Subdocs returns the subdocuments set
func (yd *YDoc) Subdocs() map[contracts.IYDoc]bool {
	return yd.subdocs
}

// Item returns the document item
func (yd *YDoc) Item() contracts.IStructItem {
	return yd.item
}

// SetItem sets the document item
func (yd *YDoc) SetItem(item contracts.IStructItem) {
	yd.item = item
}

// Share returns the shared types map
func (yd *YDoc) Share() map[string]contracts.IAbstractType {
	return yd.share
}

// SetShare sets the shared types map
func (yd *YDoc) SetShare(share map[string]contracts.IAbstractType) {
	yd.share = share
}

// ClientID returns the client ID
func (yd *YDoc) ClientID() int64 {
	return yd.clientID
}

// SetClientID sets the client ID
func (yd *YDoc) SetClientID(clientID int64) {
	yd.clientID = clientID
}

// Store returns the struct store
func (yd *YDoc) Store() contracts.IStructStore {
	return yd.store
}

// SetStore sets the struct store
func (yd *YDoc) SetStore(store contracts.IStructStore) {
	yd.store = store
}

// Load notifies the parent document to load data into this subdocument
func (yd *YDoc) Load() {
	item := yd.item
	if item != nil && !yd.shouldLoad {
		if parent, ok := item.Parent().(*AbstractType); ok {
			parent.Doc().Transact(func(tr contracts.ITransaction) {
				tr.SubdocsLoaded().Add(yd)
			}, nil, true)
		}
	}
	yd.shouldLoad = true
}

// CreateSnapshot creates a snapshot of the current document state
func (yd *YDoc) CreateSnapshot() contracts.ISnapshot {
	return contracts.NewSnapshot(contracts.NewDeleteSet(yd.store), yd.store.GetStateVector())
}

// GetSubdocGuids returns the GUIDs of all subdocuments
func (yd *YDoc) GetSubdocGuids() []string {
	guids := make([]string, 0, len(yd.subdocs))
	for sd := range yd.subdocs {
		guids = append(guids, sd.Guid())
	}
	return guids
}

// Destroy destroys the document and all subdocuments
func (yd *YDoc) Destroy() {
	for sd := range yd.subdocs {
		sd.Destroy()
	}

	item := yd.item
	if item != nil {
		yd.item = nil
		if contentDoc, ok := item.Content().(*contracts.ContentDoc); ok {
			if item.Deleted() {
				contentDoc.SetDoc(nil)
			} else {
				newOpts := contentDoc.Opts()
				newOpts.Guid = yd.Guid()
				newDoc := NewYDoc(newOpts)
				contentDoc.SetDoc(newDoc)
				newDoc.SetItem(item)
			}

			if parent, ok := item.Parent().(*AbstractType); ok {
				parent.Doc().Transact(func(tr contracts.ITransaction) {
					if !item.Deleted() {
						tr.SubdocsAdded().Add(contentDoc.Doc())
					}
					tr.SubdocsRemoved().Add(yd)
				}, nil, true)
			}
		}
	}

	yd.invokeDestroyed()
}

// Transact executes a function as a transaction
func (yd *YDoc) Transact(fn func(contracts.ITransaction), origin interface{}, local bool) {
	initialCall := false
	if yd.transaction == nil {
		initialCall = true
		yd.transaction = contracts.NewTransaction(yd, origin, local)
		yd.transactionCleanups = append(yd.transactionCleanups, yd.transaction)
		if len(yd.transactionCleanups) == 1 {
			yd.invokeBeforeAllTransactions()
		}
		yd.invokeOnBeforeTransaction(yd.transaction)
	}

	defer func() {
		if initialCall && len(yd.transactionCleanups) > 0 && yd.transactionCleanups[0] == yd.transaction {
			// The first transaction ended, now process observer calls
			contracts.CleanupTransactions(yd.transactionCleanups, 0)
		}
	}()

	fn(yd.transaction)
}

// GetArray returns a YArray with the given name
func (yd *YDoc) GetArray(name string) contracts.IYArray {
	if name == "" {
		name = "array"
	}
	return yd.Get(name, func() contracts.IAbstractType { return NewYArray() }).(contracts.IYArray)
}

// GetMap returns a YMap with the given name
func (yd *YDoc) GetMap(name string) contracts.IYMap {
	if name == "" {
		name = "map"
	}
	return yd.Get(name, func() contracts.IAbstractType { return NewYMapWithEntries(nil) }).(contracts.IYMap)
}

// GetText returns a YText with the given name
func (yd *YDoc) GetText(name string) contracts.IYText {
	if name == "" {
		name = "text"
	}
	return yd.Get(name, func() contracts.IAbstractType { return NewYText() }).(contracts.IYText)
}

// Get returns a shared type with the given name, creating it if it doesn't exist
func (yd *YDoc) Get(name string, constructor func() contracts.IAbstractType) contracts.IAbstractType {
	if typ, exists := yd.share[name]; exists {
		return typ
	}

	typ := constructor()
	typ.Integrate(yd, nil)
	yd.share[name] = typ
	return typ
}

// ApplyUpdateV2 reads and applies a document update from a stream
func (yd *YDoc) ApplyUpdateV2(input io.Reader, transactionOrigin interface{}, local bool) error {
	yd.Transact(func(tr contracts.ITransaction) {
		structDecoder := contracts.NewUpdateDecoderV2(input)
		defer structDecoder.Close()

		contracts.ReadStructs(structDecoder, tr, yd.store)
		yd.store.ReadAndApplyDeleteSet(structDecoder, tr)
	}, transactionOrigin, local)
	return nil
}

// ApplyUpdateV2Bytes reads and applies a document update from bytes
func (yd *YDoc) ApplyUpdateV2Bytes(update []byte, transactionOrigin interface{}, local bool) error {
	input := bytes.NewReader(update)
	return yd.ApplyUpdateV2(input, transactionOrigin, local)
}

// EncodeStateAsUpdateV2 writes all the document as a single update message
func (yd *YDoc) EncodeStateAsUpdateV2(encodedTargetStateVector []byte) ([]byte, error) {
	encoder := contracts.NewUpdateEncoderV2()
	defer encoder.Close()

	var targetStateVector map[int64]int64
	if encodedTargetStateVector == nil {
		targetStateVector = make(map[int64]int64)
	} else {
		var err error
		targetStateVector, err = contracts.DecodeStateVector(bytes.NewReader(encodedTargetStateVector))
		if err != nil {
			return nil, err
		}
	}

	yd.WriteStateAsUpdate(encoder, targetStateVector)
	return encoder.ToArray(), nil
}

// EncodeStateVectorV2 encodes the state vector
func (yd *YDoc) EncodeStateVectorV2() ([]byte, error) {
	encoder := contracts.NewDSEncoderV2()
	defer encoder.Close()

	yd.WriteStateVector(encoder)
	return encoder.ToArray(), nil
}

// WriteStateAsUpdate writes all the document as a single update message
func (yd *YDoc) WriteStateAsUpdate(encoder contracts.IUpdateEncoder, targetStateVector map[int64]int64) {
	contracts.WriteClientsStructs(encoder, yd.store, targetStateVector)
	deleteSet := contracts.NewDeleteSet(yd.store)
	deleteSet.Write(encoder)
}

// WriteStateVector writes the state vector
func (yd *YDoc) WriteStateVector(encoder contracts.IDSEncoder) {
	contracts.WriteStateVector(encoder, yd.store.GetStateVector())
}

// CloneOptionsWithNewGuid creates a copy of the options with a new GUID
func (yd *YDoc) CloneOptionsWithNewGuid() *contracts.YDocOptions {
	newOpts := *yd.opts // Copy struct
	newOpts.Guid = uuid.New().String()
	return &newOpts
}

// FindRootTypeKey finds the key for the given root type
func (yd *YDoc) FindRootTypeKey(typ contracts.IAbstractType) (string, error) {
	for key, value := range yd.share {
		if value == typ {
			return key, nil
		}
	}
	return "", fmt.Errorf("root type not found")
}

// Event handler methods
func (yd *YDoc) OnBeforeObserverCalls(handler func(contracts.ITransaction)) {
	yd.beforeObserverCalls = append(yd.beforeObserverCalls, handler)
}

func (yd *YDoc) OnBeforeTransaction(handler func(contracts.ITransaction)) {
	yd.beforeTransaction = append(yd.beforeTransaction, handler)
}

func (yd *YDoc) OnAfterTransaction(handler func(contracts.ITransaction)) {
	yd.afterTransaction = append(yd.afterTransaction, handler)
}

func (yd *YDoc) OnAfterTransactionCleanup(handler func(contracts.ITransaction)) {
	yd.afterTransactionCleanup = append(yd.afterTransactionCleanup, handler)
}

func (yd *YDoc) OnBeforeAllTransactions(handler func()) {
	yd.beforeAllTransactions = append(yd.beforeAllTransactions, handler)
}

func (yd *YDoc) OnAfterAllTransactions(handler func([]contracts.ITransaction)) {
	yd.afterAllTransactions = append(yd.afterAllTransactions, handler)
}

func (yd *YDoc) OnUpdateV2(handler func([]byte, interface{}, contracts.ITransaction)) {
	yd.updateV2 = append(yd.updateV2, handler)
}

func (yd *YDoc) OnDestroyed(handler func()) {
	yd.destroyed = append(yd.destroyed, handler)
}

func (yd *YDoc) OnSubdocsChanged(handler func(map[contracts.IYDoc]bool, map[contracts.IYDoc]bool, map[contracts.IYDoc]bool)) {
	yd.subdocsChanged = append(yd.subdocsChanged, handler)
}

// Event invocation methods
func (yd *YDoc) invokeSubdocsChanged(loaded, added, removed map[contracts.IYDoc]bool) {
	for _, handler := range yd.subdocsChanged {
		handler(loaded, added, removed)
	}
}

func (yd *YDoc) invokeOnBeforeObserverCalls(transaction contracts.ITransaction) {
	for _, handler := range yd.beforeObserverCalls {
		handler(transaction)
	}
}

func (yd *YDoc) invokeAfterAllTransactions(transactions []contracts.ITransaction) {
	for _, handler := range yd.afterAllTransactions {
		handler(transactions)
	}
}

func (yd *YDoc) invokeOnBeforeTransaction(transaction contracts.ITransaction) {
	for _, handler := range yd.beforeTransaction {
		handler(transaction)
	}
}

func (yd *YDoc) invokeOnAfterTransaction(transaction contracts.ITransaction) {
	for _, handler := range yd.afterTransaction {
		handler(transaction)
	}
}

func (yd *YDoc) invokeOnAfterTransactionCleanup(transaction contracts.ITransaction) {
	for _, handler := range yd.afterTransactionCleanup {
		handler(transaction)
	}
}

func (yd *YDoc) invokeBeforeAllTransactions() {
	for _, handler := range yd.beforeAllTransactions {
		handler()
	}
}

func (yd *YDoc) invokeDestroyed() {
	for _, handler := range yd.destroyed {
		handler()
	}
}

func (yd *YDoc) invokeUpdateV2(transaction contracts.ITransaction) {
	if len(yd.updateV2) > 0 {
		encoder := contracts.NewUpdateEncoderV2()
		defer encoder.Close()

		hasContent := transaction.WriteUpdateMessageFromTransaction(encoder)
		if hasContent {
			data := encoder.ToArray()
			for _, handler := range yd.updateV2 {
				handler(data, transaction.Origin(), transaction)
			}
		}
	}
}
