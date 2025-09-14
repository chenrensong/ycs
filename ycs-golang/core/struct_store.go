package core

import (
	"errors"
	"sort"
	"ycs/contracts"
)

// PendingClientStructRef represents pending client struct references
type PendingClientStructRef struct {
	NextReadOperation int
	Refs              []contracts.IStructItem
}

// NewPendingClientStructRef creates a new PendingClientStructRef
func NewPendingClientStructRef() *PendingClientStructRef {
	return &PendingClientStructRef{
		NextReadOperation: 0,
		Refs:              make([]contracts.IStructItem, 0, 1),
	}
}

// StructStore manages the storage of structs organized by client
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

// GetStateVector returns the states as a map where clock refers to the next expected clock id
func (ss *StructStore) GetStateVector() map[int64]int64 {
	result := make(map[int64]int64, len(ss.clients))

	for client, structs := range ss.clients {
		if len(structs) > 0 {
			lastStruct := structs[len(structs)-1]
			result[client] = lastStruct.GetID().Clock + int64(lastStruct.GetLength())
		}
	}

	return result
}

// GetState returns the state for a specific client
func (ss *StructStore) GetState(clientID int64) int64 {
	if structs, exists := ss.clients[clientID]; exists && len(structs) > 0 {
		lastStruct := structs[len(structs)-1]
		return lastStruct.GetID().Clock + int64(lastStruct.GetLength())
	}
	return 0
}

// IntegrityCheck performs integrity check on the store
func (ss *StructStore) IntegrityCheck() error {
	for client, structs := range ss.clients {
		if len(structs) == 0 {
			return errors.New("StructStore failed integrity check: no structs for client")
		}

		for i := 1; i < len(structs); i++ {
			left := structs[i-1]
			right := structs[i]

			if left.GetID().Clock+int64(left.GetLength()) != right.GetID().Clock {
				return errors.New("StructStore failed integrity check: missing struct")
			}
		}
	}

	if len(ss.pendingDeleteReaders) != 0 || len(ss.pendingStack) != 0 || len(ss.pendingClientStructRefs) != 0 {
		return errors.New("StructStore failed integrity check: still have pending items")
	}

	return nil
}

// CleanupPendingStructs cleans up pending structs if not fully finished
func (ss *StructStore) CleanupPendingStructs() {
	var clientsToRemove []int64

	for client, refs := range ss.pendingClientStructRefs {
		if refs.NextReadOperation == len(refs.Refs) {
			clientsToRemove = append(clientsToRemove, client)
		} else {
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
	client := str.GetID().Client
	structs, exists := ss.clients[client]

	if !exists {
		structs = make([]contracts.IStructItem, 0)
		ss.clients[client] = structs
	} else if len(structs) > 0 {
		lastStruct := structs[len(structs)-1]
		if lastStruct.GetID().Clock+int64(lastStruct.GetLength()) != str.GetID().Clock {
			return errors.New("unexpected struct clock")
		}
	}

	ss.clients[client] = append(structs, str)
	return nil
}

// Find finds a struct by ID
func (ss *StructStore) Find(id StructID) (contracts.IStructItem, error) {
	structs, exists := ss.clients[id.Client]
	if !exists {
		return nil, errors.New("no structs for client")
	}

	index := FindIndexSS(structs, id.Clock)
	if index < 0 || index >= len(structs) {
		return nil, errors.New("invalid struct index")
	}

	return structs[index], nil
}

// FindIndexCleanStart finds index with clean start
func (ss *StructStore) FindIndexCleanStart(transaction contracts.ITransaction, structs []contracts.IStructItem, clock int64) int {
	index := FindIndexSS(structs, clock)
	str := structs[index]

	if str.GetID().Clock < clock {
		splitItem := str.SplitItem(transaction, int(clock-str.GetID().Clock))
		// Insert the split item
		newStructs := make([]contracts.IStructItem, len(structs)+1)
		copy(newStructs[:index+1], structs[:index+1])
		newStructs[index+1] = splitItem
		copy(newStructs[index+2:], structs[index+1:])

		// Update the clients map
		ss.clients[str.GetID().Client] = newStructs
		return index + 1
	}

	return index
}

// GetItemCleanStart gets item with clean start
func (ss *StructStore) GetItemCleanStart(transaction contracts.ITransaction, id StructID) (contracts.IStructItem, error) {
	structs, exists := ss.clients[id.Client]
	if !exists {
		return nil, errors.New("no structs for client")
	}

	indexCleanStart := ss.FindIndexCleanStart(transaction, structs, id.Clock)
	if indexCleanStart < 0 || indexCleanStart >= len(structs) {
		return nil, errors.New("invalid index")
	}

	return structs[indexCleanStart], nil
}

// GetItemCleanEnd gets item with clean end
func (ss *StructStore) GetItemCleanEnd(transaction contracts.ITransaction, id StructID) (contracts.IStructItem, error) {
	structs, exists := ss.clients[id.Client]
	if !exists {
		return nil, errors.New("no structs for client")
	}

	index := FindIndexSS(structs, id.Clock)
	str := structs[index]

	if id.Clock != str.GetID().Clock+int64(str.GetLength())-1 && !str.IsGC() {
		splitItem := str.SplitItem(transaction, int(id.Clock-str.GetID().Clock+1))
		// Insert the split item
		newStructs := make([]contracts.IStructItem, len(structs)+1)
		copy(newStructs[:index+1], structs[:index+1])
		newStructs[index+1] = splitItem
		copy(newStructs[index+2:], structs[index+1:])

		// Update the clients map
		ss.clients[str.GetID().Client] = newStructs
	}

	return str, nil
}

// ReplaceStruct replaces an old struct with a new one
func (ss *StructStore) ReplaceStruct(oldStruct, newStruct contracts.IStructItem) error {
	structs, exists := ss.clients[oldStruct.GetID().Client]
	if !exists {
		return errors.New("no structs for client")
	}

	index := FindIndexSS(structs, oldStruct.GetID().Clock)
	structs[index] = newStruct
	return nil
}

// IterateStructs iterates over structs in a range
func (ss *StructStore) IterateStructs(transaction contracts.ITransaction, structs []contracts.IStructItem, clockStart int64, length int64, fun func(contracts.IStructItem) bool) {
	if length <= 0 {
		return
	}

	clockEnd := clockStart + length
	index := ss.FindIndexCleanStart(transaction, structs, clockStart)

	for index < len(structs) {
		str := structs[index]

		if clockEnd < str.GetID().Clock+int64(str.GetLength()) {
			ss.FindIndexCleanStart(transaction, structs, clockEnd)
		}

		if !fun(str) {
			break
		}

		index++
		if index < len(structs) && structs[index].GetID().Clock >= clockEnd {
			break
		}
	}
}

// FollowRedone follows redone items
func (ss *StructStore) FollowRedone(id StructID) (contracts.IStructItem, int, error) {
	nextID := id
	diff := 0

	for {
		if diff > 0 {
			nextID = StructID{Client: nextID.Client, Clock: nextID.Clock + int64(diff)}
		}

		item, err := ss.Find(nextID)
		if err != nil {
			return nil, 0, err
		}

		diff = int(nextID.Clock - item.GetID().Clock)

		if item.GetRedone() == nil {
			return item, diff, nil
		}

		nextID = *item.GetRedone()
	}
}

// MergeReadStructsIntoPendingReads merges read structs into pending reads
func (ss *StructStore) MergeReadStructsIntoPendingReads(clientStructRefs map[int64][]contracts.IStructItem) {
	for client, structRefs := range clientStructRefs {
		pendingStructRefs, exists := ss.pendingClientStructRefs[client]
		if !exists {
			ss.pendingClientStructRefs[client] = &PendingClientStructRef{
				NextReadOperation: 0,
				Refs:              structRefs,
			}
		} else {
			// Merge into existing structRefs
			if pendingStructRefs.NextReadOperation > 0 {
				pendingStructRefs.Refs = pendingStructRefs.Refs[pendingStructRefs.NextReadOperation:]
			}

			merged := append(pendingStructRefs.Refs, structRefs...)
			sort.Slice(merged, func(i, j int) bool {
				return merged[i].GetID().Clock < merged[j].GetID().Clock
			})

			pendingStructRefs.NextReadOperation = 0
			pendingStructRefs.Refs = merged
		}
	}
}

// ResumeStructIntegration resumes computing structs generated by struct readers
func (ss *StructStore) ResumeStructIntegration(transaction contracts.ITransaction) {
	stack := ss.pendingStack
	clientsStructRefs := ss.pendingClientStructRefs

	if len(clientsStructRefs) == 0 {
		return
	}

	// Sort client IDs
	var clientsStructRefsIds []int64
	for client := range clientsStructRefs {
		clientsStructRefsIds = append(clientsStructRefsIds, client)
	}
	sort.Slice(clientsStructRefsIds, func(i, j int) bool {
		return clientsStructRefsIds[i] < clientsStructRefsIds[j]
	})

	getNextStructTarget := func() *PendingClientStructRef {
		if len(clientsStructRefsIds) == 0 {
			return nil
		}

		nextStructsTarget := clientsStructRefs[clientsStructRefsIds[len(clientsStructRefsIds)-1]]

		for len(nextStructsTarget.Refs) == nextStructsTarget.NextReadOperation {
			clientsStructRefsIds = clientsStructRefsIds[:len(clientsStructRefsIds)-1]
			if len(clientsStructRefsIds) > 0 {
				nextStructsTarget = clientsStructRefs[clientsStructRefsIds[len(clientsStructRefsIds)-1]]
			} else {
				ss.pendingClientStructRefs = make(map[int64]*PendingClientStructRef)
				return nil
			}
		}

		return nextStructsTarget
	}

	curStructsTarget := getNextStructTarget()
	if curStructsTarget == nil && len(stack) == 0 {
		return
	}

	var stackHead contracts.IStructItem
	if len(stack) > 0 {
		stackHead = stack[len(stack)-1]
		ss.pendingStack = stack[:len(stack)-1]
	} else {
		stackHead = curStructsTarget.Refs[curStructsTarget.NextReadOperation]
		curStructsTarget.NextReadOperation++
	}

	// Caching the state because it is used very often
	state := make(map[int64]int64)

	// Iterate over all struct readers until we are done
	for {
		localClock, exists := state[stackHead.GetID().Client]
		if !exists {
			localClock = ss.GetState(stackHead.GetID().Client)
			state[stackHead.GetID().Client] = localClock
		}

		offset := int64(0)
		if stackHead.GetID().Clock < localClock {
			offset = localClock - stackHead.GetID().Clock
		}

		if stackHead.GetID().Clock+offset != localClock {
			// A previous message from this client is missing
			// Check if there is a pending structRef with a smaller clock and switch them
			structRefs, exists := clientsStructRefs[stackHead.GetID().Client]
			if !exists {
				structRefs = NewPendingClientStructRef()
			}

			if len(structRefs.Refs) != structRefs.NextReadOperation {
				r := structRefs.Refs[structRefs.NextReadOperation]
				if r.GetID().Clock < stackHead.GetID().Clock {
					// Put ref with smaller clock on stack instead and continue
					structRefs.Refs[structRefs.NextReadOperation] = stackHead
					stackHead = r

					// Sort the set because this approach might bring the list out of order
					refs := structRefs.Refs[structRefs.NextReadOperation:]
					sort.Slice(refs, func(i, j int) bool {
						return refs[i].GetID().Clock < refs[j].GetID().Clock
					})
					structRefs.Refs = structRefs.Refs[:structRefs.NextReadOperation]
					structRefs.Refs = append(structRefs.Refs, refs...)
					structRefs.NextReadOperation = 0
					continue
				}
			}

			// Wait until missing struct is available
			ss.pendingStack = append(ss.pendingStack, stackHead)
			return
		}

		missing := stackHead.GetMissing(transaction, ss)
		if missing == nil {
			if offset == 0 || offset < int64(stackHead.GetLength()) {
				stackHead.Integrate(transaction, int(offset))
				state[stackHead.GetID().Client] = stackHead.GetID().Clock + int64(stackHead.GetLength())
			}

			// Iterate to next stackHead
			if len(ss.pendingStack) > 0 {
				stackHead = ss.pendingStack[len(ss.pendingStack)-1]
				ss.pendingStack = ss.pendingStack[:len(ss.pendingStack)-1]
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
			// Get the struct reader that has the missing struct
			structRefs, exists := clientsStructRefs[*missing]
			if !exists {
				structRefs = NewPendingClientStructRef()
			}

			if len(structRefs.Refs) == structRefs.NextReadOperation {
				// This update message causally depends on another update message
				ss.pendingStack = append(ss.pendingStack, stackHead)
				return
			}

			ss.pendingStack = append(ss.pendingStack, stackHead)
			stackHead = structRefs.Refs[structRefs.NextReadOperation]
			structRefs.NextReadOperation++
		}
	}

	ss.pendingClientStructRefs = make(map[int64]*PendingClientStructRef)
}

// ReadAndApplyDeleteSet reads and applies delete set
func (ss *StructStore) ReadAndApplyDeleteSet(decoder contracts.IDSDecoder, transaction contracts.ITransaction) error {
	unappliedDs := NewDeleteSet()
	numClients := decoder.GetReader().ReadVarUint()

	for i := uint64(0); i < numClients; i++ {
		decoder.ResetDsCurVal()

		client := int64(decoder.GetReader().ReadVarUint())
		numberOfDeletes := decoder.GetReader().ReadVarUint()

		structs, exists := ss.clients[client]
		if !exists {
			structs = make([]contracts.IStructItem, 0)
			// NOTE: Clients map is not updated
		}

		state := ss.GetState(client)

		for deleteIndex := uint64(0); deleteIndex < numberOfDeletes; deleteIndex++ {
			clock := int64(decoder.ReadDsClock())
			clockEnd := clock + int64(decoder.ReadDsLength())

			if clock < state {
				if state < clockEnd {
					unappliedDs.Add(client, state, clockEnd-state)
				}

				index := FindIndexSS(structs, clock)

				// We can ignore the case of GC and Delete structs, because we are going to skip them
				str := structs[index]

				// Split the first item if necessary
				if !str.IsDeleted() && str.GetID().Clock < clock {
					splitItem := str.SplitItem(transaction, int(clock-str.GetID().Clock))
					// Insert split item
					newStructs := make([]contracts.IStructItem, len(structs)+1)
					copy(newStructs[:index+1], structs[:index+1])
					newStructs[index+1] = splitItem
					copy(newStructs[index+2:], structs[index+1:])
					structs = newStructs
					ss.clients[client] = structs

					// Increase, we now want to use the next struct
					index++
				}

				for index < len(structs) {
					str = structs[index]
					if str.GetID().Clock < clockEnd {
						if !str.IsDeleted() {
							if clockEnd < str.GetID().Clock+int64(str.GetLength()) {
								splitItem := str.SplitItem(transaction, int(clockEnd-str.GetID().Clock))
								// Insert split item
								newStructs := make([]contracts.IStructItem, len(structs)+1)
								copy(newStructs[:index+1], structs[:index+1])
								newStructs[index+1] = splitItem
								copy(newStructs[index+2:], structs[index+1:])
								structs = newStructs
								ss.clients[client] = structs
							}

							str.Delete(transaction)
						}
						index++
					} else {
						break
					}
				}
			} else {
				unappliedDs.Add(client, clock, clockEnd-clock)
			}
		}
	}

	if len(unappliedDs.GetClients()) > 0 {
		// Create encoder for unapplied delete set
		encoder := NewDSEncoderV2()
		unappliedDs.Write(encoder)

		// Create decoder from encoded data
		decoder := NewDSDecoderV2FromBytes(encoder.ToArray())
		ss.pendingDeleteReaders = append(ss.pendingDeleteReaders, decoder)
	}

	return nil
}

// TryResumePendingDeleteReaders tries to resume pending delete readers
func (ss *StructStore) TryResumePendingDeleteReaders(transaction contracts.ITransaction) {
	pendingReaders := make([]contracts.IDSDecoder, len(ss.pendingDeleteReaders))
	copy(pendingReaders, ss.pendingDeleteReaders)
	ss.pendingDeleteReaders = ss.pendingDeleteReaders[:0]

	for _, reader := range pendingReaders {
		ss.ReadAndApplyDeleteSet(reader, transaction)
	}
}

// TryToMergeWithLeft tries to merge struct with left neighbor
func TryToMergeWithLeft(structs []contracts.IStructItem, pos int) {
	if pos <= 0 || pos >= len(structs) {
		return
	}

	left := structs[pos-1]
	right := structs[pos]

	if left.TryToMergeWithRight(right) {
		// Remove the right item since it was merged
		copy(structs[pos:], structs[pos+1:])
		structs = structs[:len(structs)-1]
	}
}
