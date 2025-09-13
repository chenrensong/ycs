// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/chenrensong/ygo/lib0"
	"github.com/chenrensong/ygo/lib0/decoding"
)

var (
	ErrNegativeValue = errors.New("negative value not allowed")
)

// DSDecoderV2 is a document state decoder for version 2
type DSDecoderV2 struct {
	reader    io.Reader
	leaveOpen bool
	dsCurVal  int64
	disposed  bool
}

// NewDSDecoderV2 creates a new DSDecoderV2 instance
func NewDSDecoderV2(reader io.Reader, leaveOpen bool) *DSDecoderV2 {
	return &DSDecoderV2{
		reader:    reader,
		leaveOpen: leaveOpen,
	}
}

// ResetDsCurVal resets the current document state value
func (d *DSDecoderV2) ResetDsCurVal() {
	d.dsCurVal = 0
}

// ReadDsClock reads a document state clock value
func (d *DSDecoderV2) ReadDsClock() (int64, error) {
	diff, err := lib0.ReadVarUint(d.reader.(lib0.StreamReader))
	if err != nil {
		return 0, err
	}
	d.dsCurVal += int64(diff)
	if d.dsCurVal < 0 {
		return 0, ErrNegativeValue
	}
	return d.dsCurVal, nil
}

// ReadDsLength reads a document state length
func (d *DSDecoderV2) ReadDsLength() (int64, error) {
	diff, err := lib0.ReadVarUint(d.reader.(lib0.StreamReader))
	if err != nil {
		return 0, err
	}
	length := int64(diff) + 1
	if length <= 0 {
		return 0, ErrNegativeValue
	}
	d.dsCurVal += length
	return length, nil
}

// Dispose releases resources used by the decoder
func (d *DSDecoderV2) Dispose() {
	if !d.disposed {
		if !d.leaveOpen {
			if closer, ok := d.reader.(io.Closer); ok {
				closer.Close()
			}
		}
		d.reader = nil
		d.disposed = true
	}
}

// UpdateDecoderV2 extends DSDecoderV2 for update decoding
type UpdateDecoderV2 struct {
	*DSDecoderV2

	keys           []string
	keyDecoder     *decoding.IntDiffOptRleDecoder
	clientDecoder  *decoding.UintOptRleDecoder
	leftDecoder    *decoding.IntDiffOptRleDecoder
	rightDecoder   *decoding.IntDiffOptRleDecoder
	infoDecoder    *decoding.RleDecoder
	stringDecoder  *decoding.StringDecoder
	parentDecoder  *decoding.RleDecoder
	typeRefDecoder *decoding.UintOptRleDecoder
	lengthDecoder  *decoding.UintOptRleDecoder
}

// NewUpdateDecoderV2 creates a new UpdateDecoderV2 instance
func NewUpdateDecoderV2(reader io.Reader, leaveOpen bool) (*UpdateDecoderV2, error) {
	dsDecoder := NewDSDecoderV2(reader, leaveOpen)

	// Read feature flag (currently unused)
	if _, err := lib0.ReadByte(reader.(lib0.StreamReader)); err != nil {
		return nil, err
	}

	// Initialize all decoders
	keyDecoder, err := newDecoder(reader, decoding.NewIntDiffOptRleDecoder)
	if err != nil {
		return nil, err
	}

	clientDecoder, err := newDecoder(reader, decoding.NewUintOptRleDecoder)
	if err != nil {
		return nil, err
	}

	leftDecoder, err := newDecoder(reader, decoding.NewIntDiffOptRleDecoder)
	if err != nil {
		return nil, err
	}

	rightDecoder, err := newDecoder(reader, decoding.NewIntDiffOptRleDecoder)
	if err != nil {
		return nil, err
	}

	infoDecoder, err := newDecoder(reader, decoding.NewRleDecoder)
	if err != nil {
		return nil, err
	}

	stringDecoder, err := newDecoder(reader, decoding.NewStringDecoder)
	if err != nil {
		return nil, err
	}

	parentDecoder, err := newDecoder(reader, decoding.NewRleDecoder)
	if err != nil {
		return nil, err
	}

	typeRefDecoder, err := newDecoder(reader, decoding.NewUintOptRleDecoder)
	if err != nil {
		return nil, err
	}

	lengthDecoder, err := newDecoder(reader, decoding.NewUintOptRleDecoder)
	if err != nil {
		return nil, err
	}

	return &UpdateDecoderV2{
		DSDecoderV2:    dsDecoder,
		keys:           make([]string, 0),
		keyDecoder:     keyDecoder,
		clientDecoder:  clientDecoder,
		leftDecoder:    leftDecoder,
		rightDecoder:   rightDecoder,
		infoDecoder:    infoDecoder,
		stringDecoder:  stringDecoder,
		parentDecoder:  parentDecoder,
		typeRefDecoder: typeRefDecoder,
		lengthDecoder:  lengthDecoder,
	}, nil
}

func newDecoder[T any](reader io.Reader, constructor func(io.ReadSeekCloser, bool) T) (T, error) {
	data, err := lib0.ReadVarUint8Array(reader.(lib0.StreamReader))
	if err != nil {
		var zero T
		return zero, err
	}
	// Create a ReadSeekCloser from the byte slice
	bytesReader := &readSeekCloser{bytes.NewReader(data)}
	return constructor(bytesReader, false), nil
}

// readSeekCloser wraps bytes.Reader to implement io.ReadSeekCloser
type readSeekCloser struct {
	*bytes.Reader
}

func (r *readSeekCloser) Close() error {
	return nil // bytes.Reader doesn't need closing
}

// ReadLeftId reads a left ID
func (d *UpdateDecoderV2) ReadLeftId() (*ID, error) {
	if d.disposed {
		return nil, errors.New("decoder disposed")
	}

	client, err := d.clientDecoder.Read()
	if err != nil {
		return nil, err
	}

	clock, err := d.leftDecoder.Read()
	if err != nil {
		return nil, err
	}

	return NewID(uint64(client), int(clock)), nil
}

// ReadRightId reads a right ID
func (d *UpdateDecoderV2) ReadRightId() (*ID, error) {
	if d.disposed {
		return nil, errors.New("decoder disposed")
	}

	client, err := d.clientDecoder.Read()
	if err != nil {
		return nil, err
	}

	clock, err := d.rightDecoder.Read()
	if err != nil {
		return nil, err
	}

	return NewID(uint64(client), int(clock)), nil
}

// ReadClient reads a client ID
func (d *UpdateDecoderV2) ReadClient() (uint64, error) {
	if d.disposed {
		return 0, errors.New("decoder disposed")
	}
	client, err := d.clientDecoder.Read()
	return uint64(client), err
}

// ReadInfo reads an info byte
func (d *UpdateDecoderV2) ReadInfo() (byte, error) {
	if d.disposed {
		return 0, errors.New("decoder disposed")
	}
	return d.infoDecoder.Read()
}

// ReadString reads a string
func (d *UpdateDecoderV2) ReadString() (string, error) {
	if d.disposed {
		return "", errors.New("decoder disposed")
	}
	return d.stringDecoder.Read()
}

// ReadParentInfo reads parent info
func (d *UpdateDecoderV2) ReadParentInfo() (bool, error) {
	if d.disposed {
		return false, errors.New("decoder disposed")
	}
	val, err := d.parentDecoder.Read()
	return val == 1, err
}

// ReadTypeRef reads a type reference
func (d *UpdateDecoderV2) ReadTypeRef() (uint64, error) {
	if d.disposed {
		return 0, errors.New("decoder disposed")
	}
	typeRef, err := d.typeRefDecoder.Read()
	return uint64(typeRef), err
}

// ReadLength reads a length
func (d *UpdateDecoderV2) ReadLength() (int, error) {
	if d.disposed {
		return 0, errors.New("decoder disposed")
	}
	length, err := d.lengthDecoder.Read()
	if err != nil {
		return 0, err
	}
	if length < 0 {
		return 0, ErrNegativeValue
	}
	return int(length), nil
}

// ReadAny reads any value
func (d *UpdateDecoderV2) ReadAny() (interface{}, error) {
	if d.disposed {
		return nil, errors.New("decoder disposed")
	}
	return lib0.ReadAny(d.reader.(lib0.StreamReader))
}

// ReadBuffer reads a byte buffer
func (d *UpdateDecoderV2) ReadBuffer() ([]byte, error) {
	if d.disposed {
		return nil, errors.New("decoder disposed")
	}
	return lib0.ReadVarUint8Array(d.reader.(lib0.StreamReader))
}

// ReadKey reads a key with optimization for frequently used keys
func (d *UpdateDecoderV2) ReadKey() (string, error) {
	if d.disposed {
		return "", errors.New("decoder disposed")
	}

	keyClock, err := d.keyDecoder.Read()
	if err != nil {
		return "", err
	}

	if int(keyClock) < len(d.keys) {
		return d.keys[keyClock], nil
	}

	key, err := d.stringDecoder.Read()
	if err != nil {
		return "", err
	}

	d.keys = append(d.keys, key)
	return key, nil
}

// ReadJson reads JSON data
func (d *UpdateDecoderV2) ReadJson() (interface{}, error) {
	if d.disposed {
		return nil, errors.New("decoder disposed")
	}

	jsonStr, err := lib0.ReadVarString(d.reader.(lib0.StreamReader))
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Dispose releases resources used by the decoder
func (d *UpdateDecoderV2) Dispose() {
	if !d.disposed {
		d.keyDecoder.Dispose()
		d.clientDecoder.Dispose()
		d.leftDecoder.Dispose()
		d.rightDecoder.Dispose()
		d.infoDecoder.Dispose()
		d.stringDecoder.Dispose()
		d.parentDecoder.Dispose()
		d.typeRefDecoder.Dispose()
		d.lengthDecoder.Dispose()
		d.DSDecoderV2.Dispose()
		d.disposed = true
	}
}
