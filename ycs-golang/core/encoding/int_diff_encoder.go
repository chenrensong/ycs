package encoding

// IntDiffEncoder represents a basic diff encoder using variable length encoding.
type IntDiffEncoder struct {
	*AbstractStreamEncoder
	state int64
}

// NewIntDiffEncoder creates a new IntDiffEncoder.
func NewIntDiffEncoder(start int64) *IntDiffEncoder {
	return &IntDiffEncoder{
		AbstractStreamEncoder: NewAbstractStreamEncoder(),
		state:                 start,
	}
}

// Write writes an integer value to the underlying data stream.
func (e *IntDiffEncoder) Write(value int64) {
	e.CheckDisposed()

	// Write the difference between the current value and the previous state
	writeVarInt64(e.buffer, value-e.state)
	e.state = value
}

// Placeholder implementation for writeVarInt64
// This should be replaced with actual implementation from the core package
func writeVarInt64(w *bytes.Buffer, value int64) {
	// TODO: Implement proper varint encoding for int64
}