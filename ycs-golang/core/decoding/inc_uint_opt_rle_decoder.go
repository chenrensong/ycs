package decoding

import (
	"io"
)

// IncUintOptRleDecoder represents a decoder for incremental unsigned integer RLE encoding.
type IncUintOptRleDecoder struct {
	*AbstractStreamDecoder
	state uint32
	count uint32
}

// NewIncUintOptRleDecoder creates a new IncUintOptRleDecoder.
func NewIncUintOptRleDecoder(input io.Reader, leaveOpen bool) *IncUintOptRleDecoder {
	return &IncUintOptRleDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder(input, leaveOpen),
		state:                 0,
		count:                 0,
	}
}

// Read reads the next element from the underlying data stream.
func (d *IncUintOptRleDecoder) Read() uint32 {
	// In Go, we don't have the same disposed checking mechanism as in C#
	// but we can add a check to see if the stream is nil
	if d.Stream() == nil {
		panic("decoder has been closed")
	}

	if d.count == 0 {
		// Read variable length integer
		value, err := readVarInt32(d.Stream())
		if err != nil {
			panic(err)
		}

		sign := int32(1)
		if value < 0 {
			sign = -1
		}

		// If the sign is negative, we read the count too; otherwise, count is 1.
		isNegative := sign < 0
		if isNegative {
			d.state = uint32(-value)
			// Read variable length unsigned integer
			count, err := readVarUint(d.Stream())
			if err != nil {
				panic(err)
			}
			d.count = count + 2
		} else {
			d.state = uint32(value)
			d.count = 1
		}
	}

	d.count--
	result := d.state
	d.state++
	return result
}

// Placeholder implementation for readVarInt32
// This should be replaced with actual implementation from the core package
func readVarInt32(r io.Reader) (int32, error) {
	// TODO: Implement proper varint decoding for int32
	return 0, nil
}