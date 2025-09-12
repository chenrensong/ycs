// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"bytes"
	"ycs-golang/core"
	"ycs-golang/structs"
	"ycs-golang/types"
)

// EncodingUtils provides utility functions for encoding and decoding
type EncodingUtils struct{}

// ReadItemContent reads item content from a decoder
func ReadItemContent(decoder IUpdateDecoder, info byte) structs.IContent {
	switch info & core.Bits5 {
	case 0: // GC
		panic("GC is not ItemContent")
	case 1: // Deleted
		// Note: ContentDeleted.Read needs to be implemented
		// return ContentDeleted.Read(decoder)
		return nil
	case 2: // JSON
		// Note: ContentJson.Read needs to be implemented
		// return ContentJson.Read(decoder)
		return nil
	case 3: // Binary
		// Note: ContentBinary.Read needs to be implemented
		// return ContentBinary.Read(decoder)
		return nil
	case 4: // String
		// Note: ContentString.Read needs to be implemented
		// return ContentString.Read(decoder)
		return nil
	case 5: // Embed
		// Note: ContentEmbed.Read needs to be implemented
		// return ContentEmbed.Read(decoder)
		return nil
	case 6: // Format
		// Note: ContentFormat.Read needs to be implemented
		// return ContentFormat.Read(decoder)
		return nil
	case 7: // Type
		// Note: ContentType.Read needs to be implemented
		// return ContentType.Read(decoder)
		return nil
	case 8: // Any
		// Note: ContentAny.Read needs to be implemented
		// return ContentAny.Read(decoder)
		return nil
	case 9: // Doc
		// Note: ContentDoc.Read needs to be implemented
		// return ContentDoc.Read(decoder)
		return nil
	default:
		panic("Content type not recognized")
	}
}

// ReadStructs reads structs from a decoder
func ReadStructs(decoder IUpdateDecoder, transaction *Transaction, store *StructStore) {
	clientStructRefs := ReadClientStructRefs(decoder, transaction.Doc)
	store.MergeReadStructsIntoPendingReads(clientStructRefs)
	store.ResumeStructIntegration(transaction)
	store.CleanupPendingStructs()
	store.TryResumePendingDeleteReaders(transaction)
}

// WriteStructs writes structs to an encoder
func WriteStructs(encoder IUpdateEncoder, structs []*structs.AbstractStruct, client, clock int64) {
	// Write first id.
	startNewStructs := StructStoreFindIndexSS(structs, clock)

	// Write # encoded structs.
	core.WriteVarUint(encoder.RestWriter, uint64(len(structs)-startNewStructs))
	encoder.WriteClient(client)
	core.WriteVarUint(encoder.RestWriter, uint64(clock))

	// Write first struct with offset.
	firstStruct := structs[startNewStructs]
	firstStruct.Write(encoder, int(clock-firstStruct.Id.Clock))

	for i := startNewStructs + 1; i < len(structs); i++ {
		structs[i].Write(encoder, 0)
	}
}

// WriteClientsStructs writes client structs to an encoder
func WriteClientsStructs(encoder IUpdateEncoder, store *StructStore, _sm map[int64]int64) {
	// We filter all valid _sm entries into sm.
	sm := make(map[int64]int64)
	for client, clock := range _sm {
		// Only write if new structs are available.
		if store.GetState(client) > clock {
			sm[client] = clock
		}
	}

	for client, clock := range store.GetStateVector() {
		if _, exists := _sm[client]; !exists {
			sm[client] = 0
		}
	}

	// Write # states that were updated.
	core.WriteVarUint(encoder.RestWriter, uint64(len(sm)))

	// Write items with higher client ids first.
	// This heavily improves the conflict resolution algorithm.
	// Note: In Go, we need to implement our own sort function
	// This is a simplified version - you may need to implement a proper sort
	sortedClients := make([]int64, 0, len(sm))
	for client := range sm {
		sortedClients = append(sortedClients, client)
	}

	for _, client := range sortedClients {
		WriteStructs(encoder, store.Clients[client], client, sm[client])
	}
}

// ReadClientStructRefs reads client struct references from a decoder
func ReadClientStructRefs(decoder IUpdateDecoder, doc *YDoc) map[int64][]*structs.AbstractStruct {
	clientRefs := make(map[int64][]*structs.AbstractStruct)
	numOfStateUpdates := core.ReadVarUint(decoder.Reader)

	for i := uint64(0); i < numOfStateUpdates; i++ {
		numberOfStructs := int(core.ReadVarUint(decoder.Reader))
		refs := make([]*structs.AbstractStruct, 0, numberOfStructs)
		client := decoder.ReadClient()
		clock := int64(core.ReadVarUint(decoder.Reader))

		clientRefs[client] = refs

		for j := 0; j < numberOfStructs; j++ {
			info := decoder.ReadInfo()
			if (core.Bits5 & info) != 0 {
				// The item that was originally to the left of this item.
				var leftOrigin *ID
				if (info & core.Bit8) == core.Bit8 {
					id := decoder.ReadLeftId()
					leftOrigin = &id
				}

				// The item that was originally to the right of this item.
				var rightOrigin *ID
				if (info & core.Bit7) == core.Bit7 {
					id := decoder.ReadRightId()
					rightOrigin = &id
				}

				cantCopyParentInfo := (info & (core.Bit7 | core.Bit8)) == 0
				hasParentYKey := false
				if cantCopyParentInfo {
					hasParentYKey = decoder.ReadParentInfo()
				}

				// If parent == null and neither left nor right are defined, then we know that 'parent' is child of 'y'
				// and we read the next string as parentYKey.
				// It indicates how we store/retrieve parent from 'y.share'.
				var parentYKey string
				if cantCopyParentInfo && hasParentYKey {
					parentYKey = decoder.ReadString()
				}

				// Note: This requires implementing the Get method for YDoc
				// var parent interface{}
				// if cantCopyParentInfo && !hasParentYKey {
				//     parent = decoder.ReadLeftId()
				// } else if parentYKey != "" {
				//     parent = doc.Get(parentYKey)
				// }

				var parentSub string
				if cantCopyParentInfo && (info&core.Bit6) == core.Bit6 {
					parentSub = decoder.ReadString()
				}

				content := ReadItemContent(decoder, info)

				str := structs.NewItem(
					ID{Client: client, Clock: clock},
					nil, // left
					leftOrigin,
					nil, // right
					rightOrigin, // rightOrigin
					nil, // parent - need to implement
					parentSub,
					content)

				refs = append(refs, str)
				clock += str.Length
			} else {
				length := decoder.ReadLength()
				gc := structs.NewGC(ID{Client: client, Clock: clock}, int64(length))
				refs = append(refs, gc)
				clock += int64(length)
			}
		}
		
		clientRefs[client] = refs
	}

	return clientRefs
}

// WriteStateVector writes a state vector to an encoder
func WriteStateVector(encoder IDSDecoder, sv map[int64]int64) {
	core.WriteVarUint(encoder.RestWriter, uint64(len(sv)))

	for client, clock := range sv {
		core.WriteVarUint(encoder.RestWriter, uint64(client))
		core.WriteVarUint(encoder.RestWriter, uint64(clock))
	}
}

// ReadStateVector reads a state vector from a decoder
func ReadStateVector(decoder IDSDecoder) map[int64]int64 {
	ssLength := int(core.ReadVarUint(decoder.Reader))
	ss := make(map[int64]int64, ssLength)

	for i := 0; i < ssLength; i++ {
		client := int64(core.ReadVarUint(decoder.Reader))
		clock := int64(core.ReadVarUint(decoder.Reader))
		ss[client] = clock
	}

	return ss
}

// DecodeStateVector decodes a state vector from input
func DecodeStateVector(input *bytes.Reader) map[int64]int64 {
	// Note: DSDecoderV2 needs to be implemented
	// decoder := NewDSDecoderV2(input)
	// return ReadStateVector(decoder)
	return make(map[int64]int64) // Placeholder
}