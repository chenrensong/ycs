package encoding

// RleIntDiffEncoder represents a combination of IntDiffEncoder and RleEncoder.
type RleIntDiffEncoder struct {
	*AbstractStreamEncoder
	state int64
	count uint32
}

// NewRleIntDiffEncoder creates a new RleIntDiffEncoder.
func NewRleIntDiffEncoder(start int64) *RleIntDiffEncoder {
	return &RleIntDiffEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder(),
		state:                 start,
		count:                 0,
	}
}

// Write writes a value to the underlying data stream.
func (e *RleIntDiffEncoder) Write(value int64) {
	e.CheckDisposed()

	if e.state == value && e.count > 0 {
		e.count++
	} else {
		if e.count > 0 {
			writeVarUint(e.buffer, e.count-1)
		}

		writeVarInt64(e.buffer, value-e.state)

		e.count = 1
		e.state = value
	}
}