package encoding

import (
	"bytes"
	"io"
)

// StringEncoder represents an optimized string encoder.
type StringEncoder struct {
	buffer        *bytes.Buffer
	lengthEncoder *UintOptRleEncoder
	disposed      bool
}

// NewStringEncoder creates a new StringEncoder.
func NewStringEncoder() *StringEncoder {
	return &StringEncoder{
		buffer:        &bytes.Buffer{},
		lengthEncoder: NewUintOptRleEncoder(),
		disposed:      false,
	}
}

// Write writes a string value to the underlying data stream.
func (e *StringEncoder) Write(value string) {
	e.checkDisposed()
	
	// Write the string to the buffer
	e.buffer.WriteString(value)
	
	// Encode the length using the UintOptRleEncoder
	e.lengthEncoder.Write(uint32(len(value)))
}

// WriteBytes writes a byte slice to the underlying data stream.
func (e *StringEncoder) WriteBytes(value []byte) {
	e.checkDisposed()
	
	// Write the bytes to the buffer
	e.buffer.Write(value)
	
	// Encode the length using the UintOptRleEncoder
	e.lengthEncoder.Write(uint32(len(value)))
}

// ToArray returns a copy of the encoded contents.
func (e *StringEncoder) ToArray() []byte {
	e.checkDisposed()
	
	// Create a new buffer to hold the encoded data
	var result bytes.Buffer
	
	// Write the string data as a variable length string
	writeVarString(&result, e.buffer.String())
	
	// Get the length encoder buffer
	lengthBuffer, length := e.lengthEncoder.GetBuffer()
	
	// Write the length buffer
	result.Write(lengthBuffer[:length])
	
	return result.Bytes()
}

// GetBuffer returns the current raw buffer of the encoder.
// This buffer is valid only until the encoder is not closed.
func (e *StringEncoder) GetBuffer() ([]byte, int) {
	// StringEncoder doesn't use temporary byte buffers like the C# version
	panic("StringEncoder doesn't use temporary byte buffers")
}

// Close closes the encoder and releases resources.
func (e *StringEncoder) Close() {
	if !e.disposed {
		// Clear the buffer
		e.buffer.Reset()
		
		// Close the length encoder
		e.lengthEncoder.Close()
		
		e.buffer = nil
		e.lengthEncoder = nil
		e.disposed = true
	}
}

// checkDisposed checks if the encoder has been disposed and panics if it has.
func (e *StringEncoder) checkDisposed() {
	if e.disposed {
		panic("encoder has been closed")
	}
}

// Helper function to write a variable length string
func writeVarString(w io.Writer, str string) {
	// TODO: Implement proper varstring encoding
	// This should write the string length followed by the string data
}