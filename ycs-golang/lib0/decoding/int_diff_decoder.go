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

// IntDiffDecoder decodes integer differences from a stream.
type IntDiffDecoder struct {
	*AbstractStreamDecoder[int64]
	state int64
}

// NewIntDiffDecoder creates a new instance of IntDiffDecoder.
func NewIntDiffDecoder(stream io.ReadSeekCloser, start int64, leaveOpen bool) *IntDiffDecoder {
	return &IntDiffDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder[int64](stream, leaveOpen),
		state:                 start,
	}
}

// Read reads the next integer from the stream by applying the difference to the current state.
func (d *IntDiffDecoder) Read() (int64, error) {
	d.CheckDisposed()

	// Type assert to StreamReader interface
	streamReader, ok := d.Stream().(lib0.StreamReader)
	if !ok {
		return 0, &lib0.TypeAssertionError{Message: "failed to convert stream to StreamReader"}
	}

	value, _, err := lib0.ReadVarInt(streamReader)
	if err != nil {
		return 0, err
	}

	d.state += value
	return d.state, nil
}
