package encoding

// Encoder interface represents an encoder that can write elements to a data stream.
type Encoder[T any] interface {
	// Write writes a value to the underlying data stream.
	Write(value T)

	// ToArray returns a copy of the encoded contents.
	ToArray() []byte

	// GetBuffer returns the current raw buffer of the encoder.
	// This buffer is valid only until the encoder is not closed.
	GetBuffer() ([]byte, int)

	// Close closes the encoder and releases any resources.
	Close()
}