package core

import (
	"bytes"
	"io"
	"ycs/lib0/decoding"
)

// DSDecoderV2 represents a delete set decoder version 2
type DSDecoderV2 struct {
	dsCurVal  int64
	reader    io.Reader
	leaveOpen bool
	disposed  bool
}

// NewDSDecoderV2 creates a new DSDecoderV2
func NewDSDecoderV2(input io.Reader) *DSDecoderV2 {
	return &DSDecoderV2{
		dsCurVal:  0,
		reader:    input,
		leaveOpen: false,
		disposed:  false,
	}
}

// NewDSDecoderV2FromBytes creates a new DSDecoderV2 from byte array
func NewDSDecoderV2FromBytes(data []byte) *DSDecoderV2 {
	return NewDSDecoderV2(bytes.NewReader(data))
}

// GetReader returns the reader
func (dsd *DSDecoderV2) GetReader() io.Reader {
	return dsd.reader
}

// ResetDsCurVal resets the delete set current value
func (dsd *DSDecoderV2) ResetDsCurVal() {
	dsd.dsCurVal = 0
}

// ReadDsClock reads a delete set clock value
func (dsd *DSDecoderV2) ReadDsClock() int64 {
	diff := decoding.ReadVarUint(dsd.reader)
	dsd.dsCurVal += int64(diff)
	if dsd.dsCurVal < 0 {
		panic("dsCurVal cannot be negative")
	}
	return dsd.dsCurVal
}

// ReadDsLength reads a delete set length value
func (dsd *DSDecoderV2) ReadDsLength() int64 {
	diff := decoding.ReadVarUint(dsd.reader) + 1
	if diff < 0 {
		panic("diff cannot be negative")
	}
	dsd.dsCurVal += int64(diff)
	return int64(diff)
}

// Close closes the decoder
func (dsd *DSDecoderV2) Close() error {
	if !dsd.disposed && !dsd.leaveOpen {
		if closer, ok := dsd.reader.(io.Closer); ok {
			closer.Close()
		}
		dsd.disposed = true
		dsd.reader = nil
	}
	return nil
}

// UpdateDecoderV2 represents an update decoder version 2
type UpdateDecoderV2 struct {
	*DSDecoderV2
	keys              []string
	keyClockDecoder   *decoding.IntDiffOptRleDecoder
	clientDecoder     *decoding.UintOptRleDecoder
	leftClockDecoder  *decoding.IntDiffOptRleDecoder
	rightClockDecoder *decoding.IntDiffOptRleDecoder
	infoDecoder       *decoding.RleDecoder
	stringDecoder     *decoding.StringDecoder
	parentInfoDecoder *decoding.RleDecoder
	typeRefDecoder    *decoding.UintOptRleDecoder
	lengthDecoder     *decoding.UintOptRleDecoder
}

// NewUpdateDecoderV2 creates a new UpdateDecoderV2
func NewUpdateDecoderV2(input io.Reader) *UpdateDecoderV2 {
	ud := &UpdateDecoderV2{
		DSDecoderV2:       NewDSDecoderV2(input),
		keys:              make([]string, 0),
		keyClockDecoder:   decoding.NewIntDiffOptRleDecoder(input),
		clientDecoder:     decoding.NewUintOptRleDecoder(input),
		leftClockDecoder:  decoding.NewIntDiffOptRleDecoder(input),
		rightClockDecoder: decoding.NewIntDiffOptRleDecoder(input),
		infoDecoder:       decoding.NewRleDecoder(input),
		stringDecoder:     decoding.NewStringDecoder(input),
		parentInfoDecoder: decoding.NewRleDecoder(input),
		typeRefDecoder:    decoding.NewUintOptRleDecoder(input),
		lengthDecoder:     decoding.NewUintOptRleDecoder(input),
	}
	return ud
}

// ReadLeftID reads a left ID
func (ud *UpdateDecoderV2) ReadLeftID() StructID {
	client := int64(ud.clientDecoder.Read())
	clock := ud.leftClockDecoder.Read()
	return StructID{Client: client, Clock: clock}
}

// ReadRightID reads a right ID
func (ud *UpdateDecoderV2) ReadRightID() StructID {
	client := int64(ud.clientDecoder.Read())
	clock := ud.rightClockDecoder.Read()
	return StructID{Client: client, Clock: clock}
}

// ReadClient reads a client ID
func (ud *UpdateDecoderV2) ReadClient() int64 {
	return int64(ud.clientDecoder.Read())
}

// ReadInfo reads info byte
func (ud *UpdateDecoderV2) ReadInfo() byte {
	return byte(ud.infoDecoder.Read())
}

// ReadString reads a string
func (ud *UpdateDecoderV2) ReadString() string {
	return ud.stringDecoder.Read()
}

// ReadParentInfo reads parent info
func (ud *UpdateDecoderV2) ReadParentInfo() bool {
	return ud.parentInfoDecoder.Read() == 1
}

// ReadTypeRef reads a type reference
func (ud *UpdateDecoderV2) ReadTypeRef() uint32 {
	return uint32(ud.typeRefDecoder.Read())
}

// ReadLength reads a length
func (ud *UpdateDecoderV2) ReadLength() int {
	return int(ud.lengthDecoder.Read())
}

// ReadKey reads a key with caching support
func (ud *UpdateDecoderV2) ReadKey() string {
	keyId := int(ud.keyClockDecoder.Read())

	// Ensure keys slice is large enough
	for len(ud.keys) <= keyId {
		ud.keys = append(ud.keys, "")
	}

	// If key at this ID doesn't exist, read it from string decoder
	if ud.keys[keyId] == "" {
		ud.keys[keyId] = ud.stringDecoder.Read()
	}

	return ud.keys[keyId]
}

// Close closes the decoder and all sub-decoders
func (ud *UpdateDecoderV2) Close() error {
	if !ud.disposed {
		// Close all sub-decoders
		ud.keyClockDecoder.Close()
		ud.clientDecoder.Close()
		ud.leftClockDecoder.Close()
		ud.rightClockDecoder.Close()
		ud.infoDecoder.Close()
		ud.stringDecoder.Close()
		ud.parentInfoDecoder.Close()
		ud.typeRefDecoder.Close()
		ud.lengthDecoder.Close()

		// Close the base decoder
		ud.DSDecoderV2.Close()
	}
	return nil
}
