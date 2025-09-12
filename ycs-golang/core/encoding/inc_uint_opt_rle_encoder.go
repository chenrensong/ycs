package encoding

// IncUintOptRleEncoder represents an increasing unsigned integer optimized RLE encoder.
type IncUintOptRleEncoder struct {
	*AbstractStreamEncoder
	state uint32
	count uint32
}

// NewIncUintOptRleEncoder creates a new IncUintOptRleEncoder.
func NewIncUintOptRleEncoder() *IncUintOptRleEncoder {
	return &IncUintOptRleEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder(),
		state:                 0,
		count:                 0,
	}
}

// Write writes a value to the underlying data stream.
func (e *IncUintOptRleEncoder) Write(value uint32) {
	// In the Go version, we don't have the Debug.Assert from C#,
	// but we could add a check if needed:
	// if value > math.MaxInt32 {
	//     panic("value exceeds int32 range")
	// }
	
	e.CheckDisposed()

	if e.state+e.count == value {
		e.count++
	} else {
		e.writeEncodedValue()

		e.count = 1
		e.state = value
	}
}

// Flush flushes any buffered data.
func (e *IncUintOptRleEncoder) Flush() {
	e.writeEncodedValue()
	e.AbstractStreamEncoder.Flush()
}

// writeEncodedValue writes the encoded value to the stream.
func (e *IncUintOptRleEncoder) writeEncodedValue() {
	if e.count > 0 {
		// Flush counter, unless this is the first value (count = 0).
		// Case 1: Just a single value. Set sign to positive.
		// Case 2: Write several values. Set sign to negative to indicate that there is a length coming.
		if e.count == 1 {
			writeVarInt32(e.buffer, int32(e.state))
		} else {
			// Specify 'treatZeroAsNegative' in case we pass the '-0' value.
			writeVarInt32WithNegative(e.buffer, -int32(e.state), e.state == 0)

			// Since count is always >1, we can decrement by one. Non-standard encoding.
			writeVarUint(e.buffer, e.count-2)
		}
	}
}

// Placeholder implementations for writeVarInt32 and writeVarUint
// These should be replaced with actual implementations from the core package
func writeVarInt32(w *bytes.Buffer, value int32) {
	// TODO: Implement proper varint encoding for int32
}

func writeVarInt32WithNegative(w *bytes.Buffer, value int32, treatZeroAsNegative bool) {
	// TODO: Implement proper varint encoding with negative support for int32
}