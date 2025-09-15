package core

import (
	"errors"
	"io"
	"sort"
	"ycs/contracts"
	"ycs/lib0"
)

// EncodingUtils provides encoding and decoding utilities
type EncodingUtils struct {
	contentReaderRegistry contracts.IContentReaderRegistry
}

var encodingUtils *EncodingUtils

// SetContentReaderRegistry sets the content reader registry
func SetContentReaderRegistry(registry contracts.IContentReaderRegistry) {
	if encodingUtils == nil {
		encodingUtils = &EncodingUtils{}
	}
	encodingUtils.contentReaderRegistry = registry
}

// ReadItemContent reads item content from a decoder
func ReadItemContent(decoder contracts.IUpdateDecoder, info byte) (contracts.IContent, error) {
	contentRef := int(info & 0x1F) // Bits5
	if contentRef == 0 {           // GC
		return nil, errors.New("GC is not ItemContent")
	}

	if encodingUtils == nil || encodingUtils.contentReaderRegistry == nil {
		return nil, errors.New("ContentReaderRegistry not initialized. Call Initialize() first")
	}

	return encodingUtils.contentReaderRegistry.ReadContent(contentRef, decoder), nil
}

// ReadStructs reads the next Item in a Decoder and fills structs with the read data.
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
func WriteStructs(encoder contracts.IUpdateEncoder, structs []contracts.IStructItem, client, clock int64) error {
	// Write first id
	startNewStructs := FindIndexSS(structs, clock)

	// Write # encoded structs
	lib0.WriteVarUint(encoder.GetRestWriter(), uint32(len(structs)-startNewStructs))
	encoder.WriteClient(client)
	lib0.WriteVarUint(encoder.GetRestWriter(), uint32(clock))

	// Write first struct with offset
	firstStruct := structs[startNewStructs]
	err := firstStruct.Write(encoder, int(clock-firstStruct.GetID().Clock))
	if err != nil {
		return err
	}

	for i := startNewStructs + 1; i < len(structs); i++ {
		err := structs[i].Write(encoder, 0)
		if err != nil {
			return err
		}
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
	lib0.WriteVarUint(encoder.GetRestWriter(), uint32(len(filteredSm)))

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
	numOfStateUpdates, err := lib0.ReadVarUint(decoder.GetReader().(lib0.StreamReader))
	if err != nil {
		return nil, err
	}

	clientRefs := make(map[int64][]contracts.IStructItem)
	for i := uint32(0); i < numOfStateUpdates; i++ {
		numberOfStructsVal, err := lib0.ReadVarUint(decoder.GetReader().(lib0.StreamReader))
		if err != nil {
			return nil, err
		}
		numberOfStructs := int(numberOfStructsVal)
		if numberOfStructs < 0 {
			return nil, errors.New("invalid number of structs")
		}

		refs := make([]contracts.IStructItem, 0, numberOfStructs)
		client := decoder.ReadClient()
		clockVal, err := lib0.ReadVarUint(decoder.GetReader().(lib0.StreamReader))
		if err != nil {
			return nil, err
		}
		clock := int64(clockVal)

		clientRefs[client] = refs

		for j := 0; j < numberOfStructs; j++ {
			info := decoder.ReadInfo()
			if (info & 0x1F) != 0 { // Bits5
				// The item that was originally to the left of this item
				var leftOrigin *contracts.StructID
				if (info & 0x80) == 0x80 { // Bit8
					id := decoder.ReadLeftID()
					leftOrigin = &id
				}

				// The item that was originally to the right of this item
				var rightOrigin *contracts.StructID
				if (info & 0x40) == 0x40 { // Bit7
					id := decoder.ReadRightID()
					rightOrigin = &id
				}

				cantCopyParentInfo := (info & (0x40 | 0x80)) == 0
				var hasParentYKey bool
				if cantCopyParentInfo {
					hasParentYKey = decoder.ReadParentInfo()
				}

				// If parent == nil and neither left nor right are defined, then we know that 'parent' is child of 'y'
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
					parent = doc.Get(*parentYKey, nil) // Pass nil as typeConstructor
				}

				var parentSub *string
				if cantCopyParentInfo && (info&0x20) == 0x20 { // Bit6
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
					parentSub,
					content.(contracts.IContentEx), // Convert IContent to IContentEx
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
func WriteStateVector(encoder contracts.IDSEncoder, sv map[int64]int64) error {
	lib0.WriteVarUint(encoder.GetRestWriter(), uint32(len(sv)))

	for client, clock := range sv {
		lib0.WriteVarUint(encoder.GetRestWriter(), uint32(client))
		lib0.WriteVarUint(encoder.GetRestWriter(), uint32(clock))
	}

	return nil
}

// ReadStateVector reads state vector from decoder
func ReadStateVector(decoder contracts.IDSDecoder) (map[int64]int64, error) {
	ssLengthVal, err := lib0.ReadVarUint(decoder.GetReader().(lib0.StreamReader))
	if err != nil {
		return nil, err
	}
	ssLength := int(ssLengthVal)
	ss := make(map[int64]int64, ssLength)

	for i := 0; i < ssLength; i++ {
		clientVal, err := lib0.ReadVarUint(decoder.GetReader().(lib0.StreamReader))
		if err != nil {
			return nil, err
		}
		client := int64(clientVal)

		clockVal, err := lib0.ReadVarUint(decoder.GetReader().(lib0.StreamReader))
		if err != nil {
			return nil, err
		}
		clock := int64(clockVal)

		ss[client] = clock
	}

	return ss, nil
}

// DecodeStateVector decodes state vector from input stream
func DecodeStateVector(input io.Reader) (map[int64]int64, error) {
	decoder := NewDSDecoderV2(input)
	return ReadStateVector(decoder)
}

// FindIndexSS performs binary search on a sorted array
func FindIndexSS(structs []contracts.IStructItem, clock int64) int {
	if len(structs) == 0 {
		panic("structs array cannot be empty")
	}

	left := 0
	right := len(structs) - 1
	mid := structs[right]
	midClock := mid.GetID().Clock

	if midClock == clock {
		return right
	}

	// Binary search with pivoting
	midIndex := int(clock * int64(right) / (midClock + int64(mid.GetLength()) - 1))
	for left <= right {
		mid = structs[midIndex]
		midClock = mid.GetID().Clock

		if midClock <= clock {
			if clock < midClock+int64(mid.GetLength()) {
				return midIndex
			}
			left = midIndex + 1
		} else {
			right = midIndex - 1
		}

		midIndex = (left + right) / 2
	}

	// Always check state before looking for a struct in StructStore
	// Therefore the case of not finding a struct is unexpected
	panic("struct not found - unexpected")
}
