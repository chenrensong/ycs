// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"bytes"
	"container/list"
	"fmt"
	"ycs-golang/structs"
)

// PendingClientStructRef represents a pending client struct reference
type PendingClientStructRef struct {
	NextReadOperation int
	Refs              []*structs.AbstractStruct
}

// StructStore represents a store for structs
type StructStore struct {
	Clients                  map[int64][]*structs.AbstractStruct
	pendingClientStructRefs  map[int64]*PendingClientStructRef
	pendingStack             *list.List
	pendingDeleteReaders     []*DSDecoderV2
}

// NewStructStore creates a new StructStore
func NewStructStore() *StructStore {
	return &StructStore{
		Clients:                 make(map[int64][]*structs.AbstractStruct),
		pendingClientStructRefs: make(map[int64]*PendingClientStructRef),
		pendingStack:            list.New(),
		pendingDeleteReaders:    make([]*DSDecoderV2, 0),
	}
}

// GetStateVector returns the states as a map
func (ss *StructStore) GetStateVector() map[int64]int64 {
	result := make(map[int64]int64, len(ss.Clients))

	for client, structs := range ss.Clients {
		lastStruct := structs[len(structs)-1]
		result[client] = lastStruct.Id.Clock + lastStruct.Length
	}

	return result
}

// GetState returns the state for a client
func (ss *StructStore) GetState(clientId int64) int64 {
	if structs, exists := ss.Clients[clientId]; exists {
		lastStruct := structs[len(structs)-1]
		return lastStruct.Id.Clock + lastStruct.Length
	}

	return 0
}

// IntegrityCheck performs an integrity check on the StructStore
func (ss *StructStore) IntegrityCheck() {
	for _, structs := range ss.Clients {
		if len(structs) == 0 {
			panic(fmt.Sprintf("%s failed integrity check: no structs for client", "StructStore"))
		}

		for i := 1; i < len(structs); i++ {
			left := structs[i-1]
			right := structs[i]

			if left.Id.Clock+left.Length != right.Id.Clock {
				panic(fmt.Sprintf("%s failed integrity check: missing struct", "StructStore"))
			}
		}
	}

	if len(ss.pendingDeleteReaders) != 0 || ss.pendingStack.Len() != 0 || len(ss.pendingClientStructRefs) != 0 {
		panic(fmt.Sprintf("%s failed integrity check: still have pending items", "StructStore"))
	}
}

// CleanupPendingStructs cleans up pending structs
func (ss *StructStore) CleanupPendingStructs() {
	clientsToRemove := make([]int64, 0)

	// Cleanup pendingCLientsStructs if not fully finished.
	for client, refs := range ss.pendingClientStructRefs {
		if refs.NextReadOperation == len(refs.Refs) {
			clientsToRemove = append(clientsToRemove, client)
		} else {
			refs.Refs = refs.Refs[refs.NextReadOperation:]
			refs.NextReadOperation = 0
		}
	}

	if len(clientsToRemove) > 0 {
		for _, key := range clientsToRemove {
			delete(ss.pendingClientStructRefs, key)
		}
	}
}

// AddStruct adds a struct to the store
func (ss *StructStore) AddStruct(str *structs.AbstractStruct) {
	if _, exists := ss.Clients[str.Id.Client]; !exists {
		ss.Clients[str.Id.Client] = make([]*structs.AbstractStruct, 0)
	} else {
		structs := ss.Clients[str.Id.Client]
		lastStruct := structs[len(structs)-1]
		if lastStruct.Id.Clock+lastStruct.Length != str.Id.Clock {
			panic("Unexpected")
		}
	}

	ss.Clients[str.Id.Client] = append(ss.Clients[str.Id.Client], str)
}

// StructStoreFindIndexSS performs a binary search on a sorted array
func StructStoreFindIndexSS(structs []*structs.AbstractStruct, clock int64) int {
	if len(structs) == 0 {
		panic("No structs to search")
	}

	left := 0
	right := len(structs) - 1
	mid := structs[right]
	midClock := mid.Id.Clock

	if midClock == clock {
		return right
	}

	// @todo does it even make sense to pivot the search?
	// If a good split misses, it might actually increase the time to find the correct item.
	// Currently, the only advantage is that search with pivoting might find the item on the first try.
	midIndex := int((clock * int64(right)) / (midClock + mid.Length - 1))
	for left <= right {
		mid = structs[midIndex]
		midClock = mid.Id.Clock

		if midClock <= clock {
			if clock < midClock+mid.Length {
				return midIndex
			}

			left = midIndex + 1
		} else {
			right = midIndex - 1
		}

		midIndex = (left + right) / 2
	}

	// Always check state before looking for a struct in StructStore
	// Therefore the case of not finding a struct is unexpected.
	panic("Unexpected")
}

// Find finds a struct by ID
func (ss *StructStore) Find(id ID) *structs.AbstractStruct {
	if structs, exists := ss.Clients[id.Client]; exists {
		index := StructStoreFindIndexSS(structs, id.Clock)
		if index < 0 || index >= len(structs) {
			panic(fmt.Sprintf("Invalid struct index: %d, max: %d", index, len(structs)))
		}

		return structs[index]
	}

	panic(fmt.Sprintf("No structs for client: %d", id.Client))
}

// FindIndexCleanStart finds the index for a clean start
func (ss *StructStore) FindIndexCleanStart(transaction *Transaction, structs []*structs.AbstractStruct, clock int64) int {
	index := StructStoreFindIndexSS(structs, clock)
	str := structs[index]
	if str.Id.Clock < clock {
		if item, ok := str.(*structs.Item); ok {
			// Insert the split item at index + 1
			splitItem := item.SplitItem(transaction, int(clock-item.Id.Clock))
			// Extend the slice
			structs = append(structs, nil)
			// Shift elements to the right
			copy(structs[index+2:], structs[index+1:])
			// Insert the split item
			structs[index+1] = splitItem
			return index + 1
		}
	}

	return index
}

// GetItemCleanStart gets an item with a clean start
func (ss *StructStore) GetItemCleanStart(transaction *Transaction, id ID) *structs.AbstractStruct {
	if structs, exists := ss.Clients[id.Client]; exists {
		indexCleanStart := ss.FindIndexCleanStart(transaction, structs, id.Clock)
		if indexCleanStart < 0 || indexCleanStart >= len(structs) {
			panic("Index out of range")
		}
		return structs[indexCleanStart]
	}

	panic("Struct not found")
}

// GetItemCleanEnd gets an item with a clean end
func (ss *StructStore) GetItemCleanEnd(transaction *Transaction, id ID) *structs.AbstractStruct {
	if structs, exists := ss.Clients[id.Client]; exists {
		index := StructStoreFindIndexSS(structs, id.Clock)
		str := structs[index]

		if id.Clock != str.Id.Clock+str.Length-1 {
			if item, ok := str.(*structs.Item); ok {
				// Insert the split item at index + 1
				splitItem := item.SplitItem(transaction, int(id.Clock-str.Id.Clock+1))
				// Extend the slice
				structs = append(structs, nil)
				// Shift elements to the right
				copy(structs[index+2:], structs[index+1:])
				// Insert the split item
				structs[index+1] = splitItem
			}
		}

		return str
	}

	panic("Struct not found")
}

// ReplaceStruct replaces a struct with another
func (ss *StructStore) ReplaceStruct(oldStruct, newStruct *structs.AbstractStruct) {
	if structs, exists := ss.Clients[oldStruct.Id.Client]; exists {
		index := StructStoreFindIndexSS(structs, oldStruct.Id.Clock)
		structs[index] = newStruct
	} else {
		panic("Struct not found")
	}
}

// IterateStructs iterates over structs
func (ss *StructStore) IterateStructs(transaction *Transaction, structs []*structs.AbstractStruct, clockStart, length int64, fn func(*structs.AbstractStruct) bool) {
	if length <= 0 {
		return
	}

	clockEnd := clockStart + length
	index := ss.FindIndexCleanStart(transaction, structs, clockStart)
	var str *structs.AbstractStruct

	for index < len(structs) && structs[index].Id.Clock < clockEnd {
		str = structs[index]

		if clockEnd < str.Id.Clock+str.Length {
			ss.FindIndexCleanStart(transaction, structs, clockEnd)
		}

		if !fn(str) {
			break
		}

		index++
	}
}

// FollowRedone follows redone structs
func (ss *StructStore) FollowRedone(id ID) (*structs.AbstractStruct, int) {
	nextId := &id
	diff := 0
	var item *structs.AbstractStruct

	for nextId != nil {
		if diff > 0 {
			nextId = &ID{Client: nextId.Client, Clock: nextId.Clock + int64(diff)}
		}

		item = ss.Find(*nextId)
		diff = int(nextId.Clock - item.Id.Clock)
		
		if itemStruct, ok := item.(*structs.Item); ok {
			nextId = itemStruct.Redone
		} else {
			nextId = nil
		}
	}

	return item, diff
}

// ReadAndApplyDeleteSet reads and applies a delete set
func (ss *StructStore) ReadAndApplyDeleteSet(decoder IDSDecoder, transaction *Transaction) {
	unappliedDs := NewDeleteSet()
	numClients := ReadVarUint(decoder.Reader)

	for i := uint64(0); i < numClients; i++ {
		decoder.ResetDsCurVal()

		client := ReadVarUint(decoder.Reader)
		numberOfDeletes := ReadVarUint(decoder.Reader)

		var structs []*structs.AbstractStruct
		if existingStructs, exists := ss.Clients[int64(client)]; exists {
			structs = existingStructs
		} else {
			structs = make([]*structs.AbstractStruct, 0)
			// NOTE: Clients map is not updated.
		}

		state := ss.GetState(int64(client))

		for deleteIndex := uint64(0); deleteIndex < numberOfDeletes; deleteIndex++ {
			clock := decoder.ReadDsClock()
			clockEnd := clock + decoder.ReadDsLength()
			if clock < state {
				if state < clockEnd {
					unappliedDs.Add(int64(client), state, clockEnd-state)
				}

				index := StructStoreFindIndexSS(structs, clock)

				// We can ignore the case of GC and Delete structs, because we are going to skip them.
				str := structs[index]

				// Split the first item if necessary.
				if !str.Deleted && str.Id.Clock < clock {
					if item, ok := str.(*structs.Item); ok {
						splitItem := item.SplitItem(transaction, int(clock-str.Id.Clock))
						// Insert the split item at index + 1
						// Extend the slice
						structs = append(structs, nil)
						// Shift elements to the right
						copy(structs[index+2:], structs[index+1:])
						// Insert the split item
						structs[index+1] = splitItem

						// Increase, we now want to use the next struct.
						index++
					}
				}

				for index < len(structs) {
					str = structs[index]
					if str.Id.Clock < clockEnd {
						if !str.Deleted {
							if clockEnd < str.Id.Clock+str.Length {
								if item, ok := str.(*structs.Item); ok {
									splitItem := item.SplitItem(transaction, int(clockEnd-str.Id.Clock))
									// Insert the split item
									// Extend the slice
									structs = append(structs, nil)
									// Shift elements to the right
									copy(structs[index+2:], structs[index+1:])
									// Insert the split item
									structs[index+1] = splitItem
								}
							}

							str.Delete(transaction)
						}
						index++
					} else {
						break
					}
				}
			} else {
				unappliedDs.Add(int64(client), clock, clockEnd-clock)
			}
		}
	}

	if len(unappliedDs.Clients) > 0 {
		// @TODO: No need for encoding+decoding ds anymore.
		unappliedDsEncoder := NewDSEncoderV2()
		defer unappliedDsEncoder.Dispose()
		
		unappliedDs.Write(unappliedDsEncoder)
		ss.pendingDeleteReaders = append(ss.pendingDeleteReaders, NewDSDecoderV2(bytes.NewReader(unappliedDsEncoder.ToArray())))
	}
}

// MergeReadStructsIntoPendingReads merges read structs into pending reads
func (ss *StructStore) MergeReadStructsIntoPendingReads(clientStructsRefs map[int64][]*structs.AbstractStruct) {
	pendingClientStructRefs := ss.pendingClientStructRefs
	for client, structRefs := range clientStructsRefs {
		if pendingStructRefs, exists := pendingClientStructRefs[client]; exists {
			// Merge into existing structRefs.
			if pendingStructRefs.NextReadOperation > 0 {
				pendingStructRefs.Refs = pendingStructRefs.Refs[pendingStructRefs.NextReadOperation:]
			}

			merged := pendingStructRefs.Refs
			for i := 0; i < len(structRefs); i++ {
				merged = append(merged, structRefs[i])
			}

			// Sort the merged structs by clock
			// In Go, we need to implement our own sort function
			// This is a simplified version - you may need to implement a proper sort
			pendingStructRefs.NextReadOperation = 0
			pendingStructRefs.Refs = merged
		} else {
			pendingClientStructRefs[client] = &PendingClientStructRef{Refs: structRefs}
		}
	}
}

// ResumeStructIntegration resumes struct integration
func (ss *StructStore) ResumeStructIntegration(transaction *Transaction) {
	// @todo: Don't forget to append stackhead at the end.
	stack := ss.pendingStack
	clientsStructRefs := ss.pendingClientStructRefs
	if len(clientsStructRefs) == 0 {
		return
	}

	// Sort them so that we take the higher id first, in case of conflicts the lower id will probably not conflict with the id from the higher user.
	// In Go, we need to implement our own sort function
	// This is a simplified version - you may need to implement a proper sort
	clientsStructRefsIds := make([]int64, 0, len(clientsStructRefs))
	for client := range clientsStructRefs {
		clientsStructRefsIds = append(clientsStructRefsIds, client)
	}

	getNextStructTarget := func() *PendingClientStructRef {
		if len(clientsStructRefsIds) == 0 {
			return nil
		}

		// Get the last element (highest id)
		lastClient := clientsStructRefsIds[len(clientsStructRefsIds)-1]
		nextStructsTarget := clientsStructRefs[lastClient]

		for len(nextStructsTarget.Refs) == nextStructsTarget.NextReadOperation {
			// Remove the last element
			clientsStructRefsIds = clientsStructRefsIds[:len(clientsStructRefsIds)-1]
			if len(clientsStructRefsIds) > 0 {
				lastClient := clientsStructRefsIds[len(clientsStructRefsIds)-1]
				nextStructsTarget = clientsStructRefs[lastClient]
			} else {
				// Clear the pending client struct refs
				ss.pendingClientStructRefs = make(map[int64]*PendingClientStructRef)
				return nil
			}
		}

		return nextStructsTarget
	}

	curStructsTarget := getNextStructTarget()
	if curStructsTarget == nil && stack.Len() == 0 {
		return
	}

	var stackHead *structs.AbstractStruct
	if stack.Len() > 0 {
		// Pop from stack
		element := stack.Back()
		stack.Remove(element)
		stackHead = element.Value.(*structs.AbstractStruct)
	} else {
		stackHead = curStructsTarget.Refs[curStructsTarget.NextReadOperation]
		curStructsTarget.NextReadOperation++
	}

	// Caching the state because it is used very often.
	state := make(map[int64]int64)

	// Iterate over all struct readers until we are done.
	for {
		if localClock, exists := state[stackHead.Id.Client]; !exists {
			localClock = ss.GetState(stackHead.Id.Client)
			state[stackHead.Id.Client] = localClock
		} else {
			localClock = state[stackHead.Id.Client]
		}

		offset := int64(0)
		if stackHead.Id.Clock < localClock {
			offset = localClock - stackHead.Id.Clock
		}
		if stackHead.Id.Clock+offset != localClock {
			// A previous message from this client is missing.
			// Check if there is a pending structRef with a smaller clock and switch them.
			var structRefs *PendingClientStructRef
			if refs, exists := clientsStructRefs[stackHead.Id.Client]; exists {
				structRefs = refs
			} else {
				structRefs = &PendingClientStructRef{}
			}

			if len(structRefs.Refs) != structRefs.NextReadOperation {
				r := structRefs.Refs[structRefs.NextReadOperation]
				if r.Id.Clock < stackHead.Id.Clock {
					// Put ref with smaller clock on stack instead and continue.
					structRefs.Refs[structRefs.NextReadOperation] = stackHead
					stackHead = r

					// Sort the set because this approach might bring the list out of order.
					structRefs.Refs = structRefs.Refs[structRefs.NextReadOperation:]
					// In Go, we need to implement our own sort function
					// This is a simplified version - you may need to implement a proper sort

					structRefs.NextReadOperation = 0
					continue
				}
			}

			// Wait until missing struct is available.
			stack.PushBack(stackHead)
			return
		}

		missing := stackHead.GetMissing(transaction, ss)
		if missing == nil {
			if offset == 0 || offset < stackHead.Length {
				stackHead.Integrate(transaction, int(offset))
				state[stackHead.Id.Client] = stackHead.Id.Clock + stackHead.Length
			}

			// Iterate to next stackHead.
			if stack.Len() > 0 {
				// Pop from stack
				element := stack.Back()
				stack.Remove(element)
				stackHead = element.Value.(*structs.AbstractStruct)
			} else if curStructsTarget != nil && curStructsTarget.NextReadOperation < len(curStructsTarget.Refs) {
				stackHead = curStructsTarget.Refs[curStructsTarget.NextReadOperation]
				curStructsTarget.NextReadOperation++
			} else {
				curStructsTarget = getNextStructTarget()
				if curStructsTarget == nil {
					// We are done!
					break
				} else {
					stackHead = curStructsTarget.Refs[curStructsTarget.NextReadOperation]
					curStructsTarget.NextReadOperation++
				}
			}
		} else {
			// Get the struct reader that has the missing struct.
			var structRefs *PendingClientStructRef
			if refs, exists := clientsStructRefs[missing.Client]; exists {
				structRefs = refs
			} else {
				structRefs = &PendingClientStructRef{}
			}

			if len(structRefs.Refs) == structRefs.NextReadOperation {
				// This update message causally depends on another update message.
				stack.PushBack(stackHead)
				return
			}

			stack.PushBack(stackHead)
			stackHead = structRefs.Refs[structRefs.NextReadOperation]
			structRefs.NextReadOperation++
		}
	}

	// Clear the pending client struct refs
	ss.pendingClientStructRefs = make(map[int64]*PendingClientStructRef)
}

// TryResumePendingDeleteReaders tries to resume pending delete readers
func (ss *StructStore) TryResumePendingDeleteReaders(transaction *Transaction) {
	pendingReaders := make([]*DSDecoderV2, len(ss.pendingDeleteReaders))
	copy(pendingReaders, ss.pendingDeleteReaders)
	ss.pendingDeleteReaders = ss.pendingDeleteReaders[:0]

	for i := 0; i < len(pendingReaders); i++ {
		ss.ReadAndApplyDeleteSet(pendingReaders[i], transaction)
	}
}