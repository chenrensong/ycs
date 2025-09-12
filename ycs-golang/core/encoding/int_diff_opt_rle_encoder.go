package encoding

// IntDiffOptRleEncoder represents a combination of the IntDiffEncoder and the UintOptRleEncoder.
type IntDiffOptRleEncoder struct {
	*AbstractStreamEncoder
	state int64
	diff  int64
	count uint32
}

// NewIntDiffOptRleEncoder creates a new IntDiffOptRleEncoder.
func NewIntDiffOptRleEncoder() *IntDiffOptRleEncoder {
	return &IntDiffOptRleEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder(),
		state:                 0,
		diff:                  0,
		count:                 0,
	}
}

// Write writes a value to the underlying data stream.
func (e *IntDiffOptRleEncoder) Write(value int64) {
	// In the Go version, we don't have the Debug.Assert from C#,
	// but we could add a check if needed:
	// if value > Bits30 {
	//     panic("value exceeds 31-bit range")
	// }
	
	e.CheckDisposed()

	if e.diff == value-e.state {
		e.state = value
		e.count++
	} else {
		e.writeEncodedValue()

		e.count = 1
		e.diff = value - e.state
		e.state = value
	}
}

// Flush flushes any buffered data.
func (e *IntDiffOptRleEncoder) Flush() {
	e.writeEncodedValue()
	e.AbstractStreamEncoder.Flush()
}

// writeEncodedValue writes the encoded value to the stream.
func (e *IntDiffOptRleEncoder) writeEncodedValue() {
	if e.count > 0 {
		var encodedDiff int64
		if e.diff < 0 {
			encodedDiff = -(((-int64(e.diff))<<1)|int64(func() uint32 {
				if e.count == 1 {
					return 0
				}
				return 1
			}()))
		} else {
			// 31bit making up a diff | whether to write the counter.
			encodedDiff = (int64(e.diff)<<1)|int64(func() uint32 {
				if e.count == 1 {
					return 0
				}
				return 1
			}())
		}

		writeVarInt64(e.buffer, encodedDiff)

		if e.count > 1 {
			// Since count is always >1, we can decrement by one. Non-standard encoding.
			writeVarUint(e.buffer, e.count-2)
		}
	}
}