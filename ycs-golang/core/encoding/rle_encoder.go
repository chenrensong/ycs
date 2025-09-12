package encoding

// RleEncoder represents a basic Run Length Encoder.
type RleEncoder struct {
	*AbstractStreamEncoder
	state *byte
	count uint32
}

// NewRleEncoder creates a new RleEncoder.
func NewRleEncoder() *RleEncoder {
	return &RleEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder(),
		state:                 nil,
		count:                 0,
	}
}

// Write writes a value to the underlying data stream.
func (e *RleEncoder) Write(value byte) {
	e.CheckDisposed()

	if e.state != nil && *e.state == value {
		e.count++
	} else {
		if e.count > 0 {
			// Flush counter, unless this is the first value (count = 0).
			// Since 'count' is always >0, we can decrement by one. Non-standard encoding.
			writeVarUint(e.buffer, e.count-1)
		}

		// Write the byte value to the buffer
		e.buffer.WriteByte(value)

		e.count = 1
		e.state = &value
	}
}