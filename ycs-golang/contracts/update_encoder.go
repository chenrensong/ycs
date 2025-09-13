package contracts

import "io"

// IDSEncoder represents a delete set encoder interface
type IDSEncoder interface {
	GetRestWriter() io.Writer
	ToArray() []byte
	ResetDsCurVal()
	WriteDsClock(clock int64)
	WriteDsLength(length int64)
	Close() error
}

// IUpdateEncoder represents an update encoder interface
type IUpdateEncoder interface {
	IDSEncoder
	WriteLeftID(id StructID)
	WriteRightID(id StructID)
	WriteClient(client int64)
	WriteInfo(info byte)
	WriteString(s string)
	WriteParentInfo(isYKey bool)
	WriteTypeRef(info uint32)
	WriteLength(length int)
	WriteAny(any interface{})
	WriteBuffer(buf []byte)
	WriteKey(key string)
	WriteJSON(any interface{})
}
