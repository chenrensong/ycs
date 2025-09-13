package protocols

import (
	"bytes"
	"errors"
	"io"

	lib0 "github.com/chenrensong/ygo/lib0"
	utils "github.com/chenrensong/ygo/utils"
)

const (
	MessageYjsSyncStep1 = 0
	MessageYjsSyncStep2 = 1
	MessageYjsUpdate    = 2
)

// WriteSyncStep1 writes the initial sync message with the document's state vector.
func WriteSyncStep1(writer io.Writer, doc *utils.YDoc) error {
	buf := &bytes.Buffer{}
	if err := lib0.WriteVarUint(buf, MessageYjsSyncStep1); err != nil {
		return err
	}
	sv := doc.EncodeStateVectorV2()
	if err := lib0.WriteVarUint8Array(buf, sv); err != nil {
		return err
	}
	_, err := writer.Write(buf.Bytes())
	return err
}

// WriteSyncStep2 writes the response to a sync step 1 message with the document updates.
func WriteSyncStep2(writer io.Writer, doc *utils.YDoc, encodedStateVector []byte) error {
	buf := &bytes.Buffer{}
	if err := lib0.WriteVarUint(buf, MessageYjsSyncStep2); err != nil {
		return err
	}
	update := doc.EncodeStateAsUpdateV2(encodedStateVector)
	if err := lib0.WriteVarUint8Array(buf, update); err != nil {
		return err
	}
	_, err := writer.Write(buf.Bytes())
	return err
}

// ReadSyncStep1 reads a sync step 1 message and responds with step 2.
func ReadSyncStep1(reader io.Reader, writer io.Writer, doc *utils.YDoc) error {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, reader); err != nil {
		return err
	}
	encodedStateVector, err := lib0.ReadVarUint8Array(buf)
	if err != nil {
		return err
	}
	return WriteSyncStep2(writer, doc, encodedStateVector)
}

// ReadSyncStep2 reads and applies updates from a sync step 2 message.
func ReadSyncStep2(reader io.Reader, doc *utils.YDoc, transactionOrigin interface{}) error {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, reader); err != nil {
		return err
	}
	update, err := lib0.ReadVarUint8Array(buf)
	if err != nil {
		return err
	}
	updateReader := bytes.NewReader(update)
	doc.ApplyUpdateV2(updateReader, transactionOrigin, false)
	return nil
}

// WriteUpdate writes an update message.
func WriteUpdate(writer io.Writer, update []byte) error {
	buf := &bytes.Buffer{}
	if err := lib0.WriteVarUint(buf, MessageYjsUpdate); err != nil {
		return err
	}
	if err := lib0.WriteVarUint8Array(buf, update); err != nil {
		return err
	}
	_, err := writer.Write(buf.Bytes())
	return err
}

// ReadUpdate reads and applies an update message.
func ReadUpdate(reader io.Reader, doc *utils.YDoc, transactionOrigin interface{}) error {
	return ReadSyncStep2(reader, doc, transactionOrigin)
}

// ReadSyncMessage reads and processes a sync message.
func ReadSyncMessage(reader io.Reader, writer io.Writer, doc *utils.YDoc, transactionOrigin interface{}) (uint32, error) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, reader); err != nil {
		return 0, err
	}
	messageType, err := lib0.ReadVarUint(buf)
	if err != nil {
		return 0, err
	}

	switch messageType {
	case MessageYjsSyncStep1:
		err = ReadSyncStep1(buf, writer, doc)
	case MessageYjsSyncStep2:
		err = ReadSyncStep2(buf, doc, transactionOrigin)
	case MessageYjsUpdate:
		err = ReadUpdate(buf, doc, transactionOrigin)
	default:
		err = errors.New("unknown message type")
	}

	return messageType, err
}
