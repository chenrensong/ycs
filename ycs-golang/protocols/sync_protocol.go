// ------------------------------------------------------------------------------
//  Copyright (c) Microsoft Corporation.  All rights reserved.
// ------------------------------------------------------------------------------

package protocols

import (
	"fmt"

	"github.com/chenrensong/ygo/lib0"
	"github.com/chenrensong/ygo/types"
)

const (
	MessageYjsSyncStep1 uint32 = 0
	MessageYjsSyncStep2 uint32 = 1
	MessageYjsUpdate    uint32 = 2
)

// WriteSyncStep1 writes sync step 1 message to stream
func WriteSyncStep1(stream lib0.StreamWriter, doc *types.YDoc) error {
	if err := lib0.WriteVarUint(stream, MessageYjsSyncStep1); err != nil {
		return err
	}
	sv, err := doc.EncodeStateVectorV2()
	if err != nil {
		return err
	}
	return lib0.WriteVarUint8Array(stream, sv)
}

// WriteSyncStep2 writes sync step 2 message to stream
func WriteSyncStep2(stream lib0.StreamWriter, doc *types.YDoc, encodedStateVector []byte) error {
	if err := lib0.WriteVarUint(stream, MessageYjsSyncStep2); err != nil {
		return err
	}
	update, err := doc.EncodeStateAsUpdateV2(encodedStateVector)
	if err != nil {
		return err
	}
	return lib0.WriteVarUint8Array(stream, update)
}

// ReadSyncStep1 reads sync step 1 message and writes response
func ReadSyncStep1(reader lib0.StreamReader, writer lib0.StreamWriter, doc *types.YDoc) error {
	encodedStateVector, err := lib0.ReadVarUint8Array(reader)
	if err != nil {
		return err
	}
	return WriteSyncStep2(writer, doc, encodedStateVector)
}

// ReadSyncStep2 reads sync step 2 message and applies update
func ReadSyncStep2(stream lib0.StreamReader, doc *types.YDoc, transactionOrigin interface{}) error {
	update, err := lib0.ReadVarUint8Array(stream)
	if err != nil {
		return err
	}
	return doc.ApplyUpdateV2Bytes(update, transactionOrigin, false)
}

// WriteUpdate writes update message to stream
func WriteUpdate(stream lib0.StreamWriter, update []byte) error {
	if err := lib0.WriteVarUint(stream, MessageYjsUpdate); err != nil {
		return err
	}
	return lib0.WriteVarUint8Array(stream, update)
}

// ReadUpdate reads update message and applies it
func ReadUpdate(stream lib0.StreamReader, doc *types.YDoc, transactionOrigin interface{}) error {
	return ReadSyncStep2(stream, doc, transactionOrigin)
}

// ReadSyncMessage reads and processes sync message, returns message type
func ReadSyncMessage(reader lib0.StreamReader, writer lib0.StreamWriter, doc *types.YDoc, transactionOrigin interface{}) (uint32, error) {
	messageType, err := lib0.ReadVarUint(reader)
	if err != nil {
		return 0, err
	}

	switch messageType {
	case MessageYjsSyncStep1:
		err = ReadSyncStep1(reader, writer, doc)
	case MessageYjsSyncStep2:
		err = ReadSyncStep2(reader, doc, transactionOrigin)
	case MessageYjsUpdate:
		err = ReadUpdate(reader, doc, transactionOrigin)
	default:
		return 0, fmt.Errorf("unknown message type: %d", messageType)
	}

	if err != nil {
		return 0, err
	}
	return messageType, nil
}
