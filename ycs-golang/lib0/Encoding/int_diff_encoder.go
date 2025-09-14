package encoding

import (
	"ycs/lib0"
)

// IntDiffEncoder encodes integers with difference encoding.
type IntDiffEncoder struct {
	*AbstractStreamEncoder[int64]
	state int64
}

// NewIntDiffEncoder creates a new instance of IntDiffEncoder.
func NewIntDiffEncoder(start int64) *IntDiffEncoder {
	return &IntDiffEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder[int64](),
		state:                 start,
	}
}

// Write writes an integer value to the encoder.
func (e *IntDiffEncoder) Write(value int64) error {
	if err := e.CheckDisposed(); err != nil {
		return err
	}

	diff := value - e.state
	e.state = value

	writer, err := e.GetWriter()
	if err != nil {
		return err
	}

	return lib0.WriteVarInt(writer, diff, nil)
}
