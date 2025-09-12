package decoding

import (
	"io"
)

// StringDecoder represents a decoder for strings.
type StringDecoder struct {
	lengthDecoder *UintOptRleDecoder
	value         string
	pos           int
	disposed      bool
}

// NewStringDecoder creates a new StringDecoder.
func NewStringDecoder(input io.Reader, leaveOpen bool) *StringDecoder {
	// Read variable length string
	value, err := readVarString(input)
	if err != nil {
		panic(err)
	}

	// Create a UintOptRleDecoder for length decoding
	lengthDecoder := NewUintOptRleDecoder(input, leaveOpen)

	return &StringDecoder{
		lengthDecoder: lengthDecoder,
		value:         value,
		pos:           0,
		disposed:      false,
	}
}

// Read reads the next string from the underlying data stream.
func (d *StringDecoder) Read() string {
	d.checkDisposed()

	// Read the length using the UintOptRleDecoder
	length := int(d.lengthDecoder.Read())
	if length == 0 {
		return ""
	}

	// Extract substring
	result := d.value[d.pos : d.pos+length]
	d.pos += length

	// No need to keep the string in memory anymore.
	// This also covers the case when nothing but empty strings are left.
	if d.pos >= len(d.value) {
		d.value = ""
	}

	return result
}

// Close closes the decoder and releases resources.
func (d *StringDecoder) Close() {
	if !d.disposed {
		if d.lengthDecoder != nil {
			// In Go, we don't have a direct equivalent to Dispose,
			// but we can set the reference to nil to allow garbage collection
			d.lengthDecoder = nil
		}
		d.disposed = true
	}
}

// checkDisposed checks if the decoder has been disposed and panics if it has.
func (d *StringDecoder) checkDisposed() {
	if d.disposed {
		panic("decoder has been closed")
	}
}

// Placeholder implementation for readVarString
// This should be replaced with actual implementation from the core package
func readVarString(r io.Reader) (string, error) {
	// TODO: Implement proper varstring decoding
	return "", nil
}