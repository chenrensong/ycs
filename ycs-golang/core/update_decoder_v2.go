// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"io"

	"github.com/chenrensong/ygo/contracts"
	"github.com/chenrensong/ygo/lib0/decoding"
)

// UpdateDecoderV2 implements IUpdateDecoder for version 2
type UpdateDecoderV2 struct {
	reader            contracts.IDecoder
	keyClockDecoder   contracts.IDecoder
	keyClientDecoder  contracts.IDecoder
	leftClockDecoder  contracts.IDecoder
	rightClockDecoder contracts.IDecoder
	infoDecoder       contracts.IDecoder
	stringDecoder     contracts.IDecoder
	parentInfoDecoder contracts.IDecoder
	typeRefDecoder    contracts.IDecoder
	lengthDecoder     contracts.IDecoder
}

// NewUpdateDecoderV2 creates a new UpdateDecoderV2
func NewUpdateDecoderV2(reader io.Reader) *UpdateDecoderV2 {
	decoder := &UpdateDecoderV2{
		reader: decoding.NewStreamDecoder(reader),
	}

	// Initialize all decoders
	decoder.keyClockDecoder = decoding.NewIntDiffOptRleDecoder(decoder.reader)
	decoder.keyClientDecoder = decoding.NewUintOptRleDecoder(decoder.reader)
	decoder.leftClockDecoder = decoding.NewIntDiffOptRleDecoder(decoder.reader)
	decoder.rightClockDecoder = decoding.NewIntDiffOptRleDecoder(decoder.reader)
	decoder.infoDecoder = decoding.NewRleDecoder(decoder.reader)
	decoder.stringDecoder = decoding.NewStringDecoder(decoder.reader)
	decoder.parentInfoDecoder = decoding.NewRleDecoder(decoder.reader)
	decoder.typeRefDecoder = decoding.NewUintOptRleDecoder(decoder.reader)
	decoder.lengthDecoder = decoding.NewUintOptRleDecoder(decoder.reader)

	return decoder
}

// GetReader returns the underlying reader
func (d *UpdateDecoderV2) GetReader() io.Reader {
	return d.reader
}

// ReadLeftID reads a left ID
func (d *UpdateDecoderV2) ReadLeftID() contracts.StructID {
	client := d.keyClientDecoder.Read()
	clock := d.leftClockDecoder.Read()
	return contracts.StructID{Client: int64(client), Clock: int64(clock)}
}

// ReadRightID reads a right ID
func (d *UpdateDecoderV2) ReadRightID() contracts.StructID {
	client := d.keyClientDecoder.Read()
	clock := d.rightClockDecoder.Read()
	return contracts.StructID{Client: int64(client), Clock: int64(clock)}
}

// ReadClient reads a client ID
func (d *UpdateDecoderV2) ReadClient() int64 {
	return int64(d.keyClientDecoder.Read())
}

// ReadInfo reads info byte
func (d *UpdateDecoderV2) ReadInfo() byte {
	return byte(d.infoDecoder.Read())
}

// ReadString reads a string
func (d *UpdateDecoderV2) ReadString() string {
	return d.stringDecoder.Read()
}

// ReadParentInfo reads parent info
func (d *UpdateDecoderV2) ReadParentInfo() bool {
	return d.parentInfoDecoder.Read() == 1
}

// ReadTypeRef reads a type reference
func (d *UpdateDecoderV2) ReadTypeRef() uint32 {
	return d.typeRefDecoder.Read()
}

// ReadLength reads a length value
func (d *UpdateDecoderV2) ReadLength() int {
	return int(d.lengthDecoder.Read())
}

// ReadAny reads any value - placeholder implementation
func (d *UpdateDecoderV2) ReadAny() interface{} {
	// This would need proper implementation based on the actual protocol
	return nil
}

// ReadBuffer reads a buffer - placeholder implementation
func (d *UpdateDecoderV2) ReadBuffer() []byte {
	// This would need proper implementation based on the actual protocol
	return nil
}

// ReadJSON reads JSON - placeholder implementation
func (d *UpdateDecoderV2) ReadJSON() interface{} {
	// This would need proper implementation based on the actual protocol
	return nil
}

// ReadKey reads a key
func (d *UpdateDecoderV2) ReadKey() string {
	return d.stringDecoder.Read()
}
