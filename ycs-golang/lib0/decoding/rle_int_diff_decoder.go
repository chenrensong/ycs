// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package decoding

import (
	"io"

	"github.com/chenrensong/ygo/lib0"
)

// RleIntDiffDecoder decodes run-length encoded integer differences.
type RleIntDiffDecoder struct {
	*AbstractStreamDecoder[int64]
	state int64
	count int64
}

// NewRleIntDiffDecoder creates a new instance of RleIntDiffDecoder.
func NewRleIntDiffDecoder(stream io.ReadSeekCloser, start int64, leaveOpen bool) *RleIntDiffDecoder {
	return &RleIntDiffDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder[int64](stream, leaveOpen),
		state:                 start,
		count:                 0,
	}
}

// Read reads the next integer from the stream by applying the difference to the current state.
func (d *RleIntDiffDecoder) Read() (int64, error) {
	d.CheckDisposed()

	if d.count == 0 {
		diff, _, err := lib0.ReadVarInt(d.Stream().(lib0.StreamReader))
		if err != nil {
			return 0, err
		}

		d.state += diff

		if d.HasContent() {
			// See encoder implementation for the reason why this is incremented.
			count, err := lib0.ReadVarUint(d.Stream().(lib0.StreamReader))
			if err != nil {
				return 0, err
			}
			d.count = int64(count) + 1
		} else {
			// Read the current value forever.
			d.count = -1
		}
	}

	if d.count > 0 {
		d.count--
	} else if d.count == -1 {
		// Infinite count case
		return d.state, nil
	}

	return d.state, nil
}
