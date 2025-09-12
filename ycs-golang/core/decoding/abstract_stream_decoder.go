package decoding

import (
	"io"
)

// AbstractStreamDecoder represents a base class for stream decoders.
type AbstractStreamDecoder struct {
	stream    io.Reader
	leaveOpen bool
	disposed  bool
}

// NewAbstractStreamDecoder creates a new AbstractStreamDecoder.
func NewAbstractStreamDecoder(input io.Reader, leaveOpen bool) *AbstractStreamDecoder {
	return &AbstractStreamDecoder{
		stream:    input,
		leaveOpen: leaveOpen,
		disposed:  false,
	}
}

// Stream returns the underlying stream.
func (d *AbstractStreamDecoder) Stream() io.Reader {
	return d.stream
}

// Disposed returns whether the decoder has been disposed.
func (d *AbstractStreamDecoder) Disposed() bool {
	return d.disposed
}

// Close closes the decoder and releases resources.
func (d *AbstractStreamDecoder) Close() {
	if !d.disposed {
		// In Go, we don't have direct control over stream disposal like in C#.
		// The caller is responsible for closing the stream if needed.
		d.stream = nil
		d.disposed = true
	}
}

// HasContent checks if there is content remaining in the stream.
// Note: This is a simplified implementation since Go's io.Reader doesn't have a direct equivalent to Position/Length.
func (d *AbstractStreamDecoder) HasContent() bool {
	// This is a placeholder implementation.
	// In practice, you would need to implement this based on your specific use case.
	return !d.disposed
}