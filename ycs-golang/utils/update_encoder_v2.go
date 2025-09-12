// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"bytes"
	"ycs-golang/core"
	"ycs-golang/core/encoding"
)

// DSEncoderV2 represents an encoder for delete sets
type DSEncoderV2 struct {
	RestWriter *bytes.Buffer
	dsCurVal   int64
	disposed   bool
}

// NewDSEncoderV2 creates a new DSEncoderV2
func NewDSEncoderV2() *DSEncoderV2 {
	return &DSEncoderV2{
		RestWriter: &bytes.Buffer{},
		dsCurVal:   0,
		disposed:   false,
	}
}

// ResetDsCurVal resets the current delete set value
func (d *DSEncoderV2) ResetDsCurVal() {
	d.dsCurVal = 0
}

// WriteDsClock writes a delete set clock
func (d *DSEncoderV2) WriteDsClock(clock int64) {
	diff := clock - d.dsCurVal
	if diff < 0 {
		panic("diff must be >= 0")
	}
	d.dsCurVal = clock
	core.WriteVarUint(d.RestWriter, uint64(diff))
}

// WriteDsLength writes a delete set length
func (d *DSEncoderV2) WriteDsLength(length int64) {
	if length <= 0 {
		panic("length must be > 0")
	}
	
	core.WriteVarUint(d.RestWriter, uint64(length-1))
	d.dsCurVal += length
}

// ToArray returns the encoded data as a byte array
func (d *DSEncoderV2) ToArray() []byte {
	return d.RestWriter.Bytes()
}

// Dispose disposes the encoder
func (d *DSEncoderV2) Dispose() {
	if !d.disposed {
		// In Go, we don't explicitly dispose of the buffer
		// The garbage collector will handle it
		d.RestWriter = nil
		d.disposed = true
	}
}

// UpdateEncoderV2 represents an update encoder V2
type UpdateEncoderV2 struct {
	*DSEncoderV2
	keyClock         int
	keyMap           map[string]int
	keyClockEncoder  *encoding.IntDiffOptRleEncoder
	clientEncoder    *encoding.UintOptRleEncoder
	leftClockEncoder *encoding.IntDiffOptRleEncoder
	rightClockEncoder *encoding.IntDiffOptRleEncoder
	infoEncoder      *encoding.RleEncoder
	stringEncoder    *encoding.StringEncoder
	parentInfoEncoder *encoding.RleEncoder
	typeRefEncoder   *encoding.UintOptRleEncoder
	lengthEncoder    *encoding.UintOptRleEncoder
}

// NewUpdateEncoderV2 creates a new UpdateEncoderV2
func NewUpdateEncoderV2() *UpdateEncoderV2 {
	return &UpdateEncoderV2{
		DSEncoderV2:       NewDSEncoderV2(),
		keyClock:          0,
		keyMap:            make(map[string]int),
		keyClockEncoder:   encoding.NewIntDiffOptRleEncoder(),
		clientEncoder:     encoding.NewUintOptRleEncoder(),
		leftClockEncoder:  encoding.NewIntDiffOptRleEncoder(),
		rightClockEncoder: encoding.NewIntDiffOptRleEncoder(),
		infoEncoder:       encoding.NewRleEncoder(),
		stringEncoder:     encoding.NewStringEncoder(),
		parentInfoEncoder: encoding.NewRleEncoder(),
		typeRefEncoder:    encoding.NewUintOptRleEncoder(),
		lengthEncoder:     encoding.NewUintOptRleEncoder(),
	}
}

// ToArray returns the encoded data as a byte array
func (u *UpdateEncoderV2) ToArray() []byte {
	var stream bytes.Buffer
	
	// Read the feature flag that might be used in the future.
	stream.WriteByte(0)
	
	// TODO: [alekseyk] Maybe pass the writer directly instead of using ToArray()?
	core.WriteVarUint8Array(&stream, u.keyClockEncoder.ToArray())
	core.WriteVarUint8Array(&stream, u.clientEncoder.ToArray())
	core.WriteVarUint8Array(&stream, u.leftClockEncoder.ToArray())
	core.WriteVarUint8Array(&stream, u.rightClockEncoder.ToArray())
	core.WriteVarUint8Array(&stream, u.infoEncoder.ToArray())
	core.WriteVarUint8Array(&stream, u.stringEncoder.ToArray())
	core.WriteVarUint8Array(&stream, u.parentInfoEncoder.ToArray())
	core.WriteVarUint8Array(&stream, u.typeRefEncoder.ToArray())
	core.WriteVarUint8Array(&stream, u.lengthEncoder.ToArray())
	
	// Append the rest of the data from the RestWriter.
	// Note it's not the 'WriteVarUint8Array'.
	content := u.DSEncoderV2.ToArray()
	stream.Write(content)
	
	return stream.Bytes()
}

// WriteLeftId writes the left ID
func (u *UpdateEncoderV2) WriteLeftId(id ID) {
	u.clientEncoder.Write(uint64(id.Client))
	u.leftClockEncoder.Write(id.Clock)
}

// WriteRightId writes the right ID
func (u *UpdateEncoderV2) WriteRightId(id ID) {
	u.clientEncoder.Write(uint64(id.Client))
	u.rightClockEncoder.Write(id.Clock)
}

// WriteClient writes a client
func (u *UpdateEncoderV2) WriteClient(client int64) {
	u.clientEncoder.Write(uint64(client))
}

// WriteInfo writes info
func (u *UpdateEncoderV2) WriteInfo(info byte) {
	u.infoEncoder.Write(info)
}

// WriteString writes a string
func (u *UpdateEncoderV2) WriteString(s string) {
	u.stringEncoder.Write(s)
}

// WriteParentInfo writes parent info
func (u *UpdateEncoderV2) WriteParentInfo(isYKey bool) {
	if isYKey {
		u.parentInfoEncoder.Write(1)
	} else {
		u.parentInfoEncoder.Write(0)
	}
}

// WriteTypeRef writes a type reference
func (u *UpdateEncoderV2) WriteTypeRef(info uint32) {
	u.typeRefEncoder.Write(info)
}

// WriteLength writes a length
func (u *UpdateEncoderV2) WriteLength(len int) {
	if len < 0 {
		panic("len must be >= 0")
	}
	u.lengthEncoder.Write(uint64(len))
}

// WriteAny writes any value
func (u *UpdateEncoderV2) WriteAny(any interface{}) {
	// This would require implementing WriteAny in the core package
	// For now, we'll leave it as a placeholder
}

// WriteBuffer writes a buffer
func (u *UpdateEncoderV2) WriteBuffer(data []byte) {
	core.WriteVarUint8Array(u.RestWriter, data)
}

// WriteKey writes a key
func (u *UpdateEncoderV2) WriteKey(key string) {
	u.keyClockEncoder.Write(int64(u.keyClock))
	u.keyClock++
	
	if _, exists := u.keyMap[key]; !exists {
		u.stringEncoder.Write(key)
	}
}

// WriteJson writes JSON
func (u *UpdateEncoderV2) WriteJson(any interface{}) {
	// This would require implementing JSON serialization
	// For now, we'll leave it as a placeholder
}

// Dispose disposes the update encoder
func (u *UpdateEncoderV2) Dispose() {
	if !u.disposed {
		// Clear the key map
		for k := range u.keyMap {
			delete(u.keyMap, k)
		}
		
		u.keyClockEncoder.Dispose()
		u.clientEncoder.Dispose()
		u.leftClockEncoder.Dispose()
		u.rightClockEncoder.Dispose()
		u.infoEncoder.Dispose()
		u.stringEncoder.Dispose()
		u.parentInfoEncoder.Dispose()
		u.typeRefEncoder.Dispose()
		u.lengthEncoder.Dispose()
		
		u.keyMap = nil
		u.keyClockEncoder = nil
		u.clientEncoder = nil
		u.leftClockEncoder = nil
		u.rightClockEncoder = nil
		u.infoEncoder = nil
		u.stringEncoder = nil
		u.parentInfoEncoder = nil
		u.typeRefEncoder = nil
		u.lengthEncoder = nil
	}
	
	u.DSEncoderV2.Dispose()
}