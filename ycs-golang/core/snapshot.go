package core

import (
	"ycs/contracts"
	"ycs/lib0"
)

// Snapshot represents a snapshot of a document state
type Snapshot struct {
	DeleteSet   contracts.IDeleteSet
	StateVector map[int64]int64
}

// NewSnapshot creates a new Snapshot
func NewSnapshot(ds contracts.IDeleteSet, stateMap map[int64]int64) *Snapshot {
	return &Snapshot{
		DeleteSet:   ds,
		StateVector: stateMap,
	}
}

// RestoreDocument restores a document from this snapshot
func (s *Snapshot) RestoreDocument(originDoc contracts.IYDoc, opts *contracts.YDocOptions) contracts.IYDoc {
	if originDoc.GetGc() {
		// We should try to restore a GC-ed document, because some of the restored items might have their content deleted.
		panic("originDoc must not be garbage collected")
	}

	encoder := NewUpdateEncoderV2()
	defer encoder.Close()

	originDoc.Transact(func(tr contracts.ITransaction) {
		// Count non-zero states
		size := 0
		for _, clock := range s.StateVector {
			if clock > 0 {
				size++
			}
		}
		lib0.WriteVarUint(encoder.restWriter, uint32(size))

		// Splitting the structs before writing them to the encoder.
		for client, clock := range s.StateVector {
			if clock == 0 {
				continue
			}

			if clock < originDoc.GetStore().GetState(client) {
				tr.GetDoc().GetStore().GetItemCleanStart(tr, contracts.StructID{Client: client, Clock: clock})
			}

			structs := originDoc.GetStore().GetClients()[client]
			lastStructIndex := FindIndexSS(structs, clock-1)

			// Write # encoded structs.
			lib0.WriteVarUint(encoder.restWriter, uint32(lastStructIndex+1))
			encoder.WriteClient(client)

			// First clock written is 0.
			lib0.WriteVarUint(encoder.restWriter, uint32(0))

			for i := 0; i <= lastStructIndex; i++ {
				structs[i].Write(encoder, 0)
			}
		}

		s.DeleteSet.Write(encoder)
	}, "snapshot")

	newDoc := NewYDoc(*opts)
	newDoc.ApplyUpdateV2Bytes(encoder.ToArray(), "snapshot", true)

	return newDoc
}

// Equals compares two snapshots for equality
func (s *Snapshot) Equals(other *Snapshot) bool {
	if other == nil {
		return false
	}

	ds1 := s.DeleteSet.GetClients()
	ds2 := other.DeleteSet.GetClients()
	sv1 := s.StateVector
	sv2 := other.StateVector

	if len(sv1) != len(sv2) || len(ds1) != len(ds2) {
		return false
	}

	for client, clock := range sv1 {
		if otherClock, exists := sv2[client]; !exists || clock != otherClock {
			return false
		}
	}

	for client, deleteItems := range ds1 {
		otherDeleteItems, exists := ds2[client]
		if !exists || len(deleteItems) != len(otherDeleteItems) {
			return false
		}

		for i, item := range deleteItems {
			if item.Clock != otherDeleteItems[i].Clock || item.Length != otherDeleteItems[i].Length {
				return false
			}
		}
	}

	return true
}

// GetDeleteSet returns the delete set
func (s *Snapshot) GetDeleteSet() contracts.IDeleteSet {
	return s.DeleteSet
}

// GetStateVector returns the state vector
func (s *Snapshot) GetStateVector() map[int64]int64 {
	return s.StateVector
}

// EncodeSnapshotV2 encodes the snapshot as bytes
func (s *Snapshot) EncodeSnapshotV2() []byte {
	// TODO: Implement proper snapshot encoding
	return []byte{}
}
