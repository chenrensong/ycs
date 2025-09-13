package structs

import "io"

// IDSDecoder is an interface that extends io.Closer (for Dispose functionality)
// and provides methods for decoding data structures.
type IDSDecoder interface {
	io.Closer
	Reader() io.Reader
	ResetDsCurVal()
	ReadDsClock() int64
	ReadDsLength() int64
}

// IUpdateDecoder extends IDSDecoder with additional methods for update decoding.
type IUpdateDecoder interface {
	IDSDecoder
	ReadLeftId() ID
	ReadRightId() ID
	ReadClient() int64
	ReadInfo() byte
	ReadString() string
	ReadParentInfo() bool
	ReadTypeRef() uint32
	ReadLength() int
	ReadAny() interface{}
	ReadBuffer() []byte
	ReadKey() string
	ReadJson() interface{}
}
