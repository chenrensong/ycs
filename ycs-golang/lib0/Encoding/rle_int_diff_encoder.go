package encoding

import lib0 "github.com/chenrensong/ygo/lib0"

// RleIntDiffEncoder combines IntDiffEncoder and RleEncoder functionality.
type RleIntDiffEncoder struct {
	*AbstractStreamEncoder[int64]
	state int64
	count uint32
}

// NewRleIntDiffEncoder creates a new instance of RleIntDiffEncoder with a starting value.
func NewRleIntDiffEncoder(start int64) *RleIntDiffEncoder {
	return &RleIntDiffEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder[int64](),
		state:                 start,
	}
}

// Write writes an integer value to the encoder.
func (e *RleIntDiffEncoder) Write(value int64) error {
	if err := e.CheckDisposed(); err != nil {
		return err
	}

	if e.state == value && e.count > 0 {
		e.count++
	} else {
		if e.count > 0 {
			// Write count (non-standard encoding: count - 1)
			if err := lib0.WriteVarUint(e.buffer, e.count-1); err != nil {
				return err
			}
		}

		// Write difference
		diff := value - e.state
		if err := lib0.WriteVarInt(e.buffer, diff, nil); err != nil {
			return err
		}

		e.count = 1
		e.state = value
	}
	return nil
}

// Flush writes any remaining values to the stream.
func (e *RleIntDiffEncoder) Flush() error {
	if e.count > 0 {
		if err := lib0.WriteVarUint(e.buffer, e.count-1); err != nil {
			return err
		}
	}
	return e.AbstractStreamEncoder.Flush()
}
