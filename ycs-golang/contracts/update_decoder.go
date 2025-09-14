package contracts

import "io"

// IDSDecoder represents a delete set decoder interface
type IDSDecoder interface {
	GetReader() io.Reader
	ResetDsCurVal()
	ReadDsClock() int64
	ReadDsLength() int64
	Close() error
}

// IUpdateDecoder represents an update decoder interface
type IUpdateDecoder interface {
	IDSDecoder
	ReadLeftID() StructID
	ReadRightID() StructID
	ReadClient() int64
	ReadInfo() byte
	ReadString() string
	ReadParentInfo() bool
	ReadTypeRef() uint32
	ReadLength() int
	ReadAny() interface{}
	ReadBuffer() []byte
	ReadKey() string
	ReadJSON() interface{}
	ReadEmbed() interface{} // Add missing method
}
