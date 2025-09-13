package utils

import (
	"errors"
	"io"
	"sort"

	"github.com/chenrensong/ygo/structs"
	"github.com/chenrensong/ygo/types"
	"github.com/chenrensong/ygo/utils"
)

const (
	// Content type identifiers (first 5 bits)
	GCType      = 0
	DeletedType = 1
	JSONType    = 2
	BinaryType  = 3
	StringType  = 4
	EmbedType   = 5
	FormatType  = 6
	TypeType    = 7
	AnyType     = 8
	DocType     = 9
)

// ReadItemContent reads item content based on the info flag
func ReadItemContent(decoder IUpdateDecoder, info byte) (types.Content, error) {
	switch info & 0b00011111 { // Mask first 5 bits
	case GCType:
		return nil, errors.New("GC is not ItemContent")
	case DeletedType:
		return types.ReadContentDeleted(decoder)
	case JSONType:
		return types.ReadContentJson(decoder)
	case BinaryType:
		return types.ReadContentBinary(decoder)
	case StringType:
		return types.ReadContentString(decoder)
	case EmbedType:
		return types.ReadContentEmbed(decoder)
	case FormatType:
		return types.ReadContentFormat(decoder)
	case TypeType:
		return types.ReadContentType(decoder)
	case AnyType:
		return types.ReadContentAny(decoder)
	case DocType:
		return types.ReadContentDoc(decoder)
	default:
		return nil, errors.New("content type not recognized")
	}
}

// ReadStructs reads and integrates structs from a decoder
func ReadStructs(decoder IUpdateDecoder, transaction *Transaction, store *StructStore) error {
	clientStructRefs, err := ReadClientStructRefs(decoder, transaction.Doc())
	if err != nil {
		return err
	}

	store.MergeReadStructsIntoPendingReads(clientStructRefs)
	store.ResumeStructIntegration(transaction)
	store.CleanupPendingStructs()
	store.TryResumePendingDeleteReaders(transaction)
	return nil
}

// WriteStructs writes structs to an encoder
func WriteStructs(encoder Encoder, structs []structs.AbstractStruct, client uint64, clock uint64) error {
	startNewStructs := structs.FindIndexSS(clock)

	// Write number of encoded structs
	if err := encoder.RestWriter().WriteVarUint(uint64(len(structs) - startNewStructs)); err != nil {
		return err
	}

	if err := encoder.WriteClient(client); err != nil {
		return err
	}

	if err := encoder.RestWriter().WriteVarUint(clock); err != nil {
		return err
	}

	// Write first struct with offset
	firstStruct := structs[startNewStructs]
	if err := firstStruct.Write(encoder, int(clock-firstStruct.ID().Clock)); err != nil {
		return err
	}

	// Write remaining structs
	for i := startNewStructs + 1; i < len(structs); i++ {
		if err := structs[i].Write(encoder, 0); err != nil {
			return err
		}
	}

	return nil
}

// WriteClientsStructs writes clients' structs to an encoder
func WriteClientsStructs(encoder Encoder, store *structs.StructStore, sm map[uint64]uint64) error {
	// Filter valid sm entries
	filteredSM := make(map[uint64]uint64)
	for client, clock := range sm {
		if store.GetState(client) > clock {
			filteredSM[client] = clock
		}
	}

	// Add missing clients from state vector
	for client := range store.GetStateVector() {
		if _, exists := sm[client]; !exists {
			filteredSM[client] = 0
		}
	}

	// Write number of updated states
	if err := encoder.RestWriter().WriteVarUint(uint64(len(filteredSM))); err != nil {
		return err
	}

	// Sort clients in descending order for better conflict resolution
	sortedClients := make([]uint64, 0, len(filteredSM))
	for client := range filteredSM {
		sortedClients = append(sortedClients, client)
	}
	sort.Slice(sortedClients, func(i, j int) bool {
		return sortedClients[i] > sortedClients[j]
	})

	// Write structs for each client
	for _, client := range sortedClients {
		if err := WriteStructs(encoder, store.Clients()[client], client, filteredSM[client]); err != nil {
			return err
		}
	}

	return nil
}

// ReadClientStructRefs reads client struct references from a decoder
func ReadClientStructRefs(decoder Decoder, doc *structs.YDoc) (map[uint64][]structs.AbstractStruct, error) {
	clientRefs := make(map[uint64][]structs.AbstractStruct)
	numOfStateUpdates, err := decoder.Reader().ReadVarUint()
	if err != nil {
		return nil, err
	}

	for i := uint64(0); i < numOfStateUpdates; i++ {
		numberOfStructs, err := decoder.Reader().ReadVarUint()
		if err != nil {
			return nil, err
		}

		refs := make([]structs.AbstractStruct, 0, numberOfStructs)
		client, err := decoder.ReadClient()
		if err != nil {
			return nil, err
		}

		clock, err := decoder.Reader().ReadVarUint()
		if err != nil {
			return nil, err
		}

		clientRefs[client] = refs

		for j := uint64(0); j < numberOfStructs; j++ {
			info, err := decoder.ReadInfo()
			if err != nil {
				return nil, err
			}

			if info&0b00011111 != 0 { // Item
				// Read left origin if present
				var leftOrigin *utils.ID
				if info&0b10000000 != 0 {
					id, err := decoder.ReadLeftID()
					if err != nil {
						return nil, err
					}
					leftOrigin = &id
				}

				// Read right origin if present
				var rightOrigin *utils.ID
				if info&0b01000000 != 0 {
					id, err := decoder.ReadRightID()
					if err != nil {
						return nil, err
					}
					rightOrigin = &id
				}

				cantCopyParentInfo := info&0b11000000 == 0
				var hasParentYKey bool
				if cantCopyParentInfo {
					hasParentYKey, err = decoder.ReadParentInfo()
					if err != nil {
						return nil, err
					}
				}

				// Read parent key if needed
				var parentYKey string
				if cantCopyParentInfo && hasParentYKey {
					parentYKey, err = decoder.ReadString()
					if err != nil {
						return nil, err
					}
				}

				// Determine parent
				var parent interface{}
				if cantCopyParentInfo && !hasParentYKey {
					leftID, err := decoder.ReadLeftID()
					if err != nil {
						return nil, err
					}
					parent = leftID
				} else if parentYKey != "" {
					parent = doc.Get(parentYKey)
				}

				// Read parent sub if present
				var parentSub string
				if cantCopyParentInfo && info&0b00100000 != 0 {
					parentSub, err = decoder.ReadString()
					if err != nil {
						return nil, err
					}
				}

				// Read content
				content, err := ReadItemContent(decoder, info)
				if err != nil {
					return nil, err
				}

				// Create item
				item := structs.NewItem(
					utils.NewID(client, clock),
					nil, // left
					leftOrigin,
					nil, // right
					rightOrigin,
					parent,
					parentSub,
					content,
				)

				refs = append(refs, item)
				clock += uint64(item.Length())
			} else { // GC
				length, err := decoder.ReadLength()
				if err != nil {
					return nil, err
				}

				gc := structs.NewGC(utils.NewID(client, clock), int(length))
				refs = append(refs, gc)
				clock += uint64(length)
			}
		}
	}

	return clientRefs, nil
}

// WriteStateVector writes a state vector to an encoder
func WriteStateVector(encoder Encoder, sv map[uint64]uint64) error {
	if err := encoder.RestWriter().WriteVarUint(uint64(len(sv))); err != nil {
		return err
	}

	for client, clock := range sv {
		if err := encoder.RestWriter().WriteVarUint(client); err != nil {
			return err
		}
		if err := encoder.RestWriter().WriteVarUint(clock); err != nil {
			return err
		}
	}

	return nil
}

// ReadStateVector reads a state vector from a decoder
func ReadStateVector(decoder Decoder) (map[uint64]uint64, error) {
	ssLength, err := decoder.Reader().ReadVarUint()
	if err != nil {
		return nil, err
	}

	ss := make(map[uint64]uint64, ssLength)
	for i := uint64(0); i < ssLength; i++ {
		client, err := decoder.Reader().ReadVarUint()
		if err != nil {
			return nil, err
		}

		clock, err := decoder.Reader().ReadVarUint()
		if err != nil {
			return nil, err
		}

		ss[client] = clock
	}

	return ss, nil
}

// DecodeStateVector decodes a state vector from a byte stream
func DecodeStateVector(input io.Reader) (map[uint64]uint64, error) {
	return ReadStateVector(NewDSDecoderV2(input))
}
