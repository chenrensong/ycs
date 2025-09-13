package encoding

import lib0 "github.com/chenrensong/ygo/lib0"

// IntDiffEncoder encodes integer differences using variable length encoding.
type IntDiffEncoder struct {
	*AbstractStreamEncoder[int64]
	state int64
}

// NewIntDiffEncoder creates a new instance of IntDiffEncoder with a starting value.
func NewIntDiffEncoder(start int64) *IntDiffEncoder {
	return &IntDiffEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder[int64](),
		state:                 start,
	}
}

// Write writes an integer value to the encoder as a difference from the previous value.
func (e *IntDiffEncoder) Write(value int64) error {
	if err := e.CheckDisposed(); err != nil {
		return err
	}

	diff := value - e.state
	if err := lib0.WriteVarInt(e.buffer, diff, nil); err != nil {
		return err
	}
	e.state = value
	return nil
}
