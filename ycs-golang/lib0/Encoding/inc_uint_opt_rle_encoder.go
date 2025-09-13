// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package encoding

import lib0 "github.com/chenrensong/ygo/lib0"

// IncUintOptRleEncoder encodes increasing unsigned integers with run-length encoding optimization.
type IncUintOptRleEncoder struct {
	*AbstractStreamEncoder[uint]
	state uint
	count uint
}

// NewIncUintOptRleEncoder creates a new instance of IncUintOptRleEncoder.
func NewIncUintOptRleEncoder() *IncUintOptRleEncoder {
	return &IncUintOptRleEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder[uint](),
	}
}

// Write writes an unsigned integer value to the encoder.
func (e *IncUintOptRleEncoder) Write(value uint) error {
	if err := e.CheckDisposed(); err != nil {
		return err
	}

	if e.state+e.count == value {
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
func (e *IncUintOptRleEncoder) Flush() error {
	if err := e.writeEncodedValue(); err != nil {
		return err
	}
	return e.AbstractStreamEncoder.Flush()
}

// writeEncodedValue writes the current state and count to the stream.
func (e *IncUintOptRleEncoder) writeEncodedValue() error {
	if e.count == 0 {
		return nil
	}

	if e.count == 1 {
		if err := lib0.WriteVarInt(e.buffer, int64(e.state), nil); err != nil {
			return err
		}
	} else {
		// Write negative value to indicate there's a length coming
		negValue := -int64(e.state)
		if e.state == 0 {
			// Special case for zero to ensure it's treated as negative
			negValue = 0
		}
		if err := lib0.WriteVarInt(e.buffer, negValue, nil); err != nil {
			return err
		}

		// Write count (decremented by 2 as per non-standard encoding)
		if err := lib0.WriteVarUint(e.buffer, uint32(e.count-2)); err != nil {
			return err
		}
	}
	return nil
}
