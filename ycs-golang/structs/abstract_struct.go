package structs

// ITransaction defines the interface for transaction operations
// This interface helps avoid circular dependencies between structs and utils packages
type ITransaction interface {
	// Doc returns the document associated with this transaction
	Doc() interface{}

	// Origin returns the origin of this transaction
	Origin() interface{}

	// Local returns whether this transaction is local
	Local() bool

	// DeleteSet returns the delete set for this transaction
	DeleteSet() interface{}
}

// IStructStore defines the interface for struct store operations
// This interface helps avoid circular dependencies between structs and utils packages
type IStructStore interface {
	// GetState returns the current state for a client
	GetState(client uint64) uint64

	// GetStateVector returns the current state vector
	GetStateVector() map[uint64]uint64
}

// AbstractStruct is the base struct for all Yjs data structures.
type AbstractStruct interface {
	// ID returns the struct's identifier
	ID() ID

	// Length returns the length of the struct
	Length() int

	// SetLength updates the length of the struct
	SetLength(length int)

	// Deleted returns whether the struct is deleted
	Deleted() bool

	// MergeWith attempts to merge this struct with another
	MergeWith(right AbstractStruct) bool

	// Delete marks the struct as deleted
	Delete(transaction ITransaction)

	// Integrate integrates the struct into the document
	Integrate(transaction ITransaction, offset int)

	// GetMissing gets missing structs from the store
	GetMissing(transaction ITransaction, store IStructStore) *uint64

	// Write encodes the struct to an encoder
	Write(encoder IUpdateEncoder, offset int) error
}

// BaseStruct provides common functionality for all struct implementations
type BaseStruct struct {
	id     ID
	length int
}

// NewBaseStruct creates a new BaseStruct instance
func NewBaseStruct(id ID, length int) *BaseStruct {
	if length < 0 {
		panic("length must be non-negative")
	}
	return &BaseStruct{
		id:     id,
		length: length,
	}
}

// ID returns the struct's identifier
func (s *BaseStruct) ID() ID {
	return s.id
}

// Length returns the length of the struct
func (s *BaseStruct) Length() int {
	return s.length
}

// SetLength updates the length of the struct
func (s *BaseStruct) SetLength(length int) {
	if length < 0 {
		panic("length must be non-negative")
	}
	s.length = length
}
