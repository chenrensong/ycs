package encoding

import (
	"bytes"
	"io"
)

// StreamWriter is an interface that extends io.Writer with additional methods
type StreamWriter interface {
	io.Writer
	io.ByteWriter
}

// AbstractStreamEncoder is an abstract base class for stream encoders.
type AbstractStreamEncoder[T any] struct {
	buffer   *bytes.Buffer
	disposed bool
}

// NewAbstractStreamEncoder creates a new instance of AbstractStreamEncoder.
func NewAbstractStreamEncoder[T any]() *AbstractStreamEncoder[T] {
	return &AbstractStreamEncoder[T]{
		buffer: bytes.NewBuffer(nil),
	}
}

// Write writes a value to the stream (to be implemented by concrete encoders).
func (e *AbstractStreamEncoder[T]) Write(value T) error {
	e.CheckDisposed()
	return nil
}

// ToArray returns the encoded data as a byte array.
func (e *AbstractStreamEncoder[T]) ToArray() ([]byte, error) {
	if err := e.Flush(); err != nil {
		return nil, err
	}
	return e.buffer.Bytes(), nil
}

// GetBuffer returns the underlying buffer and its length.
func (e *AbstractStreamEncoder[T]) GetBuffer() ([]byte, int, error) {
	if err := e.Flush(); err != nil {
		return nil, 0, err
	}
	return e.buffer.Bytes(), e.buffer.Len(), nil
}

// GetWriter returns the underlying buffer as a StreamWriter for derived classes.
func (e *AbstractStreamEncoder[T]) GetWriter() (StreamWriter, error) {
	if err := e.CheckDisposed(); err != nil {
		return nil, err
	}
	return e.buffer, nil
}

// Dispose releases the resources used by the encoder.
func (e *AbstractStreamEncoder[T]) Dispose() {
	e.dispose(true)
}

// Flush ensures all data is written to the buffer.
func (e *AbstractStreamEncoder[T]) Flush() error {
	e.CheckDisposed()
	return nil
}

// dispose releases the resources used by the encoder.
func (e *AbstractStreamEncoder[T]) dispose(disposing bool) {
	if !e.disposed {
		if disposing {
			e.buffer = nil
		}
		e.disposed = true
	}
}

// CheckDisposed checks if the encoder has been disposed.
func (e *AbstractStreamEncoder[T]) CheckDisposed() error {
	if e.disposed {
		return NewObjectDisposedException("AbstractStreamEncoder")
	}
	return nil
}

// ObjectDisposedException represents an exception when accessing a disposed object.
type ObjectDisposedException struct {
	message string
}

func NewObjectDisposedException(typeName string) error {
	return &ObjectDisposedException{
		message: "ObjectDisposedException: " + typeName,
	}
}

func (e *ObjectDisposedException) Error() string {
	return e.message
}
