package encoding

import (
	"bytes"
	"errors"
	"strings"

	"ycs/lib0"
)

// StringEncoder encodes strings with optimized length encoding.
type StringEncoder struct {
	builder       *strings.Builder
	lengthEncoder *UintOptRleEncoder
	disposed      bool
}

// NewStringEncoder creates a new instance of StringEncoder.
func NewStringEncoder() *StringEncoder {
	return &StringEncoder{
		builder:       &strings.Builder{},
		lengthEncoder: NewUintOptRleEncoder(),
	}
}

// Write writes a string value to the encoder.
func (e *StringEncoder) Write(value string) error {
	if e.disposed {
		return NewObjectDisposedException("StringEncoder")
	}

	_, err := e.builder.WriteString(value)
	if err != nil {
		return err
	}
	return e.lengthEncoder.Write(uint32(len(value)))
}

// WriteChars writes a character array to the encoder.
func (e *StringEncoder) WriteChars(value []byte, offset int, count int) error {
	if e.disposed {
		return NewObjectDisposedException("StringEncoder")
	}

	if offset < 0 || count < 0 || offset+count > len(value) {
		return errors.New("invalid offset or count")
	}

	_, err := e.builder.Write(value[offset : offset+count])
	if err != nil {
		return err
	}
	return e.lengthEncoder.Write(uint32(count))
}

// ToArray returns the encoded data as a byte array.
func (e *StringEncoder) ToArray() ([]byte, error) {
	if e.disposed {
		return nil, NewObjectDisposedException("StringEncoder")
	}

	var buf bytes.Buffer

	// Write the string content
	if err := lib0.WriteVarString(&buf, e.builder.String()); err != nil {
		return nil, err
	}

	// Write the length encoding
	lengthData, err := e.lengthEncoder.ToArray()
	if err != nil {
		return nil, err
	}

	if _, err := buf.Write(lengthData); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Dispose releases the resources used by the encoder.
func (e *StringEncoder) Dispose() {
	e.dispose(true)
}

// dispose releases the resources used by the encoder.
func (e *StringEncoder) dispose(disposing bool) {
	if !e.disposed {
		if disposing {
			e.builder.Reset()
			e.lengthEncoder.Dispose()
		}

		e.builder = nil
		e.lengthEncoder = nil
		e.disposed = true
	}
}

// CheckDisposed checks if the encoder has been disposed.
func (e *StringEncoder) CheckDisposed() error {
	if e.disposed {
		return NewObjectDisposedException("StringEncoder")
	}
	return nil
}
