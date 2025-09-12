package decoding

import (
	"io"
)

// RleIntDiffDecoder represents a decoder for RLE integer difference encoding.
type RleIntDiffDecoder struct {
	*AbstractStreamDecoder
	state int64
	count int64
}

// NewRleIntDiffDecoder creates a new RleIntDiffDecoder.
func NewRleIntDiffDecoder(input io.Reader, start int64, leaveOpen bool) *RleIntDiffDecoder {
	return &RleIntDiffDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder(input, leaveOpen),
		state:                 start,
		count:                 0,
	}
}

// Read reads the next element from the underlying data stream.
func (d *RleIntDiffDecoder) Read() int64 {
	// In Go, we don't have the same disposed checking mechanism as in C#
	// but we can add a check to see if the stream is nil
	if d.Stream() == nil {
		panic("decoder has been closed")
	}

	if d.count == 0 {
		// Read variable length integer
		value, err := readVarInt64(d.Stream())
		if err != nil {
			panic(err)
		}

		// Update state with the difference
		d.state += value

		// Since we don't have the HasContent property implemented yet,
		// we'll need to implement a way to check if there's more content
		// For now, we'll assume there's content and read the count
		count, err := readVarUint(d.Stream())
		if err != nil {
			// If we can't read the count, it means we've reached the end
			// Read the current value forever
			d.count = -1
		} else {
			// See encoder implementation for the reason why this is incremented
			d.count = int64(count) + 1
		}
	}

	d.count--
	return d.state
}