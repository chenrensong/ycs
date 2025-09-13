// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package decoding

import (
	"io"

	"github.com/chenrensong/ygo/lib0"
)

// StringDecoder decodes strings from a stream using length prefixes.
type StringDecoder struct {
	lengthDecoder *UintOptRleDecoder
	value         string
	pos           int
	disposed      bool
}

// NewStringDecoder creates a new instance of StringDecoder.
func NewStringDecoder(input io.ReadSeekCloser, leaveOpen bool) *StringDecoder {
	value, err := lib0.ReadVarString(input.(lib0.StreamReader))
	if err != nil {
		panic(err) // Similar to Debug.Assert behavior
	}

	return &StringDecoder{
		value:         value,
		lengthDecoder: NewUintOptRleDecoder(input, leaveOpen),
	}
}

// Read reads the next string from the stream.
func (d *StringDecoder) Read() (string, error) {
	if err := d.CheckDisposed(); err != nil {
		return "", err
	}

	length, err := d.lengthDecoder.Read()
	if err != nil {
		return "", err
	}

	if length == 0 {
		return "", nil
	}

	endPos := d.pos + int(length)
	if endPos > len(d.value) {
		return "", io.ErrUnexpectedEOF
	}

	result := d.value[d.pos:endPos]
	d.pos = endPos

	// No need to keep the string in memory anymore
	if d.pos >= len(d.value) {
		d.value = ""
	}

	return result, nil
}

// Dispose releases the resources used by the decoder.
func (d *StringDecoder) Dispose() {
	d.dispose(true)
}

// dispose releases the resources used by the decoder.
func (d *StringDecoder) dispose(disposing bool) {
	if !d.disposed {
		if disposing && d.lengthDecoder != nil {
			d.lengthDecoder.Dispose()
		}

		d.lengthDecoder = nil
		d.disposed = true
	}
}

// CheckDisposed checks if the decoder has been disposed.
func (d *StringDecoder) CheckDisposed() error {
	if d.disposed {
		return NewObjectDisposedException("StringDecoder")
	}
	return nil
}

// ObjectDisposedException represents an exception when accessing a disposed object.
type ObjectDisposedException struct {
	message string
}

func NewObjectDisposedException(typeName string) *ObjectDisposedException {
	return &ObjectDisposedException{
		message: "ObjectDisposedException: " + typeName,
	}
}

func (e *ObjectDisposedException) Error() string {
	return e.message
}
