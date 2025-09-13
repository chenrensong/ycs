package utils

import (
	"sort"
	"sync"
)

// DeleteItem represents a deleted item range
type DeleteItem struct {
	Clock  int
	Length int
}

// DeleteSet manages deleted items by client
type DeleteSet struct {
	clients map[uint64][]DeleteItem
	mutex   sync.RWMutex
}

// NewDeleteSet creates a new DeleteSet
func NewDeleteSet() *DeleteSet {
	return &DeleteSet{
		clients: make(map[uint64][]DeleteItem),
	}
}

// Add adds a delete item to the set
func (ds *DeleteSet) Add(clientId uint64, clock int, length int) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	if ds.clients[clientId] == nil {
		ds.clients[clientId] = make([]DeleteItem, 0)
	}

	deleteItem := DeleteItem{
		Clock:  clock,
		Length: length,
	}

	ds.clients[clientId] = append(ds.clients[clientId], deleteItem)

	// Keep the delete items sorted by clock
	sort.Slice(ds.clients[clientId], func(i, j int) bool {
		return ds.clients[clientId][i].Clock < ds.clients[clientId][j].Clock
	})

	// Merge adjacent delete items
	ds.mergeDeleteItems(clientId)
}

// IsDeleted checks if an item is deleted
func (ds *DeleteSet) IsDeleted(id *ID) bool {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	deleteItems, exists := ds.clients[id.Client]
	if !exists {
		return false
	}

	for _, item := range deleteItems {
		if id.Clock >= item.Clock && id.Clock < item.Clock+item.Length {
			return true
		}
	}

	return false
}

// GetDeletedLength gets the length of deleted items for a client
func (ds *DeleteSet) GetDeletedLength(clientId uint64) int {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	deleteItems, exists := ds.clients[clientId]
	if !exists {
		return 0
	}

	totalLength := 0
	for _, item := range deleteItems {
		totalLength += item.Length
	}

	return totalLength
}

// Merge merges another DeleteSet into this one
func (ds *DeleteSet) Merge(other *DeleteSet) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	other.mutex.RLock()
	defer other.mutex.RUnlock()

	for clientId, deleteItems := range other.clients {
		for _, item := range deleteItems {
			ds.addUnsafe(clientId, item.Clock, item.Length)
		}
	}
}

// addUnsafe adds a delete item without locking (internal use)
func (ds *DeleteSet) addUnsafe(clientId uint64, clock int, length int) {
	if ds.clients[clientId] == nil {
		ds.clients[clientId] = make([]DeleteItem, 0)
	}

	deleteItem := DeleteItem{
		Clock:  clock,
		Length: length,
	}

	ds.clients[clientId] = append(ds.clients[clientId], deleteItem)

	// Keep the delete items sorted by clock
	sort.Slice(ds.clients[clientId], func(i, j int) bool {
		return ds.clients[clientId][i].Clock < ds.clients[clientId][j].Clock
	})

	// Merge adjacent delete items
	ds.mergeDeleteItems(clientId)
}

// mergeDeleteItems merges adjacent delete items for a client
func (ds *DeleteSet) mergeDeleteItems(clientId uint64) {
	deleteItems := ds.clients[clientId]
	if len(deleteItems) <= 1 {
		return
	}

	merged := make([]DeleteItem, 0, len(deleteItems))
	current := deleteItems[0]

	for i := 1; i < len(deleteItems); i++ {
		next := deleteItems[i]

		if current.Clock+current.Length == next.Clock {
			// Merge adjacent items
			current.Length += next.Length
		} else {
			merged = append(merged, current)
			current = next
		}
	}

	merged = append(merged, current)
	ds.clients[clientId] = merged
}

// GetClients gets all client IDs that have delete items
func (ds *DeleteSet) GetClients() []uint64 {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	clients := make([]uint64, 0, len(ds.clients))
	for clientId := range ds.clients {
		clients = append(clients, clientId)
	}

	return clients
}

// GetDeleteItems gets delete items for a client
func (ds *DeleteSet) GetDeleteItems(clientId uint64) []DeleteItem {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	if items, exists := ds.clients[clientId]; exists {
		// Return a copy to avoid concurrent access issues
		result := make([]DeleteItem, len(items))
		copy(result, items)
		return result
	}

	return nil
}

// Clear clears all delete items
func (ds *DeleteSet) Clear() {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	ds.clients = make(map[uint64][]DeleteItem)
}
