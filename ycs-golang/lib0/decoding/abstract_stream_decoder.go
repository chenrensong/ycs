package decoding

import (
	"io"
)

// AbstractStreamDecoder is an abstract base class for stream decoders
type AbstractStreamDecoder[T any] struct {
	stream    io.ReadSeekCloser
	leaveOpen bool
	disposed  bool
}

// NewAbstractStreamDecoder creates a new instance of AbstractStreamDecoder
func NewAbstractStreamDecoder[T any](stream io.ReadSeekCloser, leaveOpen bool) *AbstractStreamDecoder[T] {
	return &AbstractStreamDecoder[T]{
		stream:    stream,
		leaveOpen: leaveOpen,
	}
}

// Stream returns the underlying stream
func (d *AbstractStreamDecoder[T]) Stream() io.ReadSeekCloser {
	return d.stream
}

// Disposed returns true if the decoder has been disposed
func (d *AbstractStreamDecoder[T]) Disposed() bool {
	return d.disposed
}

// HasContent checks if there is more content to read
func (d *AbstractStreamDecoder[T]) HasContent() bool {
	if d.disposed || d.stream == nil {
		return false
	}

	pos, err := d.stream.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}

	size, err := d.stream.Seek(0, io.SeekEnd)
	if err != nil {
		return false
	}

	// Reset position back to where we were
	_, _ = d.stream.Seek(pos, io.SeekStart)
	return pos < size
}

// Read reads the next element from the stream (abstract method)
func (d *AbstractStreamDecoder[T]) Read() (T, error) {
	d.CheckDisposed()
	var zero T
	return zero, nil
}

// Dispose releases the resources used by the decoder
func (d *AbstractStreamDecoder[T]) Dispose() {
	d.dispose(true)
}

// dispose releases the resources used by the decoder
func (d *AbstractStreamDecoder[T]) dispose(disposing bool) {
	if !d.disposed {
		if disposing && !d.leaveOpen && d.stream != nil {
			d.stream.Close()
		}

		d.stream = nil
		d.disposed = true
	}
}

// CheckDisposed checks if the decoder has been disposed and panics if it is
func (d *AbstractStreamDecoder[T]) CheckDisposed() {
	if d.disposed {
		panic("ObjectDisposedException: AbstractStreamDecoder")
	}
}
