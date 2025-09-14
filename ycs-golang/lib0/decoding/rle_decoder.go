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

// streamWrapper wraps io.ReadSeekCloser to implement lib0.StreamReader
type streamWrapper struct {
	io.ReadSeekCloser
}

// ReadByte implements io.ByteReader interface
func (w *streamWrapper) ReadByte() (byte, error) {
	b := make([]byte, 1)
	_, err := w.Read(b)
	return b[0], err
}

// RleDecoder decodes run-length encoded bytes.
type RleDecoder struct {
	*AbstractStreamDecoder[byte]
	state byte
	count int64
}

// NewRleDecoder creates a new instance of RleDecoder.
func NewRleDecoder(stream io.ReadSeekCloser, leaveOpen bool) *RleDecoder {
	return &RleDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder[byte](stream, leaveOpen),
		count:                 0,
	}
}

// Read reads the next byte from the stream.
func (d *RleDecoder) Read() (byte, error) {
	d.CheckDisposed()

	// Type assert to StreamReader interface
	streamReader, ok := d.Stream().(lib0.StreamReader)
	if !ok {
		return 0, &lib0.TypeAssertionError{Message: "failed to convert stream to StreamReader"}
	}

	if d.count == 0 {
		var err error
		d.state, err = lib0.ReadByte(streamReader)
		if err != nil {
			return 0, err
		}

		if d.HasContent() {
			// See encoder implementation for the reason why this is incremented.
			count, err := lib0.ReadVarUint(streamReader)
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
