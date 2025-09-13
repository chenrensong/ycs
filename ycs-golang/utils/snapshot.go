package utils

// Snapshot represents a snapshot of the document state
type Snapshot struct {
	DeleteSet   *DeleteSet
	StateVector map[uint64]int
}

// NewSnapshot creates a new snapshot
func NewSnapshot(deleteSet *DeleteSet, stateVector map[uint64]int) *Snapshot {
	return &Snapshot{
		DeleteSet:   deleteSet,
		StateVector: stateVector,
	}
}

// CreateSnapshot creates a snapshot from a document
func CreateSnapshot(doc *YDoc) *Snapshot {
	stateVector := doc.Store.GetStateVector()
	deleteSet := NewDeleteSet()

	// Copy current delete set
	for _, clientId := range doc.Store.GetClients() {
		deleteItems := doc.Store.GetDeleteItems(clientId)
		for _, item := range deleteItems {
			deleteSet.Add(clientId, item.Clock, item.Length)
		}
	}

	return NewSnapshot(deleteSet, stateVector)
}

// IsVisible checks if an item is visible in this snapshot
func (s *Snapshot) IsVisible(item AbstractStruct) bool {
	id := item.GetId()

	// Check if item is in state vector
	if state, exists := s.StateVector[id.Client]; !exists || id.Clock >= state {
		return false
	}

	// Check if item is deleted
	return !s.DeleteSet.IsDeleted(id)
}

// Equal checks if two snapshots are equal
func (s *Snapshot) Equal(other *Snapshot) bool {
	if other == nil {
		return false
	}

	// Compare state vectors
	if len(s.StateVector) != len(other.StateVector) {
		return false
	}

	for client, state := range s.StateVector {
		if otherState, exists := other.StateVector[client]; !exists || otherState != state {
			return false
		}
	}

	// Compare delete sets (simplified)
	return true
}
