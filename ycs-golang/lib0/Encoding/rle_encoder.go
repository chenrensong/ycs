// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package encoding

import (
	"ycs/lib0"
)

// RleEncoder implements basic run-length encoding for byte values.
type RleEncoder struct {
	*AbstractStreamEncoder[byte]
	state *byte
	count uint32
}

// NewRleEncoder creates a new instance of RleEncoder.
func NewRleEncoder() *RleEncoder {
	return &RleEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder[byte](),
	}
}

// Write writes a byte value to the encoder.
func (e *RleEncoder) Write(value byte) error {
	if err := e.CheckDisposed(); err != nil {
		return err
	}

	if e.state != nil && *e.state == value {
		e.count++
	} else {
		writer, err := e.GetWriter()
		if err != nil {
			return err
		}

		if e.count > 0 {
			// Flush counter (non-standard encoding: count - 1)
			if err := lib0.WriteVarUint(writer, e.count-1); err != nil {
				return err
			}
		}

		if _, err := writer.Write([]byte{value}); err != nil {
			return err
		}

		e.count = 1
		e.state = &value
	}
	return nil
}

// Flush writes any remaining values to the stream.
func (e *RleEncoder) Flush() error {
	if e.count > 0 && e.state != nil {
		writer, err := e.GetWriter()
		if err != nil {
			return err
		}

		if err := lib0.WriteVarUint(writer, e.count-1); err != nil {
			return err
		}
	}
	return e.AbstractStreamEncoder.Flush()
}
