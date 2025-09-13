package structs

import (
	"io"
)

// IDSEncoder is an interface that extends io.Closer (for Dispose functionality)
// and provides methods for encoding data structures.
type IDSEncoder interface {
	io.Closer
	RestWriter() io.Writer
	ToArray() []byte

	// ResetDsCurVal resets the ds value to 0.
	// The v2 encoder uses this information to reset the initial diff value.
	ResetDsCurVal()

	WriteDsClock(clock int64)
	WriteDsLength(length int64)
}

// IUpdateEncoder extends IDSEncoder with additional methods for update encoding.
type IUpdateEncoder interface {
	IDSEncoder

	WriteLeftId(id ID)
	WriteRightId(id ID)

	// WriteClient writes client ID.
	// NOTE: Use 'WriteClient' and 'WriteClock' instead of WriteID if possible.
	WriteClient(client int64)

	WriteInfo(info byte)
	WriteString(s string)
	WriteParentInfo(isYKey bool)
	WriteTypeRef(info uint32)

	// WriteLength writes length of a struct - well suited for Opt RLE encoder.
	WriteLength(len int)

	WriteAny(any interface{})
	WriteBuffer(buf []byte)
	WriteKey(key string)
	WriteJson(any interface{}) // Note: Go doesn't support generics like C# yet
}
