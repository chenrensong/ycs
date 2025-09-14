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

// IntDiffOptRleDecoder decodes optimized run-length encoded integer differences.
type IntDiffOptRleDecoder struct {
	*AbstractStreamDecoder[int64]
	state int64
	count uint
	diff  int64
}

// NewIntDiffOptRleDecoder creates a new instance of IntDiffOptRleDecoder.
func NewIntDiffOptRleDecoder(stream io.ReadSeekCloser, leaveOpen bool) *IntDiffOptRleDecoder {
	return &IntDiffOptRleDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder[int64](stream, leaveOpen),
	}
}

// Read reads the next integer from the stream by applying the difference to the current state.
func (d *IntDiffOptRleDecoder) Read() (int64, error) {
	d.CheckDisposed()

	// Type assert to StreamReader interface
	streamReader, ok := d.Stream().(lib0.StreamReader)
	if !ok {
		return 0, &lib0.TypeAssertionError{Message: "failed to convert stream to StreamReader"}
	}

	if d.count == 0 {
		diff, _, err := lib0.ReadVarInt(streamReader)
		if err != nil {
			return 0, err
		}

		// If the first bit is set, we read more data
		hasCount := (diff & 0x1) > 0

		if diff < 0 {
			d.diff = -((-diff) >> 1)
		} else {
			d.diff = diff >> 1
		}

		if hasCount {
			count, err := lib0.ReadVarUint(streamReader)
			if err != nil {
				return 0, err
			}
			d.count = uint(count) + 2
		} else {
			d.count = 1
		}
	}

	d.state += d.diff
	d.count--
	return d.state, nil
}
