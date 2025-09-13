package structs

import "fmt"

// ID represents a unique identifier for CRDT operations
type ID struct {
	Client uint64
	Clock  int
}

// NewID creates a new ID
func NewID(client uint64, clock int) *ID {
	return &ID{
		Client: client,
		Clock:  clock,
	}
}

// String returns string representation of ID
func (id *ID) String() string {
	return fmt.Sprintf("%d:%d", id.Client, id.Clock)
}

// Equals checks if two IDs are equal
func (id *ID) Equals(other *ID) bool {
	if other == nil {
		return false
	}
	return id.Client == other.Client && id.Clock == other.Clock
}

// Compare compares two IDs
func (id *ID) Compare(other *ID) int {
	if id.Client < other.Client {
		return -1
	}
	if id.Client > other.Client {
		return 1
	}
	if id.Clock < other.Clock {
		return -1
	}
	if id.Clock > other.Clock {
		return 1
	}
	return 0
}
