package utils

import (
	"io"
)

// IDSDecoder interface represents a decoder for DS (Delete Set)
type IDSDecoder interface {
	Reader() io.Reader
	ResetDsCurVal()
	ReadDsClock() int64
	ReadDsLength() int64
	Close()
}

// IUpdateDecoder interface represents an update decoder
type IUpdateDecoder interface {
	IDSDecoder
	ReadLeftId() *ID
	ReadRightId() *ID
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