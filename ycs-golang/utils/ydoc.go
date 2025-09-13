package utils

import (
	"bytes"
	"errors"
	"math/rand"
	"sync"
	"time"

	encoding "github.com/chenrensong/ygo/lib0/encoding"
	structs "github.com/chenrensong/ygo/structs"
	types "github.com/chenrensong/ygo/types"
)

var (
	ErrTypeMismatch     = errors.New("type with this name already defined with different constructor")
	ErrUnknownType      = errors.New("unknown type")
	ErrInvalidOperation = errors.New("invalid operation")
)

// YDocOptions contains configuration options for a YDoc
type YDocOptions struct {
	GC       bool
	GCFilter func(item *structs.Item) bool
	GUID     string
	Meta     map[string]string
	AutoLoad bool
}

// Clone creates a deep copy of YDocOptions
func (o *YDocOptions) Clone() *YDocOptions {
	meta := make(map[string]string)
	for k, v := range o.Meta {
		meta[k] = v
	}
	return &YDocOptions{
		GC:       o.GC,
		GCFilter: o.GCFilter,
		GUID:     o.GUID,
		Meta:     meta,
		AutoLoad: o.AutoLoad,
	}
}

// YDoc represents a Yjs document that handles the state of shared data
type YDoc struct {
	opts *YDocOptions

	mu                  sync.RWMutex
	transaction         *Transaction
	transactionCleanups []*Transaction
	item                *structs.Item
	share               map[string]*types.AbstractType
	clientID            uint64
	store               *structs.StructStore
	subdocs             map[*YDoc]struct{}
	shouldLoad          bool

	// Event handlers
	beforeObserverCalls     func(*Transaction)
	beforeTransaction       func(*Transaction)
	afterTransaction        func(*Transaction)
	afterTransactionCleanup func(*Transaction)
	beforeAllTransactions   func()
	afterAllTransactions    func([]*Transaction)
	updateV2                func([]byte, interface{}, *Transaction)
	destroyed               func()
	subdocsChanged          func(map[*YDoc]struct{}, map[*YDoc]struct{}, map[*YDoc]struct{})
}

// NewYDoc creates a new YDoc instance
func NewYDoc(opts *YDocOptions) *YDoc {
	if opts == nil {
		opts = &YDocOptions{
			GC:       true,
			GUID:     newGUID(),
			AutoLoad: false,
		}
	}

	return &YDoc{
		opts:                opts,
		clientID:            generateNewClientID(),
		share:               make(map[string]*types.AbstractType),
		store:               structs.NewStructStore(),
		subdocs:             make(map[*YDoc]struct{}),
		shouldLoad:          opts.AutoLoad,
		transactionCleanups: make([]*Transaction, 0),
	}
}

func newGUID() string {
	// Implement GUID generation
	return ""
}

func generateNewClientID() uint64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint64()
}

// Load requests to load data into this subdocument
func (d *YDoc) Load() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.item != nil && !d.shouldLoad {
		if parent, ok := d.item.Parent.(*types.AbstractType); ok {
			parent.Doc().Transact(func(tr *Transaction) {
				tr.SubdocsLoaded[d] = struct{}{}
			}, nil, true)
		}
	}
	d.shouldLoad = true
}

// CreateSnapshot creates a snapshot of the current document state
func (d *YDoc) CreateSnapshot() *Snapshot {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return NewSnapshot(structs.NewDeleteSet(d.store), d.store.GetStateVector())
}

// GetSubdocGUIDs returns GUIDs of all subdocuments
func (d *YDoc) GetSubdocGUIDs() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	guids := make([]string, 0, len(d.subdocs))
	for subdoc := range d.subdocs {
		guids = append(guids, subdoc.opts.GUID)
	}
	return guids
}

// Destroy destroys the document and all its subdocuments
func (d *YDoc) Destroy() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Destroy all subdocs first
	for subdoc := range d.subdocs {
		subdoc.Destroy()
	}

	if d.item != nil {
		content, ok := d.item.Content.(*types.ContentDoc)
		if d.item.Deleted {
			if ok {
				content.Doc = nil
			}
		} else {
			if !ok {
				panic("invalid content type")
			}
			newOpts := content.Opts.Clone()
			newOpts.GUID = d.opts.GUID
			content.Doc = NewYDoc(newOpts)
			content.Doc.item = d.item
		}

		if parent, ok := d.item.Parent.(*types.AbstractType); ok {
			parent.Doc().Transact(func(tr *Transaction) {
				if !d.item.Deleted {
					tr.SubdocsAdded[content.Doc] = struct{}{}
				}
				tr.SubdocsRemoved[d] = struct{}{}
			}, nil, true)
		}
	}

	d.invokeDestroyed()
}

// Transact executes a function as a transaction
func (d *YDoc) Transact(fun func(*Transaction), origin interface{}, local bool) {
	d.mu.Lock()
	initialCall := false
	if d.transaction == nil {
		initialCall = true
		d.transaction = NewTransaction(d, origin, local)
		d.transactionCleanups = append(d.transactionCleanups, d.transaction)
		if len(d.transactionCleanups) == 1 {
			d.invokeBeforeAllTransactions()
		}
		d.invokeBeforeTransaction(d.transaction)
	}
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		if initialCall && len(d.transactionCleanups) > 0 && d.transactionCleanups[0] == d.transaction {
			CleanupTransactions(d.transactionCleanups, 0)
		}
	}()

	fun(d.transaction)
}

// GetArray returns or creates a YArray with the given name
func (d *YDoc) GetArray(name string) *types.YArray {
	return d.Get(name).(*types.YArray)
}

// GetMap returns or creates a YMap with the given name
func (d *YDoc) GetMap(name string) *types.YMap {
	return d.Get(name).(*types.YMap)
}

// GetText returns or creates a YText with the given name
func (d *YDoc) GetText(name string) *types.YText {
	return d.Get(name).(*types.YText)
}

// Get returns or creates a shared type with the given name
func (d *YDoc) Get(name string) types.AbstractType {
	d.mu.Lock()
	defer d.mu.Unlock()

	typ, exists := d.share[name]
	if !exists {
		// Default to YMap if no type specified
		typ = types.NewYMap()
		typ.Integrate(d, nil)
		d.share[name] = typ
		return typ
	}

	// Handle type conversion if needed
	switch t := typ.(type) {
	case *types.YArray:
		return t
	case *types.YMap:
		return t
	case *types.YText:
		return t
	default:
		// If the existing type is just AbstractType, we can convert it
		if _, ok := typ.(*types.AbstractType); ok {
			newType := types.NewYMap() // Default to YMap
			newType.SetMap(t.Map())
			newType.SetStart(t.Start())
			newType.SetLength(t.Length())
			d.share[name] = newType
			newType.Integrate(d, nil)
			return newType
		}
		panic(ErrTypeMismatch)
	}
}

// ApplyUpdateV2 applies an update to the document
func (d *YDoc) ApplyUpdateV2(update []byte, origin interface{}, local bool) {
	d.Transact(func(tr *Transaction) {
		decoder, err := NewUpdateDecoderV2(bytes.NewReader(update), false)
		if err != nil {
			panic(err) // Handle error appropriately in production code
		}
		ReadStructs(decoder, tr, d.store)
		d.store.ReadAndApplyDeleteSet(decoder, tr)
	}, origin, local)
}

// EncodeStateAsUpdateV2 encodes the document state as an update
func (d *YDoc) EncodeStateAsUpdateV2(encodedTargetStateVector []byte) []byte {
	d.mu.RLock()
	defer d.mu.RUnlock()

	encoder := encoding.NewUpdateEncoderV2()
	var targetStateVector map[uint64]uint64
	if encodedTargetStateVector != nil {
		targetStateVector = encoding.DecodeStateVector(bytes.NewReader(encodedTargetStateVector))
	} else {
		targetStateVector = make(map[uint64]uint64)
	}
	d.writeStateAsUpdate(encoder, targetStateVector)
	return encoder.ToArray()
}

// EncodeStateVectorV2 encodes the state vector
func (d *YDoc) EncodeStateVectorV2() []byte {
	d.mu.RLock()
	defer d.mu.RUnlock()

	encoder := encoding.NewDSEncoderV2()
	d.writeStateVector(encoder)
	return encoder.ToArray()
}

// writeStateAsUpdate writes the document state as an update
func (d *YDoc) writeStateAsUpdate(encoder encoding.UpdateEncoder, targetStateVector map[uint64]uint64) {
	encoding.WriteClientsStructs(encoder, d.store, targetStateVector)
	structs.NewDeleteSet(d.store).Write(encoder)
}

// writeStateVector writes the state vector
func (d *YDoc) writeStateVector(encoder encoding.DSEncoder) {
	encoding.WriteStateVector(encoder, d.store.GetStateVector())
}

// cloneOptionsWithNewGUID clones the options with a new GUID
func (d *YDoc) cloneOptionsWithNewGUID() *YDocOptions {
	newOpts := d.opts.Clone()
	newOpts.GUID = newGUID()
	return newOpts
}

// findRootTypeKey finds the root type key for a type
func (d *YDoc) findRootTypeKey(typ *types.AbstractType) string {
	for name, t := range d.share {
		if typ.Equals(t) {
			return name
		}
	}
	panic(ErrUnknownType)
}

// Event invocation methods
func (d *YDoc) invokeSubdocsChanged(loaded, added, removed map[*YDoc]struct{}) {
	if d.subdocsChanged != nil {
		d.subdocsChanged(loaded, added, removed)
	}
}

func (d *YDoc) invokeBeforeObserverCalls(tr *Transaction) {
	if d.beforeObserverCalls != nil {
		d.beforeObserverCalls(tr)
	}
}

func (d *YDoc) invokeAfterAllTransactions(transactions []*Transaction) {
	if d.afterAllTransactions != nil {
		d.afterAllTransactions(transactions)
	}
}

func (d *YDoc) invokeBeforeTransaction(tr *Transaction) {
	if d.beforeTransaction != nil {
		d.beforeTransaction(tr)
	}
}

func (d *YDoc) invokeAfterTransaction(tr *Transaction) {
	if d.afterTransaction != nil {
		d.afterTransaction(tr)
	}
}

func (d *YDoc) invokeAfterTransactionCleanup(tr *Transaction) {
	if d.afterTransactionCleanup != nil {
		d.afterTransactionCleanup(tr)
	}
}

func (d *YDoc) invokeBeforeAllTransactions() {
	if d.beforeAllTransactions != nil {
		d.beforeAllTransactions()
	}
}

func (d *YDoc) invokeDestroyed() {
	if d.destroyed != nil {
		d.destroyed()
	}
}

func (d *YDoc) invokeUpdateV2(tr *Transaction) {
	if d.updateV2 != nil {
		encoder := encoding.NewUpdateEncoderV2()
		if hasContent := tr.WriteUpdateMessageFromTransaction(encoder); hasContent {
			d.updateV2(encoder.ToArray(), tr.Origin, tr)
		}
	}
}
