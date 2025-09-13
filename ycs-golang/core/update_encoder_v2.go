// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package core

import (
	"io"

	"github.com/chenrensong/ygo/contracts"
	"github.com/chenrensong/ygo/lib0/encoding"
)

// UpdateEncoderV2 implements IUpdateEncoder for version 2
type UpdateEncoderV2 struct {
	restWriter        io.Writer
	keyClockEncoder   contracts.IEncoder
	keyClientEncoder  contracts.IEncoder
	leftClockEncoder  contracts.IEncoder
	rightClockEncoder contracts.IEncoder
	infoEncoder       contracts.IEncoder
	stringEncoder     contracts.IEncoder
	parentInfoEncoder contracts.IEncoder
	typeRefEncoder    contracts.IEncoder
	lengthEncoder     contracts.IEncoder
}

// NewUpdateEncoderV2 creates a new UpdateEncoderV2
func NewUpdateEncoderV2(writer io.Writer) *UpdateEncoderV2 {
	encoder := &UpdateEncoderV2{
		restWriter: writer,
	}

	// Initialize all encoders
	encoder.keyClockEncoder = encoding.NewIntDiffOptRleEncoder(writer)
	encoder.keyClientEncoder = encoding.NewUintOptRleEncoder(writer)
	encoder.leftClockEncoder = encoding.NewIntDiffOptRleEncoder(writer)
	encoder.rightClockEncoder = encoding.NewIntDiffOptRleEncoder(writer)
	encoder.infoEncoder = encoding.NewRleEncoder(writer)
	encoder.stringEncoder = encoding.NewStringEncoder(writer)
	encoder.parentInfoEncoder = encoding.NewRleEncoder(writer)
	encoder.typeRefEncoder = encoding.NewUintOptRleEncoder(writer)
	encoder.lengthEncoder = encoding.NewUintOptRleEncoder(writer)

	return encoder
}

// GetRestWriter returns the underlying writer
func (e *UpdateEncoderV2) GetRestWriter() io.Writer {
	return e.restWriter
}

// WriteLeftID writes a left ID
func (e *UpdateEncoderV2) WriteLeftID(id contracts.StructID) {
	e.keyClientEncoder.Write(uint32(id.Client))
	e.leftClockEncoder.Write(uint32(id.Clock))
}

// WriteRightID writes a right ID
func (e *UpdateEncoderV2) WriteRightID(id contracts.StructID) {
	e.keyClientEncoder.Write(uint32(id.Client))
	e.rightClockEncoder.Write(uint32(id.Clock))
}

// WriteClient writes a client ID
func (e *UpdateEncoderV2) WriteClient(client int64) {
	e.keyClientEncoder.Write(uint32(client))
}

// WriteInfo writes info byte
func (e *UpdateEncoderV2) WriteInfo(info byte) {
	e.infoEncoder.Write(uint32(info))
}

// WriteString writes a string
func (e *UpdateEncoderV2) WriteString(s string) {
	e.stringEncoder.Write(s)
}

// WriteParentInfo writes parent info
func (e *UpdateEncoderV2) WriteParentInfo(hasParentYKey bool) {
	if hasParentYKey {
		e.parentInfoEncoder.Write(1)
	} else {
		e.parentInfoEncoder.Write(0)
	}
}

// WriteTypeRef writes a type reference
func (e *UpdateEncoderV2) WriteTypeRef(typeRef uint32) {
	e.typeRefEncoder.Write(typeRef)
}

// WriteLength writes a length value
func (e *UpdateEncoderV2) WriteLength(length int) {
	e.lengthEncoder.Write(uint32(length))
}

// WriteAny writes any value - placeholder implementation
func (e *UpdateEncoderV2) WriteAny(obj interface{}) {
	// This would need proper implementation based on the actual protocol
}

// WriteBuffer writes a buffer - placeholder implementation
func (e *UpdateEncoderV2) WriteBuffer(buf []byte) {
	// This would need proper implementation based on the actual protocol
}

// WriteJSON writes JSON - placeholder implementation
func (e *UpdateEncoderV2) WriteJSON(obj interface{}) {
	// This would need proper implementation based on the actual protocol
}

// WriteKey writes a key
func (e *UpdateEncoderV2) WriteKey(key string) {
	e.stringEncoder.Write(key)
}

// ToArray converts the encoded data to a byte array - placeholder implementation
func (e *UpdateEncoderV2) ToArray() []byte {
	// This would need proper implementation to collect all encoded data
	return nil
}
