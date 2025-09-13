// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"errors"
	"fmt"

	"github.com/chenrensong/ygo/contracts"
)

// PendingClientStructRef represents pending client struct reference
type PendingClientStructRef struct {
	NextReadOperation int
	Refs              []contracts.IStructItem
}

// StructStore represents the structure store
type StructStore struct {
	clients                 map[int64][]contracts.IStructItem
	pendingClientStructRefs map[int64]*PendingClientStructRef
	pendingStack            []contracts.IStructItem
	pendingDeleteReaders    []contracts.IDSDecoder
}

// NewStructStore creates a new StructStore
func NewStructStore() *StructStore {
	return &StructStore{
		clients:                 make(map[int64][]contracts.IStructItem),
		pendingClientStructRefs: make(map[int64]*PendingClientStructRef),
		pendingStack:            make([]contracts.IStructItem, 0),
		pendingDeleteReaders:    make([]contracts.IDSDecoder, 0),
	}
}

// GetClients returns the clients map
func (ss *StructStore) GetClients() map[int64][]contracts.IStructItem {
	return ss.clients
}

// GetStateVector returns the states as a Map<int64,int64>.
// Note that clock refers to the next expected clock id.
func (ss *StructStore) GetStateVector() map[int64]int64 {
	result := make(map[int64]int64, len(ss.clients))

	for clientId, structs := range ss.clients {
		if len(structs) > 0 {
			lastStruct := structs[len(structs)-1]
			result[clientId] = lastStruct.GetID().Clock + int64(lastStruct.GetLength())
		}
	}

	return result
}

// GetState returns the state for a specific client
func (ss *StructStore) GetState(clientId int64) int64 {
	if structs, exists := ss.clients[clientId]; exists && len(structs) > 0 {
		lastStruct := structs[len(structs)-1]
		return lastStruct.GetID().Clock + int64(lastStruct.GetLength())
	}
	return 0
}

// IntegrityCheck performs integrity check on the store
func (ss *StructStore) IntegrityCheck() error {
	for clientId, structs := range ss.clients {
		if len(structs) == 0 {
			return fmt.Errorf("StructStore failed integrity check: no structs for client %d", clientId)
		}

		for i := 1; i < len(structs); i++ {
			left := structs[i-1]
			right := structs[i]

			if left.GetID().Clock+int64(left.GetLength()) != right.GetID().Clock {
				return fmt.Errorf("StructStore failed integrity check: missing struct for client %d", clientId)
			}
		}
	}

	if len(ss.pendingDeleteReaders) != 0 || len(ss.pendingStack) != 0 || len(ss.pendingClientStructRefs) != 0 {
		return errors.New("StructStore failed integrity check: still have pending items")
	}

	return nil
}

// CleanupPendingStructs cleans up pending structs
func (ss *StructStore) CleanupPendingStructs() {
	var clientsToRemove []int64

	// Cleanup pendingClientStructRefs if not fully finished
	for client, refs := range ss.pendingClientStructRefs {
		if refs.NextReadOperation == len(refs.Refs) {
			clientsToRemove = append(clientsToRemove, client)
		} else {
			// Remove processed refs
			refs.Refs = refs.Refs[refs.NextReadOperation:]
			refs.NextReadOperation = 0
		}
	}

	for _, client := range clientsToRemove {
		delete(ss.pendingClientStructRefs, client)
	}
}

// AddStruct adds a struct to the store
func (ss *StructStore) AddStruct(str contracts.IStructItem) error {
	clientId := str.GetID().Client
	structs, exists := ss.clients[clientId]

	if !exists {
		structs = make([]contracts.IStructItem, 0)
		ss.clients[clientId] = structs
	} else if len(structs) > 0 {
		lastStruct := structs[len(structs)-1]
		if lastStruct.GetID().Clock+int64(lastStruct.GetLength()) != str.GetID().Clock {
			return errors.New("unexpected struct order")
		}
	}

	ss.clients[clientId] = append(structs, str)
	return nil
}

// MergeReadStructsIntoPendingReads merges read structs into pending reads
func (ss *StructStore) MergeReadStructsIntoPendingReads(clientStructRefs map[int64][]contracts.IStructItem) {
	for client, refs := range clientStructRefs {
		if existing, exists := ss.pendingClientStructRefs[client]; exists {
			// Merge with existing refs
			existing.Refs = append(existing.Refs, refs...)
		} else {
			ss.pendingClientStructRefs[client] = &PendingClientStructRef{
				NextReadOperation: 0,
				Refs:              refs,
			}
		}
	}
}

// ResumeStructIntegration resumes struct integration
func (ss *StructStore) ResumeStructIntegration(transaction contracts.ITransaction) {
	// Process pending structs
	for len(ss.pendingStack) > 0 {
		// Pop from stack
		item := ss.pendingStack[len(ss.pendingStack)-1]
		ss.pendingStack = ss.pendingStack[:len(ss.pendingStack)-1]

		// Try to integrate
		item.Integrate(transaction, 0)
	}

	// Process pending client struct refs
	for _, refs := range ss.pendingClientStructRefs {
		for refs.NextReadOperation < len(refs.Refs) {
			item := refs.Refs[refs.NextReadOperation]
			refs.NextReadOperation++

			// Try to integrate
			item.Integrate(transaction, 0)
		}
	}
}

// TryResumePendingDeleteReaders tries to resume pending delete readers
func (ss *StructStore) TryResumePendingDeleteReaders(transaction contracts.ITransaction) {
	// Process pending delete readers
	var remainingReaders []contracts.IDSDecoder

	for _, reader := range ss.pendingDeleteReaders {
		// Try to process the delete reader
		// This is a simplified implementation
		if ss.canProcessDeleteReader(reader) {
			ss.processDeleteReader(reader, transaction)
		} else {
			remainingReaders = append(remainingReaders, reader)
		}
	}

	ss.pendingDeleteReaders = remainingReaders
}

// canProcessDeleteReader checks if a delete reader can be processed
func (ss *StructStore) canProcessDeleteReader(reader contracts.IDSDecoder) bool {
	// Simplified check - in real implementation this would check dependencies
	return true
}

// processDeleteReader processes a delete reader
func (ss *StructStore) processDeleteReader(reader contracts.IDSDecoder, transaction contracts.ITransaction) {
	// Simplified implementation - would read and apply deletes
	// This would typically read the delete set and mark items as deleted
}

// GetStruct gets a struct by ID
func (ss *StructStore) GetStruct(id contracts.StructID) contracts.IStructItem {
	if structs, exists := ss.clients[id.Client]; exists {
		// Binary search for the struct
		for _, item := range structs {
			if item.GetID().Clock <= id.Clock && id.Clock < item.GetID().Clock+int64(item.GetLength()) {
				return item
			}
		}
	}
	return nil
}

// GetStructs gets all structs for a client
func (ss *StructStore) GetStructs(client int64) []contracts.IStructItem {
	if structs, exists := ss.clients[client]; exists {
		return structs
	}
	return make([]contracts.IStructItem, 0)
}
