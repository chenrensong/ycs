package decoding

// Decoder interface represents a decoder that can read elements from a data stream.
type Decoder[T any] interface {
	// Read reads the next element from the underlying data stream.
	Read() T
	
	// Close closes the decoder and releases any resources.
	Close()
}