// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package lib0

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"unicode/utf8"
)

var (
	ErrIntegerOutOfRange = errors.New("integer out of range")
	ErrEndOfStream       = errors.New("end of stream reached")
	ErrUnknownObjectType = errors.New("unknown object type")
)

// StreamReader is an interface that extends io.Reader with additional methods
type StreamReader interface {
	io.Reader
	io.ByteReader
}

// ReadUint16 reads two bytes as an unsigned integer (little-endian)
func ReadUint16(reader StreamReader) (uint16, error) {
	var b [2]byte
	if _, err := io.ReadFull(reader, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b[:]), nil
}

// ReadUint32 reads four bytes as an unsigned integer (little-endian)
func ReadUint32(reader StreamReader) (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(reader, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

// ReadVarUint reads an unsigned integer (32-bit) with variable length
func ReadVarUint(reader StreamReader) (uint32, error) {
	var num uint32
	var len int

	for {
		r, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		num |= uint32(r&0x7F) << len
		len += 7

		if r < 0x80 {
			return num, nil
		}

		if len > 35 {
			return 0, ErrIntegerOutOfRange
		}
	}
}

// ReadVarInt reads a 32-bit variable length signed integer
func ReadVarInt(reader StreamReader) (int64, int, error) {
	r, err := reader.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	num := uint32(r & 0x3F)
	len := 6
	sign := 1
	if (r & 0x40) > 0 {
		sign = -1
	}

	if (r & 0x80) == 0 {
		return int64(sign) * int64(num), sign, nil
	}

	for {
		r, err = reader.ReadByte()
		if err != nil {
			return 0, 0, err
		}

		num |= uint32(r&0x7F) << len
		len += 7

		if r < 0x80 {
			return int64(sign) * int64(num), sign, nil
		}

		if len > 41 {
			return 0, 0, ErrIntegerOutOfRange
		}
	}
}

// ReadVarString reads a variable length string
func ReadVarString(reader StreamReader) (string, error) {
	length, err := ReadVarUint(reader)
	if err != nil {
		return "", err
	}

	if length == 0 {
		return "", nil
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return "", err
	}

	if !utf8.Valid(data) {
		return "", errors.New("invalid UTF-8 string")
	}

	return string(data), nil
}

// ReadVarUint8Array reads a variable length byte array
func ReadVarUint8Array(reader StreamReader) ([]byte, error) {
	length, err := ReadVarUint(reader)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, err
	}

	return data, nil
}

// ReadByte reads a byte from the stream
func ReadByte(reader StreamReader) (byte, error) {
	b, err := reader.ReadByte()
	if err == io.EOF {
		return 0, ErrEndOfStream
	}
	return b, err
}

// ReadBytes reads a sequence of bytes from the stream
func ReadBytes(reader StreamReader, count int) ([]byte, error) {
	data := make([]byte, count)
	if _, err := io.ReadFull(reader, data); err != nil {
		if err == io.EOF {
			return nil, ErrEndOfStream
		}
		return nil, err
	}
	return data, nil
}

// ReadAny decodes data from the stream
func ReadAny(reader StreamReader) (interface{}, error) {
	typ, err := ReadByte(reader)
	if err != nil {
		return nil, err
	}

	switch typ {
	case 119: // String
		return ReadVarString(reader)
	case 120: // boolean true
		return true, nil
	case 121: // boolean false
		return false, nil
	case 123: // Float64
		data, err := ReadBytes(reader, 8)
		if err != nil {
			return nil, err
		}
		return math.Float64frombits(binary.BigEndian.Uint64(data)), nil
	case 124: // Float32
		data, err := ReadBytes(reader, 4)
		if err != nil {
			return nil, err
		}
		return math.Float32frombits(binary.BigEndian.Uint32(data)), nil
	case 125: // integer
		val, _, err := ReadVarInt(reader)
		return val, err
	case 126, 127: // null or undefined
		return nil, nil
	case 116: // ArrayBuffer
		return ReadVarUint8Array(reader)
	case 117: // Array<object>
		length, err := ReadVarUint(reader)
		if err != nil {
			return nil, err
		}
		arr := make([]interface{}, length)
		for i := range arr {
			val, err := ReadAny(reader)
			if err != nil {
				return nil, err
			}
			arr[i] = val
		}
		return arr, nil
	case 118: // object (map[string]interface{})
		length, err := ReadVarUint(reader)
		if err != nil {
			return nil, err
		}
		obj := make(map[string]interface{}, length)
		for i := 0; i < int(length); i++ {
			key, err := ReadVarString(reader)
			if err != nil {
				return nil, err
			}
			val, err := ReadAny(reader)
			if err != nil {
				return nil, err
			}
			obj[key] = val
		}
		return obj, nil
	default:
		return nil, ErrUnknownObjectType
	}
}
