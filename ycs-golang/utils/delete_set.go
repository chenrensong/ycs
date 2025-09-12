// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"container/list"
	"sort"
	"ycs-golang/core"
	"ycs-golang/structs"
	"ycs-golang/types"
)

// DeleteItem represents a deleted item in the DeleteSet
type DeleteItem struct {
	Clock  int64
	Length int64
}

// DeleteSet is a temporary object that is created when needed
type DeleteSet struct {
	Clients map[int64][]DeleteItem
}

// NewDeleteSet creates a new DeleteSet
func NewDeleteSet() *DeleteSet {
	return &DeleteSet{
		Clients: make(map[int64][]DeleteItem),
	}
}

// NewDeleteSetFromList creates a new DeleteSet from a list of DeleteSets
func NewDeleteSetFromList(dss []*DeleteSet) *DeleteSet {
	ds := NewDeleteSet()
	ds.MergeDeleteSets(dss)
	return ds
}

// NewDeleteSetFromStructStore creates a new DeleteSet from a StructStore
func NewDeleteSetFromStructStore(ss *StructStore) *DeleteSet {
	ds := NewDeleteSet()
	ds.CreateDeleteSetFromStructStore(ss)
	return ds
}

// Add adds a new delete item to the DeleteSet
func (ds *DeleteSet) Add(client, clock, length int64) {
	if _, exists := ds.Clients[client]; !exists {
		ds.Clients[client] = make([]DeleteItem, 0, 2)
	}
	
	ds.Clients[client] = append(ds.Clients[client], DeleteItem{Clock: clock, Length: length})
}

// IterateDeletedStructs iterates over all structs that the DeleteSet gc'd
func (ds *DeleteSet) IterateDeletedStructs(transaction *Transaction, fn func(*structs.AbstractStruct) bool) {
	for client, deleteItems := range ds.Clients {
		structs := transaction.Doc.Store.Clients[client]
		for _, del := range deleteItems {
			transaction.Doc.Store.IterateStructs(transaction, structs, del.Clock, del.Length, fn)
		}
	}
}

// FindIndexSS finds the index of a DeleteItem in a list
func (ds *DeleteSet) FindIndexSS(dis []DeleteItem, clock int64) *int {
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

// IsDeleted checks if an ID is deleted
func (ds *DeleteSet) IsDeleted(id ID) bool {
	if dis, exists := ds.Clients[id.Client]; exists {
		return ds.FindIndexSS(dis, id.Clock) != nil
	}
	return false
}

// SortAndMergeDeleteSet sorts and merges the DeleteSet
func (ds *DeleteSet) SortAndMergeDeleteSet() {
	for client, dels := range ds.Clients {
		// Sort by clock
		sort.Slice(dels, func(i, j int) bool {
			return dels[i].Clock < dels[j].Clock
		})

		// Merge items without filtering or splicing the array
		// i is the current pointer
		// j refers to the current insert position for the pointed item
		// Try to merge dels[i] into dels[j-1] or set dels[j]=dels[i]
		j := 1
		for i := 1; i < len(dels); i++ {
			left := dels[j-1]
			right := dels[i]

			if left.Clock+left.Length == right.Clock {
				dels[j-1] = DeleteItem{Clock: left.Clock, Length: left.Length + right.Length}
			} else {
				if j < i {
					dels[j] = right
				}
				j++
			}
		}

		// Trim the collection
		if j < len(dels) {
			ds.Clients[client] = dels[:j]
		} else {
			ds.Clients[client] = dels
		}
	}
}

// TryGc tries to garbage collect
func (ds *DeleteSet) TryGc(store *StructStore, gcFilter func(*structs.Item) bool) {
	ds.TryGcDeleteSet(store, gcFilter)
	ds.TryMergeDeleteSet(store)
}

// TryGcDeleteSet tries to garbage collect the delete set
func (ds *DeleteSet) TryGcDeleteSet(store *StructStore, gcFilter func(*structs.Item) bool) {
	for client, deleteItems := range ds.Clients {
		structs := store.Clients[client]

		for di := len(deleteItems) - 1; di >= 0; di-- {
			deleteItem := deleteItems[di]
			endDeleteItemClock := deleteItem.Clock + deleteItem.Length

			si := StructStoreFindIndexSS(structs, deleteItem.Clock)
			for si < len(structs) {
				str := structs[si]
				if str.Id.Clock >= endDeleteItemClock {
					break
				}

				if strItem, ok := str.(*structs.Item); ok && strItem.Deleted && !strItem.Keep && gcFilter(strItem) {
					strItem.Gc(store, false)
				}
				si++
			}
		}
	}
}

// TryMergeDeleteSet tries to merge the delete set
func (ds *DeleteSet) TryMergeDeleteSet(store *StructStore) {
	// Try to merge deleted / gc'd items
	// Merge from right to left for better efficiency and so we don't miss any merge targets
	for client, deleteItems := range ds.Clients {
		structs := store.Clients[client]

		for di := len(deleteItems) - 1; di >= 0; di-- {
			deleteItem := deleteItems[di]

			// Start with merging the item next to the last deleted item
			mostRightIndexToCheck := min(len(structs)-1, 1+StructStoreFindIndexSS(structs, deleteItem.Clock+deleteItem.Length-1))
			for si := mostRightIndexToCheck; si > 0 && structs[si].Id.Clock >= deleteItem.Clock; si-- {
				TryToMergeWithLeft(structs, si)
			}
		}
	}
}

// TryToMergeWithLeft tries to merge with the left item
func TryToMergeWithLeft(structs []*structs.AbstractStruct, pos int) {
	left := structs[pos-1]
	right := structs[pos]

	if left.Deleted == right.Deleted && left.GetType() == right.GetType() {
		if left.MergeWith(right) {
			// Remove the right item
			copy(structs[pos:], structs[pos+1:])
			structs[len(structs)-1] = nil
			structs = structs[:len(structs)-1]

			if rightItem, ok := right.(*structs.Item); ok && rightItem.ParentSub != "" {
				if parent, ok := rightItem.Parent.(*types.AbstractType); ok {
					if value, exists := parent.Map[rightItem.ParentSub]; exists && value == rightItem {
						parent.Map[rightItem.ParentSub] = left.(*structs.Item)
					}
				}
			}
		}
	}
}

// MergeDeleteSets merges multiple DeleteSets
func (ds *DeleteSet) MergeDeleteSets(dss []*DeleteSet) {
	for dssI := 0; dssI < len(dss); dssI++ {
		for client, delsLeft := range dss[dssI].Clients {
			if _, exists := ds.Clients[client]; !exists {
				// Write all missing keys from current ds and all following
				// If merged already contains 'client' current ds has already been added
				dels := make([]DeleteItem, len(delsLeft))
				copy(dels, delsLeft)

				for i := dssI + 1; i < len(dss); i++ {
					if appends, exists := dss[i].Clients[client]; exists {
						dels = append(dels, appends...)
					}
				}

				ds.Clients[client] = dels
			}
		}
	}

	ds.SortAndMergeDeleteSet()
}

// CreateDeleteSetFromStructStore creates a DeleteSet from a StructStore
func (ds *DeleteSet) CreateDeleteSetFromStructStore(ss *StructStore) {
	for client, structs := range ss.Clients {
		dsItems := make([]DeleteItem, 0)

		for i := 0; i < len(structs); i++ {
			str := structs[i]
			if str.Deleted {
				clock := str.Id.Clock
				length := str.Length

				for i+1 < len(structs) {
					next := structs[i+1]
					if next.Id.Clock == clock+length && next.Deleted {
						length += next.Length
						i++
					} else {
						break
					}
				}

				dsItems = append(dsItems, DeleteItem{Clock: clock, Length: length})
			}
		}

		if len(dsItems) > 0 {
			ds.Clients[client] = dsItems
		}
	}
}

// Write writes the DeleteSet to an encoder
func (ds *DeleteSet) Write(encoder IDSDecoder) {
	// Write the number of clients
	core.WriteVarUint(encoder.RestWriter, uint64(len(ds.Clients)))

	for client, dsItems := range ds.Clients {
		length := len(dsItems)

		encoder.ResetDsCurVal()
		core.WriteVarUint(encoder.RestWriter, uint64(client))
		core.WriteVarUint(encoder.RestWriter, uint64(length))

		for i := 0; i < length; i++ {
			item := dsItems[i]
			encoder.WriteDsClock(item.Clock)
			encoder.WriteDsLength(item.Length)
		}
	}
}

// Read reads a DeleteSet from a decoder
func ReadDeleteSet(decoder IDSDecoder) *DeleteSet {
	ds := NewDeleteSet()

	numClients := core.ReadVarUint(decoder.Reader)
	
	for i := uint64(0); i < numClients; i++ {
		decoder.ResetDsCurVal()

		client := core.ReadVarUint(decoder.Reader)
		numberOfDeletes := core.ReadVarUint(decoder.Reader)

		if numberOfDeletes > 0 {
			if _, exists := ds.Clients[int64(client)]; !exists {
				ds.Clients[int64(client)] = make([]DeleteItem, 0, int(numberOfDeletes))
			}

			for j := uint64(0); j < numberOfDeletes; j++ {
				deleteItem := DeleteItem{
					Clock:  decoder.ReadDsClock(),
					Length: decoder.ReadDsLength(),
				}
				ds.Clients[int64(client)] = append(ds.Clients[int64(client)], deleteItem)
			}
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