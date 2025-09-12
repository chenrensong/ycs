package decoding

import (
	"io"
)

// RleDecoder represents a decoder for RLE encoding.
type RleDecoder struct {
	*AbstractStreamDecoder
	state byte
	count int64
}

// NewRleDecoder creates a new RleDecoder.
func NewRleDecoder(input io.Reader, leaveOpen bool) *RleDecoder {
	return &RleDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder(input, leaveOpen),
		state:                 0,
		count:                 0,
	}
}

// Read reads the next element from the underlying data stream.
func (d *RleDecoder) Read() byte {
	// In Go, we don't have the same disposed checking mechanism as in C#
	// but we can add a check to see if the stream is nil
	if d.Stream() == nil {
		panic("decoder has been closed")
	}

	if d.count == 0 {
		// Read a single byte
		buf := make([]byte, 1)
		_, err := io.ReadFull(d.Stream(), buf)
		if err != nil {
			panic(err)
		}
		d.state = buf[0]

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