package encoding

import (
	"errors"

	"ycs/lib0"
)

const (
	bits30 = 1<<30 - 1
)

// IntDiffOptRleEncoder combines IntDiffEncoder and UintOptRleEncoder functionality.
type IntDiffOptRleEncoder struct {
	*AbstractStreamEncoder[int64]
	state int64
	diff  int64
	count uint32
}

// NewIntDiffOptRleEncoder creates a new instance of IntDiffOptRleEncoder.
func NewIntDiffOptRleEncoder() *IntDiffOptRleEncoder {
	return &IntDiffOptRleEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder[int64](),
	}
}

// Write writes an integer value to the encoder.
func (e *IntDiffOptRleEncoder) Write(value int64) error {
	if value > bits30 || value < -bits30 {
		return errors.New("value exceeds 31-bit range")
	}

	if err := e.CheckDisposed(); err != nil {
		return err
	}

	if e.diff == value-e.state {
		e.state = value
		e.count++
	} else {
		if err := e.writeEncodedValue(); err != nil {
			return err
		}
		e.count = 1
		e.diff = value - e.state
		e.state = value
	}
	return nil
}

// Flush writes any remaining values to the stream.
func (e *IntDiffOptRleEncoder) Flush() error {
	if err := e.writeEncodedValue(); err != nil {
		return err
	}
	return e.AbstractStreamEncoder.Flush()
}

// writeEncodedValue writes the current diff and count to the stream.
func (e *IntDiffOptRleEncoder) writeEncodedValue() error {
	if e.count == 0 {
		return nil
	}

	writer, err := e.GetWriter()
	if err != nil {
		return err
	}

	var encodedDiff int64
	var bitFlag uint32
	if e.count == 1 {
		bitFlag = 0
	} else {
		bitFlag = 1
	}

	if e.diff < 0 {
		encodedDiff = -int64((uint32(-e.diff) << 1) | bitFlag)
	} else {
		encodedDiff = int64((uint32(e.diff) << 1) | bitFlag)
	}

	if err := lib0.WriteVarInt(writer, encodedDiff, nil); err != nil {
		return err
	}

	if e.count > 1 {
		if err := lib0.WriteVarUint(writer, uint32(e.count-2)); err != nil {
			return err
		}
	}
	return nil
}
