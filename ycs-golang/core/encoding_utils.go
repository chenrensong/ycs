// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"errors"
	"fmt"
	"sort"

	"github.com/chenrensong/ygo/contracts"
	"github.com/chenrensong/ygo/lib0"
)

// EncodingUtils provides utilities for encoding and decoding structs
// We use first five bits in the info flag for determining the type of the struct.
// 0: GC
// 1: Deleted content
// 2: JSON content
// 3: Binary content
// 4: String content
// 5: Embed content (for richtext content)
// 6: Format content (a formatting marker for richtext content)
// 7: Type content
// 8: Any content
// 9: Doc content
type EncodingUtils struct{}

var contentReaderRegistry contracts.IContentReaderRegistry

// SetContentReaderRegistry sets the content reader registry
func SetContentReaderRegistry(registry contracts.IContentReaderRegistry) {
	contentReaderRegistry = registry
}

// ReadItemContent reads item content from decoder
func ReadItemContent(decoder contracts.IUpdateDecoder, info byte) (contracts.IContent, error) {
	contentRef := int(info & byte(lib0.Bits5))
	if contentRef == 0 { // GC
		return nil, errors.New("GC is not ItemContent")
	}

	if contentReaderRegistry == nil {
		return nil, errors.New("ContentReaderRegistry not initialized. Call YcsBootstrap.Initialize() first")
	}

	return contentReaderRegistry.ReadContent(contentRef, decoder), nil
}

// ReadStructs reads the next Item in a Decoder and fills this Item with the read data.
// This is called when data is received from a remote peer.
func ReadStructs(decoder contracts.IUpdateDecoder, transaction contracts.ITransaction, store contracts.IStructStore) error {
	clientStructRefs, err := ReadClientStructRefs(decoder, transaction.GetDoc())
	if err != nil {
		return err
	}

	store.MergeReadStructsIntoPendingReads(clientStructRefs)
	store.ResumeStructIntegration(transaction)
	store.CleanupPendingStructs()
	store.TryResumePendingDeleteReaders(transaction)

	return nil
}

// WriteStructs writes structs starting with ID(client,clock)
func WriteStructs(encoder contracts.IUpdateEncoder, structs []contracts.IStructItem, client int64, clock int64) error {
	// Write first id
	startNewStructs := FindIndexSS(structs, clock)

	// Write # encoded structs
	encoder.GetRestWriter().WriteVarUint(uint64(len(structs) - startNewStructs))
	encoder.WriteClient(client)
	encoder.GetRestWriter().WriteVarUint(uint64(clock))

	// Write first struct with offset
	firstStruct := structs[startNewStructs]
	firstStruct.Write(encoder, int(clock-firstStruct.GetID().Clock))

	for i := startNewStructs + 1; i < len(structs); i++ {
		structs[i].Write(encoder, 0)
	}

	return nil
}

// WriteClientsStructs writes client structs to encoder
func WriteClientsStructs(encoder contracts.IUpdateEncoder, store contracts.IStructStore, sm map[int64]int64) error {
	// We filter all valid sm entries into filteredSm
	filteredSm := make(map[int64]int64)

	for client, clock := range sm {
		// Only write if new structs are available
		if store.GetState(client) > clock {
			filteredSm[client] = clock
		}
	}

	stateVector := store.GetStateVector()
	for client := range stateVector {
		if _, exists := sm[client]; !exists {
			filteredSm[client] = 0
		}
	}

	// Write # states that were updated
	encoder.GetRestWriter().WriteVarUint(uint64(len(filteredSm)))

	// Write items with higher client ids first
	// This heavily improves the conflict resolution algorithm
	var sortedClients []int64
	for client := range filteredSm {
		sortedClients = append(sortedClients, client)
	}
	sort.Slice(sortedClients, func(i, j int) bool {
		return sortedClients[i] > sortedClients[j]
	})

	for _, client := range sortedClients {
		err := WriteStructs(encoder, store.GetClients()[client], client, filteredSm[client])
		if err != nil {
			return err
		}
	}

	return nil
}

// ReadClientStructRefs reads client struct references from decoder
func ReadClientStructRefs(decoder contracts.IUpdateDecoder, doc contracts.IYDoc) (map[int64][]contracts.IStructItem, error) {
	clientRefs := make(map[int64][]contracts.IStructItem)
	numOfStateUpdates := decoder.GetReader().ReadVarUint()

	for i := uint64(0); i < numOfStateUpdates; i++ {
		numberOfStructs := int(decoder.GetReader().ReadVarUint())
		if numberOfStructs < 0 {
			return nil, fmt.Errorf("invalid numberOfStructs: %d", numberOfStructs)
		}

		refs := make([]contracts.IStructItem, 0, numberOfStructs)
		client := decoder.ReadClient()
		clock := int64(decoder.GetReader().ReadVarUint())

		clientRefs[client] = refs

		for j := 0; j < numberOfStructs; j++ {
			info := decoder.ReadInfo()
			if (lib0.Bits5 & info) != 0 {
				// The item that was originally to the left of this item
				var leftOrigin *contracts.StructID
				if (info & lib0.Bit8) == lib0.Bit8 {
					id := decoder.ReadLeftID()
					leftOrigin = &id
				}

				// The item that was originally to the right of this item
				var rightOrigin *contracts.StructID
				if (info & lib0.Bit7) == lib0.Bit7 {
					id := decoder.ReadRightID()
					rightOrigin = &id
				}

				cantCopyParentInfo := (info & (lib0.Bit7 | lib0.Bit8)) == 0
				var hasParentYKey bool
				if cantCopyParentInfo {
					hasParentYKey = decoder.ReadParentInfo()
				}

				// If parent == null and neither left nor right are defined, then we know that 'parent' is child of 'y'
				// and we read the next string as parentYKey.
				// It indicates how we store/retrieve parent from 'y.share'.
				var parentYKey *string
				if cantCopyParentInfo && hasParentYKey {
					key := decoder.ReadString()
					parentYKey = &key
				}

				var parent interface{}
				if cantCopyParentInfo && !hasParentYKey {
					id := decoder.ReadLeftID()
					parent = id
				} else if parentYKey != nil {
					parent = doc.Get(*parentYKey)
				}

				var parentSub *string
				if cantCopyParentInfo && (info&lib0.Bit6) == lib0.Bit6 {
					sub := decoder.ReadString()
					parentSub = &sub
				}

				content, err := ReadItemContent(decoder, info)
				if err != nil {
					return nil, err
				}

				str := NewStructItem(
					contracts.StructID{Client: client, Clock: clock},
					nil, // left
					leftOrigin,
					nil, // right
					rightOrigin,
					parent,
					getStringValue(parentSub),
					content,
				)

				refs = append(refs, str)
				clock += int64(str.GetLength())
			} else {
				length := decoder.ReadLength()
				refs = append(refs, NewStructGC(contracts.StructID{Client: client, Clock: clock}, int(length)))
				clock += int64(length)
			}
		}

		clientRefs[client] = refs
	}

	return clientRefs, nil
}

// WriteStateVector writes state vector to encoder
func WriteStateVector(encoder contracts.IDSEncoder, sv map[int64]int64) {
	encoder.GetRestWriter().WriteVarUint(uint64(len(sv)))

	for client, clock := range sv {
		encoder.GetRestWriter().WriteVarUint(uint64(client))
		encoder.GetRestWriter().WriteVarUint(uint64(clock))
	}
}

// ReadStateVector reads state vector from decoder
func ReadStateVector(decoder contracts.IDSDecoder) map[int64]int64 {
	ssLength := int(decoder.GetReader().ReadVarUint())
	ss := make(map[int64]int64, ssLength)

	for i := 0; i < ssLength; i++ {
		client := int64(decoder.GetReader().ReadVarUint())
		clock := int64(decoder.GetReader().ReadVarUint())
		ss[client] = clock
	}

	return ss
}

// FindIndexSS performs a binary search on a sorted array
func FindIndexSS(structs []contracts.IStructItem, clock int64) int {
	if len(structs) == 0 {
		return 0
	}

	left, right := 0, len(structs)-1
	for left <= right {
		mid := left + (right-left)/2
		if structs[mid].GetID().Clock <= clock {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return left
}

// Helper function to get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
