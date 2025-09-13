// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"github.com/chenrensong/ygo/contracts"
)

// Snapshot represents a snapshot of the document state
type Snapshot struct {
	deleteSet   contracts.IDeleteSet
	stateVector map[int64]int64
}

// NewSnapshot creates a new Snapshot
func NewSnapshot(deleteSet contracts.IDeleteSet, stateVector map[int64]int64) *Snapshot {
	return &Snapshot{
		deleteSet:   deleteSet,
		stateVector: stateVector,
	}
}

// GetDeleteSet returns the delete set
func (s *Snapshot) GetDeleteSet() contracts.IDeleteSet {
	return s.deleteSet
}

// GetStateVector returns the state vector
func (s *Snapshot) GetStateVector() map[int64]int64 {
	return s.stateVector
}

// CreateSnapshot creates a snapshot from a document
func CreateSnapshot(doc contracts.IYDoc) *Snapshot {
	return NewSnapshot(
		NewDeleteSetFromStructStore(doc.GetStore()),
		doc.GetStore().GetStateVector(),
	)
}

// CreateSnapshotFromDeleteSet creates a snapshot from a delete set and state vector
func CreateSnapshotFromDeleteSet(deleteSet contracts.IDeleteSet, stateVector map[int64]int64) *Snapshot {
	return NewSnapshot(deleteSet, stateVector)
}

// EmptySnapshot creates an empty snapshot
func EmptySnapshot() *Snapshot {
	return NewSnapshot(
		NewDeleteSet(),
		make(map[int64]int64),
	)
}

// SnapshotContainsUpdate checks if a snapshot contains an update
func SnapshotContainsUpdate(snapshot *Snapshot, update []byte) bool {
	// This is a simplified implementation
	// In a full implementation, this would decode the update and check against the snapshot
	return false
}

// IsVisible checks if an item is visible in the snapshot
func (s *Snapshot) IsVisible(item contracts.IStructItem) bool {
	if item.GetDeleted() {
		return false
	}

	// Check if the item's clock is within the state vector
	clientState, exists := s.stateVector[item.GetID().Client]
	if !exists {
		return false
	}

	return item.GetID().Clock < clientState
}

// SplitSnapshotAffectedStructs splits structs affected by the snapshot
func (s *Snapshot) SplitSnapshotAffectedStructs(transaction contracts.ITransaction) {
	store := transaction.GetDoc().GetStore()

	for client, clock := range s.stateVector {
		structs := store.GetClients()[client]
		if len(structs) == 0 {
			continue
		}

		// Find the struct at the snapshot boundary
		index := FindIndexSS(structs, clock)
		if index < len(structs) {
			item := structs[index]
			if item.GetID().Clock < clock && item.GetID().Clock+int64(item.GetLength()) > clock {
				// Split the item at the snapshot boundary
				diff := int(clock - item.GetID().Clock)
				if splitItem, err := item.SplitItem(transaction, diff); err == nil {
					structs = append(structs[:index+1], append([]contracts.IStructItem{splitItem}, structs[index+1:]...)...)
					store.GetClients()[client] = structs
				}
			}
		}
	}
}

// RestoreDocument restores a document to the state of this snapshot
func (s *Snapshot) RestoreDocument(doc contracts.IYDoc) contracts.IYDoc {
	// Create a new document
	newDoc := doc.Clone()

	// Apply the snapshot state
	newDoc.Transact(func(transaction contracts.ITransaction) {
		s.SplitSnapshotAffectedStructs(transaction)

		// Mark items as deleted that are not in the snapshot
		for client, structs := range newDoc.GetStore().GetClients() {
			snapshotClock, exists := s.stateVector[client]
			if !exists {
				snapshotClock = 0
			}

			for _, item := range structs {
				if item.GetID().Clock >= snapshotClock && !item.GetDeleted() {
					item.Delete(transaction)
				}
			}
		}

		// Apply delete set
		s.deleteSet.IterateDeletedStructs(transaction, func(item contracts.IStructItem) bool {
			if !item.GetDeleted() {
				item.Delete(transaction)
			}
			return true
		})
	}, nil)

	return newDoc
}

// EncodeSnapshotV2 encodes a snapshot to bytes
func (s *Snapshot) EncodeSnapshotV2() []byte {
	// This is a simplified implementation
	// In a full implementation, this would properly encode the snapshot
	return nil
}

// DecodeSnapshotV2 decodes a snapshot from bytes
func DecodeSnapshotV2(data []byte) *Snapshot {
	// This is a simplified implementation
	// In a full implementation, this would properly decode the snapshot
	return EmptySnapshot()
}

// EqualSnapshots checks if two snapshots are equal
func EqualSnapshots(snap1, snap2 *Snapshot) bool {
	// Compare state vectors
	if len(snap1.stateVector) != len(snap2.stateVector) {
		return false
	}

	for client, clock1 := range snap1.stateVector {
		if clock2, exists := snap2.stateVector[client]; !exists || clock1 != clock2 {
			return false
		}
	}

	// Compare delete sets (simplified)
	clients1 := snap1.deleteSet.GetClients()
	clients2 := snap2.deleteSet.GetClients()

	if len(clients1) != len(clients2) {
		return false
	}

	for client, deletes1 := range clients1 {
		deletes2, exists := clients2[client]
		if !exists || len(deletes1) != len(deletes2) {
			return false
		}

		for i, delete1 := range deletes1 {
			delete2 := deletes2[i]
			if delete1.Clock != delete2.Clock || delete1.Length != delete2.Length {
				return false
			}
		}
	}

	return true
}
