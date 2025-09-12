package decoding

import (
	"io"
)

// IntDiffOptRleDecoder represents a decoder for integer difference optimized RLE encoding.
type IntDiffOptRleDecoder struct {
	*AbstractStreamDecoder
	state int64
	count uint32
	diff  int64
}

// NewIntDiffOptRleDecoder creates a new IntDiffOptRleDecoder.
func NewIntDiffOptRleDecoder(input io.Reader, leaveOpen bool) *IntDiffOptRleDecoder {
	return &IntDiffOptRleDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder(input, leaveOpen),
		state:                 0,
		count:                 0,
		diff:                  0,
	}
}

// Read reads the next element from the underlying data stream.
func (d *IntDiffOptRleDecoder) Read() int64 {
	// In Go, we don't have the same disposed checking mechanism as in C#
	// but we can add a check to see if the stream is nil
	if d.Stream() == nil {
		panic("decoder has been closed")
	}

	if d.count == 0 {
		// Read variable length integer
		diff, err := readVarInt64(d.Stream())
		if err != nil {
			panic(err)
		}

		// If the first bit is set, we read more data.
		hasCount := (diff & 1) > 0

		if diff < 0 {
			d.diff = -((-diff) >> 1)
		} else {
			d.diff = diff >> 1
		}

		if hasCount {
			// Read variable length unsigned integer
			count, err := readVarUint(d.Stream())
			if err != nil {
				panic(err)
			}
			d.count = count + 2
		} else {
			d.count = 1
		}
	}

	d.state += d.diff
	d.count--
	return d.state
}