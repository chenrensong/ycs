package core

import (
	"bytes"
	"io"
	"ycs/lib0/encoding"
)

// DSEncoderV2 represents a delete set encoder version 2
type DSEncoderV2 struct {
	dsCurVal   int64
	restWriter *bytes.Buffer
	disposed   bool
}

// NewDSEncoderV2 creates a new DSEncoderV2
func NewDSEncoderV2() *DSEncoderV2 {
	return &DSEncoderV2{
		dsCurVal:   0,
		restWriter: &bytes.Buffer{},
		disposed:   false,
	}
}

// GetRestWriter returns the rest writer
func (dse *DSEncoderV2) GetRestWriter() io.Writer {
	return dse.restWriter
}

// ResetDsCurVal resets the delete set current value
func (dse *DSEncoderV2) ResetDsCurVal() {
	dse.dsCurVal = 0
}

// WriteDsClock writes a delete set clock value
func (dse *DSEncoderV2) WriteDsClock(clock int64) {
	diff := clock - dse.dsCurVal
	if diff < 0 {
		panic("clock diff cannot be negative")
	}
	dse.dsCurVal = clock
	encoding.WriteVarUint(dse.restWriter, uint64(diff))
}

// WriteDsLength writes a delete set length value
func (dse *DSEncoderV2) WriteDsLength(length int64) {
	if length <= 0 {
		panic("length must be positive")
	}
	encoding.WriteVarUint(dse.restWriter, uint64(length-1))
	dse.dsCurVal += length
}

// ToArray returns the encoded bytes
func (dse *DSEncoderV2) ToArray() []byte {
	return dse.restWriter.Bytes()
}

// Close closes the encoder
func (dse *DSEncoderV2) Close() error {
	if !dse.disposed {
		dse.disposed = true
		dse.restWriter = nil
	}
	return nil
}

// UpdateEncoderV2 represents an update encoder version 2
type UpdateEncoderV2 struct {
	*DSEncoderV2
	keyClock          int
	keyMap            map[string]int
	keyClockEncoder   *encoding.IntDiffOptRleEncoder
	clientEncoder     *encoding.UintOptRleEncoder
	leftClockEncoder  *encoding.IntDiffOptRleEncoder
	rightClockEncoder *encoding.IntDiffOptRleEncoder
	infoEncoder       *encoding.RleEncoder
	stringEncoder     *encoding.StringEncoder
	parentInfoEncoder *encoding.RleEncoder
	typeRefEncoder    *encoding.UintOptRleEncoder
	lengthEncoder     *encoding.UintOptRleEncoder
}

// NewUpdateEncoderV2 creates a new UpdateEncoderV2
func NewUpdateEncoderV2() *UpdateEncoderV2 {
	ue := &UpdateEncoderV2{
		DSEncoderV2:       NewDSEncoderV2(),
		keyClock:          0,
		keyMap:            make(map[string]int),
		keyClockEncoder:   encoding.NewIntDiffOptRleEncoder(),
		clientEncoder:     encoding.NewUintOptRleEncoder(),
		leftClockEncoder:  encoding.NewIntDiffOptRleEncoder(),
		rightClockEncoder: encoding.NewIntDiffOptRleEncoder(),
		infoEncoder:       encoding.NewRleEncoder(),
		stringEncoder:     encoding.NewStringEncoder(),
		parentInfoEncoder: encoding.NewRleEncoder(),
		typeRefEncoder:    encoding.NewUintOptRleEncoder(),
		lengthEncoder:     encoding.NewUintOptRleEncoder(),
	}
	return ue
}

// WriteLeftID writes a left ID
func (ue *UpdateEncoderV2) WriteLeftID(id StructID) {
	ue.clientEncoder.Write(uint64(id.Client))
	ue.leftClockEncoder.Write(int64(id.Clock))
}

// WriteRightID writes a right ID
func (ue *UpdateEncoderV2) WriteRightID(id StructID) {
	ue.clientEncoder.Write(uint64(id.Client))
	ue.rightClockEncoder.Write(int64(id.Clock))
}

// WriteClient writes a client ID
func (ue *UpdateEncoderV2) WriteClient(client int64) {
	ue.clientEncoder.Write(uint64(client))
}

// WriteInfo writes info byte
func (ue *UpdateEncoderV2) WriteInfo(info byte) {
	ue.infoEncoder.Write(uint64(info))
}

// WriteString writes a string
func (ue *UpdateEncoderV2) WriteString(s string) {
	ue.stringEncoder.Write(s)
}

// WriteParentInfo writes parent info
func (ue *UpdateEncoderV2) WriteParentInfo(isYKey bool) {
	if isYKey {
		ue.parentInfoEncoder.Write(1)
	} else {
		ue.parentInfoEncoder.Write(0)
	}
}

// WriteTypeRef writes a type reference
func (ue *UpdateEncoderV2) WriteTypeRef(typeRef uint32) {
	ue.typeRefEncoder.Write(uint64(typeRef))
}

// WriteLength writes a length
func (ue *UpdateEncoderV2) WriteLength(length int) {
	ue.lengthEncoder.Write(uint64(length))
}

// WriteKey writes a key with optional caching
func (ue *UpdateEncoderV2) WriteKey(key string) {
	if keyId, exists := ue.keyMap[key]; exists {
		ue.keyClockEncoder.Write(int64(keyId))
	} else {
		ue.keyMap[key] = ue.keyClock
		ue.keyClockEncoder.Write(int64(ue.keyClock))
		ue.keyClock++
		ue.stringEncoder.Write(key)
	}
}

// ToArray finalizes encoding and returns the complete byte array
func (ue *UpdateEncoderV2) ToArray() []byte {
	// Finalize all encoders and write to the main buffer
	ue.writeEncoders()
	return ue.DSEncoderV2.ToArray()
}

// writeEncoders writes all encoder contents to the main buffer
func (ue *UpdateEncoderV2) writeEncoders() {
	// Write in the order expected by the decoder
	ue.keyClockEncoder.WriteTo(ue.restWriter)
	ue.clientEncoder.WriteTo(ue.restWriter)
	ue.leftClockEncoder.WriteTo(ue.restWriter)
	ue.rightClockEncoder.WriteTo(ue.restWriter)
	ue.infoEncoder.WriteTo(ue.restWriter)
	ue.stringEncoder.WriteTo(ue.restWriter)
	ue.parentInfoEncoder.WriteTo(ue.restWriter)
	ue.typeRefEncoder.WriteTo(ue.restWriter)
	ue.lengthEncoder.WriteTo(ue.restWriter)
}

// Close closes the encoder and all sub-encoders
func (ue *UpdateEncoderV2) Close() error {
	if !ue.disposed {
		// Close all sub-encoders
		ue.keyClockEncoder.Close()
		ue.clientEncoder.Close()
		ue.leftClockEncoder.Close()
		ue.rightClockEncoder.Close()
		ue.infoEncoder.Close()
		ue.stringEncoder.Close()
		ue.parentInfoEncoder.Close()
		ue.typeRefEncoder.Close()
		ue.lengthEncoder.Close()

		// Close the base encoder
		ue.DSEncoderV2.Close()
	}
	return nil
}
