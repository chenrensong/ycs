// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package encoding

import "github.com/chenrensong/ygo/lib0"

// UintOptRleEncoder implements optimized run-length encoding for unsigned integers.
type UintOptRleEncoder struct {
	*AbstractStreamEncoder[uint32]
	state uint32
	count uint32
}

// NewUintOptRleEncoder creates a new instance of UintOptRleEncoder.
func NewUintOptRleEncoder() *UintOptRleEncoder {
	return &UintOptRleEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder[uint32](),
	}
}

// Write writes an unsigned integer value to the encoder.
func (e *UintOptRleEncoder) Write(value uint32) error {
	if err := e.CheckDisposed(); err != nil {
		return err
	}

	if e.state == value {
		e.count++
	} else {
		if err := e.writeEncodedValue(); err != nil {
			return err
		}
		e.count = 1
		e.state = value
	}
	return nil
}

// Flush writes any remaining values to the stream.
func (e *UintOptRleEncoder) Flush() error {
	if err := e.writeEncodedValue(); err != nil {
		return err
	}
	return e.AbstractStreamEncoder.Flush()
}

// writeEncodedValue writes the current state and count to the stream.
func (e *UintOptRleEncoder) writeEncodedValue() error {
	if e.count == 0 {
		return nil
	}

	if e.count == 1 {
		// Single value - write as positive varint
		if err := lib0.WriteVarInt(e.buffer, int64(e.state), nil); err != nil {
			return err
		}
	} else {
		// Multiple values - write as negative varint followed by count
		var encodedValue int64
		if e.state == 0 {
			// Special case for zero to ensure it's treated as negative
			encodedValue = 0
		} else {
			encodedValue = -int64(e.state)
		}

		if err := lib0.WriteVarInt(e.buffer, encodedValue, nil); err != nil {
			return err
		}

		// Write count (non-standard encoding: count - 2)
		if err := lib0.WriteVarUint(e.buffer, e.count-2); err != nil {
			return err
		}
	}
	return nil
}
