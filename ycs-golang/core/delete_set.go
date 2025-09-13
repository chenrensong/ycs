// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"sort"

	"github.com/chenrensong/ygo/contracts"
)

// DeleteSet is a temporary object that is created when needed.
// - When created in a transaction, it must only be accessed after sorting and merging.
//   - This DeleteSet is sent to other clients.
//   - We do not create a DeleteSet when we send a sync message. The DeleteSet message is created
//     directly from StructStore.
//   - We read a DeleteSet as a apart of sync/update message. In this case the DeleteSet is already
//     sorted and merged.
type DeleteSet struct {
	clients map[int64][]contracts.DeleteItem
}

// NewDeleteSet creates a new DeleteSet
func NewDeleteSet() *DeleteSet {
	return &DeleteSet{
		clients: make(map[int64][]contracts.DeleteItem),
	}
}

// NewDeleteSetFromDeleteSets creates a new DeleteSet by merging multiple DeleteSets
func NewDeleteSetFromDeleteSets(dss []contracts.IDeleteSet) *DeleteSet {
	ds := NewDeleteSet()
	ds.MergeDeleteSets(dss)
	return ds
}

// NewDeleteSetFromStructStore creates a new DeleteSet from a StructStore
func NewDeleteSetFromStructStore(ss contracts.IStructStore) *DeleteSet {
	ds := NewDeleteSet()
	ds.CreateDeleteSetFromStructStore(ss)
	return ds
}

// GetClients returns the clients map
func (ds *DeleteSet) GetClients() map[int64][]contracts.DeleteItem {
	return ds.clients
}

// Add adds a delete item to the set
func (ds *DeleteSet) Add(client int64, clock int64, length int64) {
	deletes, exists := ds.clients[client]
	if !exists {
		deletes = make([]contracts.DeleteItem, 0, 2)
	}

	deletes = append(deletes, contracts.DeleteItem{
		Clock:  clock,
		Length: length,
	})
	ds.clients[client] = deletes
}

// IterateDeletedStructs iterates over all structs that the DeleteSet gc'd
func (ds *DeleteSet) IterateDeletedStructs(transaction contracts.ITransaction, predicate func(contracts.IStructItem) bool) {
	for client, deleteItems := range ds.clients {
		structs := transaction.GetDoc().GetStore().GetClients()[client]
		for _, del := range deleteItems {
			transaction.GetDoc().GetStore().IterateStructs(transaction, structs, del.Clock, del.Length, predicate)
		}
	}
}

// FindIndexSS performs binary search to find index of delete item
func (ds *DeleteSet) FindIndexSS(dis []contracts.DeleteItem, clock int64) *int {
	left := 0
	right := len(dis) - 1

	for left <= right {
		midIndex := (left + right) / 2
		mid := dis[midIndex]
		midClock := mid.Clock

		if midClock <= clock {
			if clock < midClock+mid.Length {
				return &midIndex
			}
			left = midIndex + 1
		} else {
			right = midIndex - 1
		}
	}

	return nil
}

// IsDeleted checks if a struct ID is deleted
func (ds *DeleteSet) IsDeleted(id contracts.StructID) bool {
	if dis, exists := ds.clients[id.Client]; exists {
		return ds.FindIndexSS(dis, id.Clock) != nil
	}
	return false
}

// SortAndMergeDeleteSet sorts and merges the delete set
func (ds *DeleteSet) SortAndMergeDeleteSet() {
	for _, dels := range ds.clients {
		// Sort by clock
		sort.Slice(dels, func(i, j int) bool {
			return dels[i].Clock < dels[j].Clock
		})

		// Merge items without filtering or splicing the array
		// i is the current pointer
		// j refers to the current insert position for the pointed item
		// Try to merge dels[i] into dels[j-1] or set dels[j]=dels[i]
		if len(dels) <= 1 {
			continue
		}

		j := 1
		for i := 1; i < len(dels); i++ {
			left := dels[j-1]
			right := dels[i]

			if left.Clock+left.Length == right.Clock {
				// Merge items
				dels[j-1] = contracts.DeleteItem{
					Clock:  left.Clock,
					Length: left.Length + right.Length,
				}
			} else {
				if j < i {
					dels[j] = right
				}
				j++
			}
		}

		// Trim the slice
		if j < len(dels) {
			// Update the slice in the map
			trimmed := make([]contracts.DeleteItem, j)
			copy(trimmed, dels[:j])
			for client, clientDels := range ds.clients {
				if &clientDels[0] == &dels[0] { // Find the matching slice
					ds.clients[client] = trimmed
					break
				}
			}
		}
	}
}

// TryGc tries to garbage collect the delete set
func (ds *DeleteSet) TryGc(store contracts.IStructStore, gcFilter func(contracts.IStructItem) bool) {
	ds.TryGcDeleteSet(store, gcFilter)
	ds.TryMergeDeleteSet(store)
}

// TryGcDeleteSet tries to garbage collect deleted structs
func (ds *DeleteSet) TryGcDeleteSet(store contracts.IStructStore, gcFilter func(contracts.IStructItem) bool) {
	for client, deleteItems := range ds.clients {
		structs := store.GetClients()[client]

		for di := len(deleteItems) - 1; di >= 0; di-- {
			deleteItem := deleteItems[di]
			endDeleteItemClock := deleteItem.Clock + deleteItem.Length

			for si := FindIndexSS(structs, deleteItem.Clock); si < len(structs); si++ {
				str := structs[si]
				if str.GetID().Clock >= endDeleteItemClock {
					break
				}

				if str.GetDeleted() && !str.GetKeep() && gcFilter(str) {
					str.Gc(store, false)
				}
			}
		}
	}
}

// TryMergeDeleteSet tries to merge deleted/gc'd items
func (ds *DeleteSet) TryMergeDeleteSet(store contracts.IStructStore) {
	// Try to merge deleted / gc'd items
	// Merge from right to left for better efficiency and so we don't miss any merge targets
	for client, deleteItems := range ds.clients {
		structs := store.GetClients()[client]

		for di := len(deleteItems) - 1; di >= 0; di-- {
			deleteItem := deleteItems[di]

			// Start with merging the item next to the last deleted item
			mostRightIndexToCheck := min(len(structs)-1, 1+FindIndexSS(structs, deleteItem.Clock+deleteItem.Length-1))
			for si := mostRightIndexToCheck; si > 0 && structs[si].GetID().Clock >= deleteItem.Clock; si-- {
				TryToMergeWithLeft(structs, si)
			}
		}
	}
}

// TryToMergeWithLeft tries to merge a struct with the one to its left
func TryToMergeWithLeft(structs []contracts.IStructItem, pos int) {
	left := structs[pos-1]
	right := structs[pos]

	if left.GetDeleted() == right.GetDeleted() {
		if left.MergeWith(right) {
			// Remove the right item from the slice
			copy(structs[pos:], structs[pos+1:])
			structs = structs[:len(structs)-1]

			// Update parent map if necessary
			if right.GetParentSub() != "" {
				if parent, ok := right.GetParent().(contracts.IAbstractType); ok {
					if mapItem, exists := parent.GetMap()[right.GetParentSub()]; exists && mapItem == right {
						parent.GetMap()[right.GetParentSub()] = left
					}
				}
			}
		}
	}
}

// MergeDeleteSets merges multiple delete sets into this one
func (ds *DeleteSet) MergeDeleteSets(dss []contracts.IDeleteSet) {
	for dssI := 0; dssI < len(dss); dssI++ {
		for client, delsLeft := range dss[dssI].GetClients() {
			if _, exists := ds.clients[client]; !exists {
				// Write all missing keys from current ds and all following
				// If merged already contains 'client' current ds has already been added
				dels := make([]contracts.DeleteItem, len(delsLeft))
				copy(dels, delsLeft)

				for i := dssI + 1; i < len(dss); i++ {
					if appends, exists := dss[i].GetClients()[client]; exists {
						dels = append(dels, appends...)
					}
				}

				ds.clients[client] = dels
			}
		}
	}

	ds.SortAndMergeDeleteSet()
}

// CreateDeleteSetFromStructStore creates delete set from struct store
func (ds *DeleteSet) CreateDeleteSetFromStructStore(ss contracts.IStructStore) {
	for client, structs := range ss.GetClients() {
		var dsItems []contracts.DeleteItem

		for i := 0; i < len(structs); i++ {
			str := structs[i]
			if str.GetDeleted() {
				clock := str.GetID().Clock
				length := int64(str.GetLength())

				// Merge consecutive deleted items
				for i+1 < len(structs) {
					next := structs[i+1]
					if next.GetID().Clock == clock+length && next.GetDeleted() {
						length += int64(next.GetLength())
						i++
					} else {
						break
					}
				}

				dsItems = append(dsItems, contracts.DeleteItem{
					Clock:  clock,
					Length: length,
				})
			}
		}

		if len(dsItems) > 0 {
			ds.clients[client] = dsItems
		}
	}
}

// Write writes the delete set to an encoder
func (ds *DeleteSet) Write(encoder contracts.IDSEncoder) {
	encoder.GetRestWriter().WriteVarUint(uint64(len(ds.clients)))

	for client, dsItems := range ds.clients {
		length := len(dsItems)

		encoder.ResetDsCurVal()
		encoder.GetRestWriter().WriteVarUint(uint64(client))
		encoder.GetRestWriter().WriteVarUint(uint64(length))

		for i := 0; i < length; i++ {
			item := dsItems[i]
			encoder.WriteDsClock(item.Clock)
			encoder.WriteDsLength(item.Length)
		}
	}
}

// ReadDeleteSet reads a delete set from a decoder
func ReadDeleteSet(decoder contracts.IDSDecoder) *DeleteSet {
	ds := NewDeleteSet()

	numClients := decoder.GetReader().ReadVarUint()

	for i := uint64(0); i < numClients; i++ {
		decoder.ResetDsCurVal()

		client := int64(decoder.GetReader().ReadVarUint())
		numberOfDeletes := decoder.GetReader().ReadVarUint()

		if numberOfDeletes > 0 {
			dsField, exists := ds.clients[client]
			if !exists {
				dsField = make([]contracts.DeleteItem, 0, numberOfDeletes)
			}

			for j := uint64(0); j < numberOfDeletes; j++ {
				deleteItem := contracts.DeleteItem{
					Clock:  decoder.ReadDsClock(),
					Length: decoder.ReadDsLength(),
				}
				dsField = append(dsField, deleteItem)
			}

			ds.clients[client] = dsField
		}
	}

	return ds
}

// Helper function to find minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
