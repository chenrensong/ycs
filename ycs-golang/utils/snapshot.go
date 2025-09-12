// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"bytes"
	"ycs-golang/core"
)

// Snapshot represents a snapshot of a document
type Snapshot struct {
	DeleteSet   *DeleteSet
	StateVector map[int64]int64
}

// NewSnapshot creates a new Snapshot
func NewSnapshot(ds *DeleteSet, stateMap map[int64]int64) *Snapshot {
	return &Snapshot{
		DeleteSet:   ds,
		StateVector: stateMap,
	}
}

// RestoreDocument restores a document from a snapshot
func (s *Snapshot) RestoreDocument(originDoc *YDoc, opts *YDocOptions) *YDoc {
	if originDoc.Gc {
		// We should try to restore a GC-ed document, because some of the restored items might have their content deleted.
		panic("originDoc must not be garbage collected")
	}

	// Note: In Go, we don't have a direct equivalent of using statement.
	// We'll handle resource cleanup manually.
	encoder := NewUpdateEncoderV2()
	defer encoder.Dispose()

	originDoc.Transact(func(tr *Transaction) {
		size := 0
		for _, clock := range s.StateVector {
			if clock > 0 {
				size++
			}
		}
		core.WriteVarUint(encoder.RestWriter, uint64(size))

		// Splitting the structs before writing them to the encoder.
		for client, clock := range s.StateVector {
			if clock == 0 {
				continue
			}

			if clock < originDoc.Store.GetState(client) {
				tr.Doc.Store.GetItemCleanStart(tr, ID{Client: client, Clock: clock})
			}

			structs := originDoc.Store.Clients[client]
			lastStructIndex := StructStoreFindIndexSS(structs, clock-1)

			// Write # encoded structs.
			core.WriteVarUint(encoder.RestWriter, uint64(lastStructIndex+1))
			encoder.WriteClient(client)

			// First clock written is 0.
			core.WriteVarUint(encoder.RestWriter, 0)

			for i := 0; i <= lastStructIndex; i++ {
				structs[i].Write(encoder, 0)
			}
		}

		s.DeleteSet.Write(encoder)
	})

	newDoc := NewYDoc(opts)
	if opts == nil {
		newDoc.ApplyUpdateV2(encoder.ToArray(), "snapshot")
	} else {
		newDoc.ApplyUpdateV2(encoder.ToArray(), "snapshot")
	}
	
	return newDoc
}

// Equals checks if two snapshots are equal
func (s *Snapshot) Equals(other *Snapshot) bool {
	if other == nil {
		return false
	}

	ds1 := s.DeleteSet.Clients
	ds2 := other.DeleteSet.Clients
	sv1 := s.StateVector
	sv2 := other.StateVector

	if len(sv1) != len(sv2) || len(ds1) != len(ds2) {
		return false
	}

	for client, clock := range sv1 {
		if otherClock, exists := sv2[client]; !exists || otherClock != clock {
			return false
		}
	}

	for client, dsItems1 := range ds1 {
		dsItems2, exists := ds2[client]
		if !exists {
			return false
		}

		if len(dsItems1) != len(dsItems2) {
			return false
		}

		for i := 0; i < len(dsItems1); i++ {
			dsItem1 := dsItems1[i]
			dsItem2 := dsItems2[i]
			if dsItem1.Clock != dsItem2.Clock || dsItem1.Length != dsItem2.Length {
				return false
			}
		}
	}

	return true
}

// EncodeSnapshotV2 encodes a snapshot to V2 format
func (s *Snapshot) EncodeSnapshotV2() []byte {
	encoder := NewDSEncoderV2()
	defer encoder.Dispose()

	s.DeleteSet.Write(encoder)
	// Note: EncodingUtils.WriteStateVector needs to be implemented
	// EncodingUtils.WriteStateVector(encoder, s.StateVector)
	
	return encoder.ToArray()
}

// DecodeSnapshot decodes a snapshot from a stream
func DecodeSnapshot(input *bytes.Reader) *Snapshot {
	decoder := NewDSDecoderV2(input)
	defer decoder.Dispose()

	ds := ReadDeleteSet(decoder)
	// Note: EncodingUtils.ReadStateVector needs to be implemented
	// sv := EncodingUtils.ReadStateVector(decoder)
	
	// Placeholder for StateVector - will need to be implemented
	sv := make(map[int64]int64)
	
	return NewSnapshot(ds, sv)
}