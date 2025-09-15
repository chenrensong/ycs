package core

import (
	"sort"
	"ycs/contracts"
	"ycs/lib0"
)

// DeleteItem represents a deleted item range
type DeleteItem struct {
	Clock  int64
	Length int64
}

// NewDeleteItem creates a new DeleteItem
func NewDeleteItem(clock, length int64) *DeleteItem {
	return &DeleteItem{
		Clock:  clock,
		Length: length,
	}
}

// DeleteSet represents a set of deleted items organized by client
type DeleteSet struct {
	clients map[int64][]*DeleteItem
}

// NewDeleteSet creates a new DeleteSet
func NewDeleteSet() *DeleteSet {
	return &DeleteSet{
		clients: make(map[int64][]*DeleteItem),
	}
}

// NewDeleteSetFromStore creates a DeleteSet from a struct store
func NewDeleteSetFromStore(store contracts.IStructStore) *DeleteSet {
	ds := NewDeleteSet()

	for client, structs := range store.GetClients() {
		var deleteItems []*DeleteItem

		for _, item := range structs {
			if item.GetDeleted() {
				if len(deleteItems) > 0 {
					lastItem := deleteItems[len(deleteItems)-1]
					// Check if we can merge with the last delete item
					if lastItem.Clock+lastItem.Length == item.GetID().Clock {
						lastItem.Length += int64(item.GetLength())
						continue
					}
				}

				deleteItems = append(deleteItems, NewDeleteItem(item.GetID().Clock, int64(item.GetLength())))
			}
		}

		if len(deleteItems) > 0 {
			ds.clients[client] = deleteItems
		}
	}

	return ds
}

// GetClients returns the clients map
func (ds *DeleteSet) GetClients() map[int64][]contracts.DeleteItem {
	result := make(map[int64][]contracts.DeleteItem)
	for client, items := range ds.clients {
		contractItems := make([]contracts.DeleteItem, len(items))
		for i, item := range items {
			contractItems[i] = contracts.DeleteItem{Clock: item.Clock, Length: item.Length}
		}
		result[client] = contractItems
	}
	return result
}

// IsDeleted checks if a struct ID is deleted
func (ds *DeleteSet) IsDeleted(id contracts.StructID) bool {
	deleteItems, exists := ds.clients[id.Client]
	if !exists {
		return false
	}

	for _, item := range deleteItems {
		if item.Clock <= id.Clock && id.Clock < item.Clock+item.Length {
			return true
		}
	}

	return false
}

// Add adds a delete range to the set
func (ds *DeleteSet) Add(client, clock, length int64) {
	if length <= 0 {
		return
	}

	deleteItems := ds.clients[client]
	newItem := NewDeleteItem(clock, length)

	// Find insertion point and merge if possible
	insertPos := 0
	for i, item := range deleteItems {
		if item.Clock > clock {
			insertPos = i
			break
		}
		insertPos = i + 1

		// Check if we can merge with existing item
		if item.Clock+item.Length == clock {
			// Merge with this item
			item.Length += length

			// Check if we can merge with the next item too
			if i+1 < len(deleteItems) {
				nextItem := deleteItems[i+1]
				if item.Clock+item.Length == nextItem.Clock {
					item.Length += nextItem.Length
					// Remove the next item
					deleteItems = append(deleteItems[:i+1], deleteItems[i+2:]...)
				}
			}

			ds.clients[client] = deleteItems
			return
		}

		if clock+length == item.Clock {
			// Merge by extending this item backwards
			item.Clock = clock
			item.Length += length
			ds.clients[client] = deleteItems
			return
		}
	}

	// Insert new item at the correct position
	if insertPos >= len(deleteItems) {
		deleteItems = append(deleteItems, newItem)
	} else {
		deleteItems = append(deleteItems[:insertPos+1], deleteItems[insertPos:]...)
		deleteItems[insertPos] = newItem
	}

	ds.clients[client] = deleteItems
}

// SortAndMergeDeleteSet sorts and merges the delete set
func (ds *DeleteSet) SortAndMergeDeleteSet() {
	for client, deleteItems := range ds.clients {
		if len(deleteItems) <= 1 {
			continue
		}

		// Sort by clock
		sort.Slice(deleteItems, func(i, j int) bool {
			return deleteItems[i].Clock < deleteItems[j].Clock
		})

		// Merge adjacent items
		merged := make([]*DeleteItem, 0, len(deleteItems))
		current := deleteItems[0]

		for i := 1; i < len(deleteItems); i++ {
			next := deleteItems[i]

			if current.Clock+current.Length == next.Clock {
				// Merge with current
				current.Length += next.Length
			} else {
				// Add current and move to next
				merged = append(merged, current)
				current = next
			}
		}

		// Add the last item
		merged = append(merged, current)
		ds.clients[client] = merged
	}
}

// Write writes the delete set to an encoder
func (ds *DeleteSet) Write(encoder contracts.IDSEncoder) error {
	lib0.WriteVarUint(encoder.GetRestWriter(), uint32(len(ds.clients)))

	// Write clients in sorted order for consistency
	var clients []int64
	for client := range ds.clients {
		clients = append(clients, client)
	}
	sort.Slice(clients, func(i, j int) bool {
		return clients[i] < clients[j]
	})

	for _, client := range clients {
		deleteItems := ds.clients[client]
		lib0.WriteVarUint(encoder.GetRestWriter(), uint32(client))
		lib0.WriteVarUint(encoder.GetRestWriter(), uint32(len(deleteItems)))

		encoder.ResetDsCurVal()
		for _, item := range deleteItems {
			encoder.WriteDsClock(item.Clock)
			encoder.WriteDsLength(item.Length)
		}
	}

	return nil
}

// TryGcDeleteSet tries to garbage collect the delete set
func (ds *DeleteSet) TryGcDeleteSet(store contracts.IStructStore, gcFilter func(contracts.IStructItem) bool) {
	for client, deleteItems := range ds.clients {
		structs, exists := store.GetClients()[client]
		if !exists {
			continue
		}

		for _, deleteItem := range deleteItems {
			ds.iterateDeletedStructs(store, structs, deleteItem.Clock, deleteItem.Length, func(item contracts.IStructItem) bool {
				if item.GetDeleted() && (gcFilter == nil || gcFilter(item)) {
					// Replace with GC item
					itemID := item.GetID()
					coreItemID := itemID
					gcItem := NewStructGC(coreItemID, item.GetLength())
					store.ReplaceStruct(item, gcItem)
				}
				return true
			})
		}
	}
}

// TryMergeDeleteSet tries to merge the delete set
func (ds *DeleteSet) TryMergeDeleteSet(store contracts.IStructStore) {
	for client, deleteItems := range ds.clients {
		structs, exists := store.GetClients()[client]
		if !exists {
			continue
		}

		for _, deleteItem := range deleteItems {
			ds.iterateDeletedStructs(store, structs, deleteItem.Clock, deleteItem.Length, func(item contracts.IStructItem) bool {
				if item.GetDeleted() {
					// Try to merge with adjacent items
					ds.tryToMergeWithLeft(structs, item)
				}
				return true
			})
		}
	}
}

// IterateDeletedStructs iterates over deleted structs in a range
func (ds *DeleteSet) IterateDeletedStructs(transaction contracts.ITransaction, fn func(contracts.IStructItem) bool) {
	// This method needs to be implemented according to the interface
	// For now, we'll provide a basic implementation
	for client, deleteItems := range ds.clients {
		structs, exists := transaction.GetStructStore().GetClients()[client]
		if !exists {
			continue
		}

		for _, deleteItem := range deleteItems {
			ds.iterateDeletedStructs(transaction.GetStructStore(), structs, deleteItem.Clock, deleteItem.Length, fn)
		}
	}
}

// iterateDeletedStructs iterates over deleted structs in a range
func (ds *DeleteSet) iterateDeletedStructs(store contracts.IStructStore, structs []contracts.IStructItem, clock, length int64, fn func(contracts.IStructItem) bool) {
	if length <= 0 {
		return
	}

	clockEnd := clock + length
	index := FindIndexSS(structs, clock)

	for index < len(structs) {
		item := structs[index]

		if item.GetID().Clock >= clockEnd {
			break
		}

		if item.GetID().Clock >= clock {
			if !fn(item) {
				break
			}
		}

		index++
	}
}

// TryGc performs garbage collection on the delete set
func (ds *DeleteSet) TryGc(store contracts.IStructStore, gcFilter func(contracts.IStructItem) bool) {
	// TryGcDeleteSet already implements the GC functionality
	ds.TryGcDeleteSet(store, gcFilter)
}

// FindIndexSS finds the index of a delete item with the given clock
func (ds *DeleteSet) FindIndexSS(dis []contracts.DeleteItem, clock int64) *int {
	for i, item := range dis {
		if item.Clock == clock {
			return &i
		}
	}
	return nil
}

// tryToMergeWithLeft tries to merge an item with its left neighbor
func (ds *DeleteSet) tryToMergeWithLeft(structs []contracts.IStructItem, item contracts.IStructItem) {
	index := FindIndexSS(structs, item.GetID().Clock)
	if index <= 0 {
		return
	}

	left := structs[index-1]
	if left.TryToMergeWithRight(item) {
		// Remove the current item since it was merged into left
		copy(structs[index:], structs[index+1:])
		structs = structs[:len(structs)-1]
	}
}
