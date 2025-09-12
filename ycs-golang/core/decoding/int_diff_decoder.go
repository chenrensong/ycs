package decoding

import (
	"io"
)

// IntDiffDecoder represents a decoder for integer difference encoding.
type IntDiffDecoder struct {
	*AbstractStreamDecoder
	state int64
}

// NewIntDiffDecoder creates a new IntDiffDecoder.
func NewIntDiffDecoder(input io.Reader, start int64, leaveOpen bool) *IntDiffDecoder {
	return &IntDiffDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder(input, leaveOpen),
		state:                 start,
	}
}

// Read reads the next integer difference from the underlying data stream.
func (d *IntDiffDecoder) Read() int64 {
	// In Go, we don't have the same disposed checking mechanism as in C#
	// but we can add a check to see if the stream is nil
	if d.Stream() == nil {
		panic("decoder has been closed")
	}

	// Read variable length integer
	value, err := readVarInt64(d.Stream())
	if err != nil {
		panic(err)
	}

	// Update state with the difference
	d.state += value
	return d.state
}

// Placeholder implementation for readVarInt64
// This should be replaced with actual implementation from the core package
func readVarInt64(r io.Reader) (int64, error) {
	// TODO: Implement proper varint decoding for int64
	return 0, nil
}