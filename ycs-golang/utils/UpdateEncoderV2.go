package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/chenrensong/ygo/lib0"
	"github.com/chenrensong/ygo/lib0/encoding"
)

var (
	ErrInvalidLength = errors.New("length must be positive")
)

// DSEncoderV2 is a document state encoder for version 2
type DSEncoderV2 struct {
	dsCurVal   int64
	restWriter *bytes.Buffer
	disposed   bool
}

// NewDSEncoderV2 creates a new DSEncoderV2 instance
func NewDSEncoderV2() *DSEncoderV2 {
	return &DSEncoderV2{
		restWriter: bytes.NewBuffer(nil),
	}
}

// ResetDsCurVal resets the current document state value
func (e *DSEncoderV2) ResetDsCurVal() {
	e.dsCurVal = 0
}

// WriteDsClock writes a document state clock value
func (e *DSEncoderV2) WriteDsClock(clock int64) error {
	diff := clock - e.dsCurVal
	if diff < 0 {
		return errors.New("clock cannot decrease")
	}
	e.dsCurVal = clock
	return lib0.WriteVarUint(e.restWriter, uint32(diff))
}

// WriteDsLength writes a document state length
func (e *DSEncoderV2) WriteDsLength(length int64) error {
	if length <= 0 {
		return ErrInvalidLength
	}
	if err := lib0.WriteVarUint(e.restWriter, uint32(length-1)); err != nil {
		return err
	}
	e.dsCurVal += length
	return nil
}

// ToArray returns the encoded data as a byte array
func (e *DSEncoderV2) ToArray() []byte {
	return e.restWriter.Bytes()
}

// Dispose releases resources used by the encoder
func (e *DSEncoderV2) Dispose() {
	if !e.disposed {
		e.restWriter = nil
		e.disposed = true
	}
}

// UpdateEncoderV2 extends DSEncoderV2 for update encoding
type UpdateEncoderV2 struct {
	*DSEncoderV2

	keyClock       int
	keyMap         map[string]int
	keyEncoder     *encoding.IntDiffOptRleEncoder
	clientEncoder  *encoding.UintOptRleEncoder
	leftEncoder    *encoding.IntDiffOptRleEncoder
	rightEncoder   *encoding.IntDiffOptRleEncoder
	infoEncoder    *encoding.RleEncoder
	stringEncoder  *encoding.StringEncoder
	parentEncoder  *encoding.RleEncoder
	typeRefEncoder *encoding.UintOptRleEncoder
	lengthEncoder  *encoding.UintOptRleEncoder
}

// NewUpdateEncoderV2 creates a new UpdateEncoderV2 instance
func NewUpdateEncoderV2() *UpdateEncoderV2 {
	return &UpdateEncoderV2{
		DSEncoderV2:    NewDSEncoderV2(),
		keyMap:         make(map[string]int),
		keyEncoder:     encoding.NewIntDiffOptRleEncoder(),
		clientEncoder:  encoding.NewUintOptRleEncoder(),
		leftEncoder:    encoding.NewIntDiffOptRleEncoder(),
		rightEncoder:   encoding.NewIntDiffOptRleEncoder(),
		infoEncoder:    encoding.NewRleEncoder(),
		stringEncoder:  encoding.NewStringEncoder(),
		parentEncoder:  encoding.NewRleEncoder(),
		typeRefEncoder: encoding.NewUintOptRleEncoder(),
		lengthEncoder:  encoding.NewUintOptRleEncoder(),
	}
}

// ToArray returns the encoded data as a byte array
func (e *UpdateEncoderV2) ToArray() []byte {
	buf := bytes.NewBuffer(nil)

	// Write feature flag (currently unused)
	buf.WriteByte(0)

	// Write all encoded data
	writeEncoderData(buf, e.keyEncoder)
	writeEncoderData(buf, e.clientEncoder)
	writeEncoderData(buf, e.leftEncoder)
	writeEncoderData(buf, e.rightEncoder)
	writeEncoderData(buf, e.infoEncoder)
	writeEncoderData(buf, e.stringEncoder)
	writeEncoderData(buf, e.parentEncoder)
	writeEncoderData(buf, e.typeRefEncoder)
	writeEncoderData(buf, e.lengthEncoder)

	// Append rest of the data
	buf.Write(e.DSEncoderV2.ToArray())

	return buf.Bytes()
}

func writeEncoderData(w io.Writer, encoder interface {
	ToArray() ([]byte, error)
}) {
	data, err := encoder.ToArray()
	if err != nil {
		panic(err)
	}

	// 使用 bytes.Buffer 作为适配器，同时实现 io.Writer 和 io.ByteWriter
	buf := &bytes.Buffer{}
	if err := lib0.WriteVarUint8Array(buf, data); err != nil {
		panic(err)
	}

	// 将结果写入原始的 writer
	if _, err := w.Write(buf.Bytes()); err != nil {
		panic(err)
	}
}

// WriteLeftId writes a left ID
func (e *UpdateEncoderV2) WriteLeftId(id *ID) error {
	if err := e.clientEncoder.Write(uint32(id.Client)); err != nil {
		return err
	}
	return e.leftEncoder.Write(int64(id.Clock))
}

// WriteRightId writes a right ID
func (e *UpdateEncoderV2) WriteRightId(id *ID) error {
	if err := e.clientEncoder.Write(uint32(id.Client)); err != nil {
		return err
	}
	return e.rightEncoder.Write(int64(id.Clock))
}

// WriteClient writes a client ID
func (e *UpdateEncoderV2) WriteClient(client uint64) error {
	return e.clientEncoder.Write(uint32(client))
}

// WriteInfo writes info byte
func (e *UpdateEncoderV2) WriteInfo(info byte) error {
	return e.infoEncoder.Write(info)
}

// WriteString writes a string
func (e *UpdateEncoderV2) WriteString(s string) error {
	return e.stringEncoder.Write(s)
}

// WriteParentInfo writes parent info
func (e *UpdateEncoderV2) WriteParentInfo(isYKey bool) error {
	var val byte
	if isYKey {
		val = 1
	}
	return e.parentEncoder.Write(val)
}

// WriteTypeRef writes a type reference
func (e *UpdateEncoderV2) WriteTypeRef(info uint64) error {
	return e.typeRefEncoder.Write(uint32(info))
}

// WriteLength writes a length
func (e *UpdateEncoderV2) WriteLength(length int) error {
	if length < 0 {
		return ErrInvalidLength
	}
	return e.lengthEncoder.Write(uint32(length))
}

// WriteAny writes any value
func (e *UpdateEncoderV2) WriteAny(value interface{}) error {
	return lib0.WriteAny(e.restWriter, value)
}

// WriteBuffer writes a byte buffer
func (e *UpdateEncoderV2) WriteBuffer(data []byte) error {
	return lib0.WriteVarUint8Array(e.restWriter, data)
}

// WriteKey writes a key with optimization for frequently used keys
func (e *UpdateEncoderV2) WriteKey(key string) error {
	if err := e.keyEncoder.Write(int64(e.keyClock)); err != nil {
		return err
	}
	e.keyClock++

	if _, exists := e.keyMap[key]; !exists {
		if err := e.stringEncoder.Write(key); err != nil {
			return err
		}
		e.keyMap[key] = e.keyClock
	}
	return nil
}

// WriteJson writes JSON data
func (e *UpdateEncoderV2) WriteJson(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return lib0.WriteVarString(e.restWriter, string(jsonData))
}

// Dispose releases resources used by the encoder
func (e *UpdateEncoderV2) Dispose() {
	if !e.disposed {
		e.keyMap = nil
		e.keyEncoder.Dispose()
		e.clientEncoder.Dispose()
		e.leftEncoder.Dispose()
		e.rightEncoder.Dispose()
		e.infoEncoder.Dispose()
		e.stringEncoder.Dispose()
		e.parentEncoder.Dispose()
		e.typeRefEncoder.Dispose()
		e.lengthEncoder.Dispose()
		e.DSEncoderV2.Dispose()
		e.disposed = true
	}
}
