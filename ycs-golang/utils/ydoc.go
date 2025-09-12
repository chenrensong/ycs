// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"bytes"
	"math/rand"
	"time"
	"ycs-golang/structs"
	"ycs-golang/types"
)

// YDocOptions represents options for YDoc
type YDocOptions struct {
	Gc       bool
	GcFilter func(*structs.Item) bool
	Guid     string
	Meta     map[string]string
	AutoLoad bool
}

// Clone clones the YDocOptions
func (opts *YDocOptions) Clone() *YDocOptions {
	clone := &YDocOptions{
		Gc:       opts.Gc,
		Guid:     opts.Guid,
		AutoLoad: opts.AutoLoad,
	}
	
	if opts.Meta != nil {
		clone.Meta = make(map[string]string)
		for k, v := range opts.Meta {
			clone.Meta[k] = v
		}
	}
	
	if opts.GcFilter != nil {
		clone.GcFilter = opts.GcFilter
	}
	
	return clone
}

// Write writes the options to an encoder
func (opts *YDocOptions) Write(encoder IUpdateEncoder, offset int) {
	dict := make(map[string]interface{})
	dict["gc"] = opts.Gc
	dict["guid"] = opts.Guid
	dict["autoLoad"] = opts.AutoLoad

	if opts.Meta != nil {
		dict["meta"] = opts.Meta
	}

	// Note: encoder.WriteAny needs to be implemented
	// encoder.WriteAny(dict)
}

// Read reads options from a decoder
func ReadYDocOptions(decoder IUpdateDecoder) *YDocOptions {
	// Note: decoder.ReadAny needs to be implemented
	// dict := decoder.ReadAny().(map[string]interface{})
	
	result := &YDocOptions{
		Gc:       true, // Default value
		Guid:     generateGUID(), // Default value
		AutoLoad: false, // Default value
	}
	
	// Placeholder for reading from decoder
	// if val, exists := dict["gc"]; exists {
	//     result.Gc = val.(bool)
	// }
	// 
	// if val, exists := dict["guid"]; exists {
	//     result.Guid = val.(string)
	// } else {
	//     result.Guid = generateGUID()
	// }
	// 
	// if val, exists := dict["meta"]; exists {
	//     result.Meta = val.(map[string]string)
	// }
	// 
	// if val, exists := dict["autoLoad"]; exists {
	//     result.AutoLoad = val.(bool)
	// }
	
	return result
}

// DefaultGcFilter is the default garbage collection filter
func DefaultGcFilter(item *structs.Item) bool {
	return true
}

// YDoc represents a Yjs document
type YDoc struct {
	opts                  *YDocOptions
	ClientId              int64
	Store                 *StructStore
	transactionCleanups   []*Transaction
	transaction           *Transaction
	Subdocs               map[*YDoc]bool
	item                  *structs.Item
	share                 map[string]*types.AbstractType
	ShouldLoad            bool
	
	// Event handlers
	BeforeObserverCalls        []func(*Transaction)
	BeforeTransaction          []func(*Transaction)
	AfterTransaction           []func(*Transaction)
	AfterTransactionCleanup    []func(*Transaction)
	BeforeAllTransactions      []func()
	AfterAllTransactions       []func([]*Transaction)
	UpdateV2                   []func([]byte, interface{}, *Transaction)
	Destroyed                  []func()
	SubdocsChanged             []func(map[*YDoc]bool, map[*YDoc]bool, map[*YDoc]bool)
}

// NewYDoc creates a new YDoc
func NewYDoc(opts *YDocOptions) *YDoc {
	if opts == nil {
		opts = &YDocOptions{
			Gc:       true,
			GcFilter: DefaultGcFilter,
			Guid:     generateGUID(),
			Meta:     nil,
			AutoLoad: false,
		}
	}
	
	doc := &YDoc{
		opts:                opts,
		transactionCleanups: make([]*Transaction, 0),
		share:               make(map[string]*types.AbstractType),
		Store:               NewStructStore(),
		Subdocs:             make(map[*YDoc]bool),
		ShouldLoad:          opts.AutoLoad,
	}
	
	doc.ClientId = GenerateNewClientId()
	return doc
}

// Guid returns the document's GUID
func (doc *YDoc) Guid() string {
	return doc.opts.Guid
}

// Gc returns whether garbage collection is enabled
func (doc *YDoc) Gc() bool {
	return doc.opts.Gc
}

// GcFilter returns the garbage collection filter
func (doc *YDoc) GcFilter() func(*structs.Item) bool {
	return doc.opts.GcFilter
}

// AutoLoad returns whether auto load is enabled
func (doc *YDoc) AutoLoad() bool {
	return doc.opts.AutoLoad
}

// Meta returns the document's metadata
func (doc *YDoc) Meta() map[string]string {
	return doc.opts.Meta
}

// GenerateNewClientId generates a new client ID
func GenerateNewClientId() int64 {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	return int64(rand.Intn(2147483647)) // Max value for 32-bit integer
}

// generateGUID generates a new GUID
func generateGUID() string {
	// This is a simplified GUID generation
	// In a real implementation, you might want to use a proper GUID library
	return "guid-" + time.Now().String()
}

// Load loads the document
func (doc *YDoc) Load() {
	item := doc.item
	if item != nil && !doc.ShouldLoad {
		if parent, ok := item.Parent.(*types.AbstractType); ok {
			parent.Doc.Transact(func(tr *Transaction) {
				tr.SubdocsLoaded[doc] = true
			}, nil, true)
		}
	}
	doc.ShouldLoad = true
}

// CreateSnapshot creates a snapshot of the document
func (doc *YDoc) CreateSnapshot() *Snapshot {
	return NewSnapshot(NewDeleteSetFromStructStore(doc.Store), doc.Store.GetStateVector())
}

// GetSubdocGuids gets the GUIDs of subdocuments
func (doc *YDoc) GetSubdocGuids() []string {
	guids := make([]string, 0, len(doc.Subdocs))
	for subdoc := range doc.Subdocs {
		guids = append(guids, subdoc.Guid())
	}
	return guids
}

// Destroy destroys the document
func (doc *YDoc) Destroy() {
	for subdoc := range doc.Subdocs {
		subdoc.Destroy()
	}

	item := doc.item
	if item != nil {
		doc.item = nil
		// Note: ContentDoc type needs to be implemented
		// content := item.Content.(*ContentDoc)

		if item.Deleted {
			// if content != nil {
			//     content.Doc = nil
			// }
		} else {
			// Debug.Assert(content != nil)
			// newOpts := content.Opts
			// newOpts.Guid = doc.Guid()

			// content.Doc = NewYDoc(newOpts)
			// content.Doc.item = item
		}

		if parent, ok := item.Parent.(*types.AbstractType); ok {
			parent.Doc.Transact(func(tr *Transaction) {
				// if !item.Deleted {
				//     Debug.Assert(content != nil)
				//     tr.SubdocsAdded[content.Doc] = true
				// }

				tr.SubdocsRemoved[doc] = true
			}, nil, true)
		}
	}

	doc.InvokeDestroyed()
}

// Transact executes a function as a transaction
func (doc *YDoc) Transact(fun func(*Transaction), origin interface{}, local bool) {
	initialCall := false
	if doc.transaction == nil {
		initialCall = true
		doc.transaction = NewTransaction(doc, origin, local)
		doc.transactionCleanups = append(doc.transactionCleanups, doc.transaction)
		if len(doc.transactionCleanups) == 1 {
			doc.InvokeBeforeAllTransactions()
		}

		doc.InvokeOnBeforeTransaction(doc.transaction)
	}

	defer func() {
		if initialCall && len(doc.transactionCleanups) > 0 && doc.transactionCleanups[0] == doc.transaction {
			// The first transaction ended, now process observer calls.
			// Observer call may create new transactions for which we need to call the observers and do cleanup.
			// We don't want to nest these calls, so we execute these calls one after another.
			// Also we need to ensure that all cleanups are called, even if the observers throw errors.
			CleanupTransactions(doc.transactionCleanups, 0)
		}
	}()

	fun(doc.transaction)
}

// GetArray gets an array from the document
func (doc *YDoc) GetArray(name string) *types.YArray {
	return doc.Get(name).(*types.YArray)
}

// GetMap gets a map from the document
func (doc *YDoc) GetMap(name string) *types.YMap {
	return doc.Get(name).(*types.YMap)
}

// GetText gets a text from the document
func (doc *YDoc) GetText(name string) *types.YText {
	return doc.Get(name).(*types.YText)
}

// Get gets a type from the document
func (doc *YDoc) Get(name string) interface{} {
	if typ, exists := doc.share[name]; exists {
		return typ
	}

	// Create a new abstract type
	// Note: This is a simplified implementation
	// In a real implementation, you would need to handle different types
	typ := types.NewAbstractType()
	typ.Integrate(doc, nil)
	doc.share[name] = typ
	return typ
}

// ApplyUpdateV2 applies an update from a stream
func (doc *YDoc) ApplyUpdateV2(input []byte, transactionOrigin interface{}, local bool) {
	doc.Transact(func(tr *Transaction) {
		// Note: UpdateDecoderV2 needs to be implemented
		// structDecoder := NewUpdateDecoderV2(bytes.NewReader(input))
		// defer structDecoder.Dispose()
		
		// EncodingUtils.ReadStructs(structDecoder, tr, doc.Store)
		// doc.Store.ReadAndApplyDeleteSet(structDecoder, tr)
	}, transactionOrigin, local)
}

// EncodeStateAsUpdateV2 encodes the document state as an update
func (doc *YDoc) EncodeStateAsUpdateV2(encodedTargetStateVector []byte) []byte {
	// Note: UpdateEncoderV2 needs to be implemented
	// encoder := NewUpdateEncoderV2()
	// defer encoder.Dispose()
	
	// var targetStateVector map[int64]int64
	// if encodedTargetStateVector == nil {
	//     targetStateVector = make(map[int64]int64)
	// } else {
	//     // Note: EncodingUtils.DecodeStateVector needs to be implemented
	//     // targetStateVector = EncodingUtils.DecodeStateVector(bytes.NewReader(encodedTargetStateVector))
	// }
	// 
	// doc.WriteStateAsUpdate(encoder, targetStateVector)
	// return encoder.ToArray()
	
	return []byte{} // Placeholder
}

// EncodeStateVectorV2 encodes the document state vector
func (doc *YDoc) EncodeStateVectorV2() []byte {
	// Note: DSEncoderV2 needs to be implemented
	// encoder := NewDSEncoderV2()
	// defer encoder.Dispose()
	
	// doc.WriteStateVector(encoder)
	// return encoder.ToArray()
	
	return []byte{} // Placeholder
}

// WriteStateAsUpdate writes the document state as an update
func (doc *YDoc) WriteStateAsUpdate(encoder IUpdateEncoder, targetStateVector map[int64]int64) {
	// Note: EncodingUtils.WriteClientsStructs needs to be implemented
	// EncodingUtils.WriteClientsStructs(encoder, doc.Store, targetStateVector)
	// NewDeleteSetFromStructStore(doc.Store).Write(encoder)
}

// WriteStateVector writes the document state vector
func (doc *YDoc) WriteStateVector(encoder IDSDecoder) {
	// Note: EncodingUtils.WriteStateVector needs to be implemented
	// EncodingUtils.WriteStateVector(encoder, doc.Store.GetStateVector())
}

// InvokeSubdocsChanged invokes the subdocs changed event
func (doc *YDoc) InvokeSubdocsChanged(loaded, added, removed map[*YDoc]bool) {
	for _, handler := range doc.SubdocsChanged {
		handler(loaded, added, removed)
	}
}

// InvokeOnBeforeObserverCalls invokes the before observer calls event
func (doc *YDoc) InvokeOnBeforeObserverCalls(transaction *Transaction) {
	for _, handler := range doc.BeforeObserverCalls {
		handler(transaction)
	}
}

// InvokeAfterAllTransactions invokes the after all transactions event
func (doc *YDoc) InvokeAfterAllTransactions(transactions []*Transaction) {
	for _, handler := range doc.AfterAllTransactions {
		handler(transactions)
	}
}

// InvokeOnBeforeTransaction invokes the before transaction event
func (doc *YDoc) InvokeOnBeforeTransaction(transaction *Transaction) {
	for _, handler := range doc.BeforeTransaction {
		handler(transaction)
	}
}

// InvokeOnAfterTransaction invokes the after transaction event
func (doc *YDoc) InvokeOnAfterTransaction(transaction *Transaction) {
	for _, handler := range doc.AfterTransaction {
		handler(transaction)
	}
}

// InvokeOnAfterTransactionCleanup invokes the after transaction cleanup event
func (doc *YDoc) InvokeOnAfterTransactionCleanup(transaction *Transaction) {
	for _, handler := range doc.AfterTransactionCleanup {
		handler(transaction)
	}
}

// InvokeBeforeAllTransactions invokes the before all transactions event
func (doc *YDoc) InvokeBeforeAllTransactions() {
	for _, handler := range doc.BeforeAllTransactions {
		handler()
	}
}

// InvokeDestroyed invokes the destroyed event
func (doc *YDoc) InvokeDestroyed() {
	for _, handler := range doc.Destroyed {
		handler()
	}
}

// InvokeUpdateV2 invokes the update V2 event
func (doc *YDoc) InvokeUpdateV2(transaction *Transaction) {
	for _, handler := range doc.UpdateV2 {
		// Note: UpdateEncoderV2 needs to be implemented
		// encoder := NewUpdateEncoderV2()
		// defer encoder.Dispose()
		
		// hasContent := transaction.WriteUpdateMessageFromTransaction(encoder)
		// if hasContent {
		//     handler(encoder.ToArray(), transaction.Origin, transaction)
		// }
	}
}

// CloneOptionsWithNewGuid clones the options with a new GUID
func (doc *YDoc) CloneOptionsWithNewGuid() *YDocOptions {
	newOpts := doc.opts.Clone()
	newOpts.Guid = generateGUID()
	return newOpts
}

// FindRootTypeKey finds the root type key
func (doc *YDoc) FindRootTypeKey(typ *types.AbstractType) string {
	for key, value := range doc.share {
		if value == typ {
			return key
		}
	}
	
	panic("Type not found")
}