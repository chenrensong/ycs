package protocols

import (
	"fmt"
	"io"
	"ycs/core"
	"ycs/lib0"
)

// Message type constants for Y.js sync protocol
const (
	MessageYjsSyncStep1 = 0
	MessageYjsSyncStep2 = 1
	MessageYjsUpdate    = 2
)

// WriteSyncStep1 writes sync step 1 message to stream
func WriteSyncStep1(writer io.Writer, doc *core.YDoc) error {
	streamWriter := writer.(lib0.StreamWriter)
	if err := lib0.WriteVarUint(streamWriter, MessageYjsSyncStep1); err != nil {
		return err
	}

	sv := doc.EncodeStateVectorV2()
	return lib0.WriteVarUint8Array(streamWriter, sv)
}

// WriteSyncStep2 writes sync step 2 message to stream
func WriteSyncStep2(writer io.Writer, doc *core.YDoc, encodedStateVector []byte) error {
	streamWriter := writer.(lib0.StreamWriter)
	if err := lib0.WriteVarUint(streamWriter, MessageYjsSyncStep2); err != nil {
		return err
	}

	update := doc.EncodeStateAsUpdateV2(encodedStateVector)
	return lib0.WriteVarUint8Array(streamWriter, update)
}

// ReadSyncStep1 reads sync step 1 message and responds with step 2
func ReadSyncStep1(reader io.Reader, writer io.Writer, doc *core.YDoc) error {
	streamReader := reader.(lib0.StreamReader)
	encodedStateVector, err := lib0.ReadVarUint8Array(streamReader)
	if err != nil {
		return err
	}

	return WriteSyncStep2(writer, doc, encodedStateVector)
}

// ReadSyncStep2 reads sync step 2 message and applies update to document
func ReadSyncStep2(reader io.Reader, doc *core.YDoc, transactionOrigin interface{}) error {
	streamReader := reader.(lib0.StreamReader)
	update, err := lib0.ReadVarUint8Array(streamReader)
	if err != nil {
		return err
	}

	doc.ApplyUpdateV2(update, transactionOrigin, false)
	return nil
}

// WriteUpdate writes an update message to stream
func WriteUpdate(writer io.Writer, update []byte) error {
	streamWriter := writer.(lib0.StreamWriter)
	if err := lib0.WriteVarUint(streamWriter, MessageYjsUpdate); err != nil {
		return err
	}

	return lib0.WriteVarUint8Array(streamWriter, update)
}

// ReadUpdate reads an update message from stream
func ReadUpdate(reader io.Reader, doc *core.YDoc, transactionOrigin interface{}) error {
	return ReadSyncStep2(reader, doc, transactionOrigin)
}

// ReadSyncMessage reads and processes a sync message, returning the message type
func ReadSyncMessage(reader io.Reader, writer io.Writer, doc *core.YDoc, transactionOrigin interface{}) (uint32, error) {
	streamReader := reader.(lib0.StreamReader)
	messageType, err := lib0.ReadVarUint(streamReader)
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
		return messageType, fmt.Errorf("unknown message type: %d", messageType)
	}

	if err != nil {
		return messageType, err
	}

	return messageType, nil
}
