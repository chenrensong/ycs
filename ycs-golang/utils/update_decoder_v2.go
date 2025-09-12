// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package utils

import (
	"bytes"
	"io"
	"ycs-golang/core"
	"ycs-golang/core/decoding"
)

// DSDecoderV2 represents a decoder for delete sets
type DSDecoderV2 struct {
	Reader    io.Reader
	leaveOpen bool
	dsCurVal  int64
	disposed  bool
}

// NewDSDecoderV2 creates a new DSDecoderV2
func NewDSDecoderV2(input io.Reader) *DSDecoderV2 {
	return &DSDecoderV2{
		Reader:    input,
		leaveOpen: false,
		dsCurVal:  0,
		disposed:  false,
	}
}

// ResetDsCurVal resets the current delete set value
func (d *DSDecoderV2) ResetDsCurVal() {
	d.dsCurVal = 0
}

// ReadDsClock reads a delete set clock
func (d *DSDecoderV2) ReadDsClock() int64 {
	val := core.ReadVarUint(d.Reader)
	d.dsCurVal += int64(val)
	return d.dsCurVal
}

// ReadDsLength reads a delete set length
func (d *DSDecoderV2) ReadDsLength() int64 {
	diff := core.ReadVarUint(d.Reader) + 1
	d.dsCurVal += int64(diff)
	return int64(diff)
}

// Dispose disposes the decoder
func (d *DSDecoderV2) Dispose() {
	if !d.disposed {
		if !d.leaveOpen {
			// In Go, we don't explicitly dispose of the reader
			// The garbage collector will handle it
		}
		d.Reader = nil
		d.disposed = true
	}
}

// UpdateDecoderV2 represents an update decoder V2
type UpdateDecoderV2 struct {
	*DSDecoderV2
	keys              []string
	keyClockDecoder   *decoding.IntDiffOptRleDecoder
	clientDecoder     *decoding.UintOptRleDecoder
	leftClockDecoder  *decoding.IntDiffOptRleDecoder
	rightClockDecoder *decoding.IntDiffOptRleDecoder
	infoDecoder       *decoding.RleDecoder
	stringDecoder     *decoding.StringDecoder
	parentInfoDecoder *decoding.RleDecoder
	typeRefDecoder    *decoding.UintOptRleDecoder
	lengthDecoder     *decoding.UintOptRleDecoder
}

// NewUpdateDecoderV2 creates a new UpdateDecoderV2
func NewUpdateDecoderV2(input io.Reader) *UpdateDecoderV2 {
	ud := &UpdateDecoderV2{
		DSDecoderV2: NewDSDecoderV2(input),
		keys:        make([]string, 0),
	}
	
	// Read feature flag - currently unused
	io.ReadFull(input, make([]byte, 1))
	
	// Read the various decoders
	keyClockData := core.ReadVarUint8ArrayAsStream(input)
	ud.keyClockDecoder = decoding.NewIntDiffOptRleDecoder(bytes.NewReader(keyClockData))
	
	clientData := core.ReadVarUint8ArrayAsStream(input)
	ud.clientDecoder = decoding.NewUintOptRleDecoder(bytes.NewReader(clientData))
	
	leftClockData := core.ReadVarUint8ArrayAsStream(input)
	ud.leftClockDecoder = decoding.NewIntDiffOptRleDecoder(bytes.NewReader(leftClockData))
	
	rightClockData := core.ReadVarUint8ArrayAsStream(input)
	ud.rightClockDecoder = decoding.NewIntDiffOptRleDecoder(bytes.NewReader(rightClockData))
	
	infoData := core.ReadVarUint8ArrayAsStream(input)
	ud.infoDecoder = decoding.NewRleDecoder(bytes.NewReader(infoData))
	
	stringData := core.ReadVarUint8ArrayAsStream(input)
	ud.stringDecoder = decoding.NewStringDecoder(bytes.NewReader(stringData))
	
	parentInfoData := core.ReadVarUint8ArrayAsStream(input)
	ud.parentInfoDecoder = decoding.NewRleDecoder(bytes.NewReader(parentInfoData))
	
	typeRefData := core.ReadVarUint8ArrayAsStream(input)
	ud.typeRefDecoder = decoding.NewUintOptRleDecoder(bytes.NewReader(typeRefData))
	
	lengthData := core.ReadVarUint8ArrayAsStream(input)
	ud.lengthDecoder = decoding.NewUintOptRleDecoder(bytes.NewReader(lengthData))
	
	return ud
}

// ReadLeftId reads the left ID
func (u *UpdateDecoderV2) ReadLeftId() ID {
	if u.disposed {
		panic("Object disposed")
	}
	return ID{
		Client: u.clientDecoder.Read(),
		Clock:  u.leftClockDecoder.Read(),
	}
}

// ReadRightId reads the right ID
func (u *UpdateDecoderV2) ReadRightId() ID {
	if u.disposed {
		panic("Object disposed")
	}
	return ID{
		Client: u.clientDecoder.Read(),
		Clock:  u.rightClockDecoder.Read(),
	}
}

// ReadClient reads the next client ID
func (u *UpdateDecoderV2) ReadClient() int64 {
	if u.disposed {
		panic("Object disposed")
	}
	return u.clientDecoder.Read()
}

// ReadInfo reads info
func (u *UpdateDecoderV2) ReadInfo() byte {
	if u.disposed {
		panic("Object disposed")
	}
	return u.infoDecoder.Read()
}

// ReadString reads a string
func (u *UpdateDecoderV2) ReadString() string {
	if u.disposed {
		panic("Object disposed")
	}
	return u.stringDecoder.Read()
}

// ReadParentInfo reads parent info
func (u *UpdateDecoderV2) ReadParentInfo() bool {
	if u.disposed {
		panic("Object disposed")
	}
	return u.parentInfoDecoder.Read() == 1
}

// ReadTypeRef reads a type reference
func (u *UpdateDecoderV2) ReadTypeRef() uint32 {
	if u.disposed {
		panic("Object disposed")
	}
	return u.typeRefDecoder.Read()
}

// ReadLength reads a length
func (u *UpdateDecoderV2) ReadLength() int {
	if u.disposed {
		panic("Object disposed")
	}
	value := int(u.lengthDecoder.Read())
	return value
}

// ReadAny reads any value
func (u *UpdateDecoderV2) ReadAny() interface{} {
	if u.disposed {
		panic("Object disposed")
	}
	// This would require implementing ReadAny in the core package
	// For now, we'll return nil as a placeholder
	return nil
}

// ReadBuffer reads a buffer
func (u *UpdateDecoderV2) ReadBuffer() []byte {
	if u.disposed {
		panic("Object disposed")
	}
	return core.ReadVarUint8Array(u.Reader)
}

// ReadKey reads a key
func (u *UpdateDecoderV2) ReadKey() string {
	if u.disposed {
		panic("Object disposed")
	}
	
	keyClock := int(u.keyClockDecoder.Read())
	if keyClock < len(u.keys) {
		return u.keys[keyClock]
	} else {
		key := u.stringDecoder.Read()
		u.keys = append(u.keys, key)
		return key
	}
}

// ReadJson reads JSON
func (u *UpdateDecoderV2) ReadJson() interface{} {
	if u.disposed {
		panic("Object disposed")
	}
	
	// This would require implementing JSON deserialization
	// For now, we'll return nil as a placeholder
	return nil
}

// Dispose disposes the update decoder
func (u *UpdateDecoderV2) Dispose() {
	if !u.disposed {
		u.keyClockDecoder.Dispose()
		u.clientDecoder.Dispose()
		u.leftClockDecoder.Dispose()
		u.rightClockDecoder.Dispose()
		u.infoDecoder.Dispose()
		u.stringDecoder.Dispose()
		u.parentInfoDecoder.Dispose()
		u.typeRefDecoder.Dispose()
		u.lengthDecoder.Dispose()
		
		u.keyClockDecoder = nil
		u.clientDecoder = nil
		u.leftClockDecoder = nil
		u.rightClockDecoder = nil
		u.infoDecoder = nil
		u.stringDecoder = nil
		u.parentInfoDecoder = nil
		u.typeRefDecoder = nil
		u.lengthDecoder = nil
	}
	
	u.DSDecoderV2.Dispose()
}