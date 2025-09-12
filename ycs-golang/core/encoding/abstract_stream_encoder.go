package encoding

import (
	"bytes"
)

// AbstractStreamEncoder represents a base class for stream encoders.
type AbstractStreamEncoder struct {
	buffer   *bytes.Buffer
	disposed bool
}

// NewAbstractStreamEncoder creates a new AbstractStreamEncoder.
func NewAbstractStreamEncoder() *AbstractStreamEncoder {
	return &AbstractStreamEncoder{
		buffer:   &bytes.Buffer{},
		disposed: false,
	}
}

// Buffer returns the underlying buffer.
func (e *AbstractStreamEncoder) Buffer() *bytes.Buffer {
	return e.buffer
}

// Disposed returns whether the encoder has been disposed.
func (e *AbstractStreamEncoder) Disposed() bool {
	return e.disposed
}

// ToArray returns a copy of the encoded contents.
func (e *AbstractStreamEncoder) ToArray() []byte {
	e.Flush()
	return e.buffer.Bytes()
}

// GetBuffer returns the current raw buffer of the encoder.
// This buffer is valid only until the encoder is not closed.
func (e *AbstractStreamEncoder) GetBuffer() ([]byte, int) {
	e.Flush()
	return e.buffer.Bytes(), e.buffer.Len()
}

// Flush flushes any buffered data.
func (e *AbstractStreamEncoder) Flush() {
	// In this simple implementation, we don't have anything to flush
	// but in more complex encoders, this might be needed
	e.CheckDisposed()
}

// Close closes the encoder and releases resources.
func (e *AbstractStreamEncoder) Close() {
	if !e.disposed {
		// In Go, we don't have direct control over buffer disposal like in C#.
		// The garbage collector will handle memory management.
		e.buffer = nil
		e.disposed = true
	}
}

// CheckDisposed checks if the encoder has been disposed and panics if it has.
// This mimics the behavior of the C# version in debug mode.
func (e *AbstractStreamEncoder) CheckDisposed() {
	if e.disposed {
		panic("encoder has been closed")
	}
}