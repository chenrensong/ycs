package decoding

import (
	"io"

	lib0 "github.com/chenrensong/ygo/lib0"
)

// IncUintOptRleDecoder decodes optimized run-length encoded unsigned integers.
type IncUintOptRleDecoder struct {
	*AbstractStreamDecoder[uint]
	state uint
	count uint
}

// NewIncUintOptRleDecoder creates a new instance of IncUintOptRleDecoder.
func NewIncUintOptRleDecoder(stream io.ReadSeekCloser, leaveOpen bool) *IncUintOptRleDecoder {
	return &IncUintOptRleDecoder{
		AbstractStreamDecoder: NewAbstractStreamDecoder[uint](stream, leaveOpen),
	}
}

// Read reads the next unsigned integer from the stream.
func (d *IncUintOptRleDecoder) Read() (uint, error) {
	d.CheckDisposed()

	if d.count == 0 {
		value, sign, err := lib0.ReadVarInt(d.Stream().(lib0.StreamReader))
		if err != nil {
			return 0, err
		}

		// If the sign is negative, we read the count too; otherwise, count is 1
		if sign < 0 {
			count, err := lib0.ReadVarUint(d.Stream().(lib0.StreamReader))
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
	result := d.state
	d.state++
	return result, nil
}
