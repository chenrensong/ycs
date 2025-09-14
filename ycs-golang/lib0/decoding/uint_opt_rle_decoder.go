// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package decoding

import (
	"io"

	"ycs/lib0"
)

// UintOptRleDecoder decodes optimized run-length encoded unsigned integers.
type UintOptRleDecoder struct {
	*AbstractStreamDecoder[uint]
	state uint
	count uint
}

// NewUintOptRleDecoder creates a new instance of UintOptRleDecoder.
func NewUintOptRleDecoder(stream io.ReadSeekCloser, leaveOpen bool) *UintOptRleDecoder {
	return &UintOptRleDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder[uint](stream, leaveOpen),
	}
}

// Read reads the next unsigned integer from the stream.
func (d *UintOptRleDecoder) Read() (uint, error) {
	d.CheckDisposed()

	if d.count == 0 {
		// Type assert to StreamReader interface
		streamReader, ok := d.Stream().(lib0.StreamReader)
		if !ok {
			return 0, &lib0.TypeAssertionError{Message: "failed to convert stream to StreamReader"}
		}

		value, sign, err := lib0.ReadVarInt(streamReader)
		if err != nil {
			return 0, err
		}

		// If the sign is negative, we read the count too; otherwise, count is 1
		if sign < 0 {
			count, err := lib0.ReadVarUint(streamReader)
			if err != nil {
				return 0, err
			}
			d.state = uint(-value)
			d.count = uint(count) + 2
		} else {
			d.state = uint(value)
			d.count = 1
		}
	}

	d.count--
	return d.state, nil
}
