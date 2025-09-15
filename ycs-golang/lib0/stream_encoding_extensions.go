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
	"reflect"
)

var (
	ErrUnsupportedType = errors.New("unsupported object type")
)

// StreamWriter is an interface that extends io.Writer with additional methods
type StreamWriter interface {
	io.Writer
	io.ByteWriter
}

// WriteUint16 writes two bytes as an unsigned integer (little-endian)
func WriteUint16(writer StreamWriter, num uint16) error {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], num)
	_, err := writer.Write(b[:])
	return err
}

// WriteUint32 writes four bytes as an unsigned integer (little-endian)
func WriteUint32(writer StreamWriter, num uint32) error {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], num)
	_, err := writer.Write(b[:])
	return err
}

// WriteVarUint writes a variable length unsigned integer
func WriteVarUint(writer io.Writer, num uint32) error {
	// Cast to StreamWriter if possible, otherwise use a wrapper
	if streamWriter, ok := writer.(StreamWriter); ok {
		for num > 0x7F {
			if err := streamWriter.WriteByte(byte(0x80 | (num & 0x7F))); err != nil {
				return err
			}
			num >>= 7
		}
		return streamWriter.WriteByte(byte(num & 0x7F))
	}
	
	// Fallback implementation for regular io.Writer
	buf := make([]byte, 1)
	for num > 0x7F {
		buf[0] = byte(0x80 | (num & 0x7F))
		if _, err := writer.Write(buf); err != nil {
			return err
		}
		num >>= 7
	}
	buf[0] = byte(num & 0x7F)
	_, err := writer.Write(buf)
	return err
}

// WriteVarInt writes a variable length integer
func WriteVarInt(writer StreamWriter, num int64, treatZeroAsNegative *bool) error {
	isNegative := false
	if num != 0 {
		isNegative = num < 0
	} else if treatZeroAsNegative != nil {
		isNegative = *treatZeroAsNegative
	}

	if isNegative {
		num = -num
	}

	// First byte contains sign and 6 bits of value
	firstByte := byte(num & 0x3F)
	if isNegative {
		firstByte |= 0x40
	}
	if num > 0x3F {
		firstByte |= 0x80
	}

	if err := writer.WriteByte(firstByte); err != nil {
		return err
	}
	num >>= 6

	// Subsequent bytes
	for num > 0 {
		b := byte(num & 0x7F)
		if num > 0x7F {
			b |= 0x80
		}
		if err := writer.WriteByte(b); err != nil {
			return err
		}
		num >>= 7
	}

	return nil
}

// WriteVarString writes a variable length string
func WriteVarString(writer StreamWriter, str string) error {
	data := []byte(str)
	if err := WriteVarUint(writer, uint32(len(data))); err != nil {
		return err
	}
	_, err := writer.Write(data)
	return err
}

// WriteVarUint8Array writes a variable length byte array
func WriteVarUint8Array(writer StreamWriter, data []byte) error {
	if err := WriteVarUint(writer, uint32(len(data))); err != nil {
		return err
	}
	_, err := writer.Write(data)
	return err
}

// WriteAny encodes data with efficient binary format
func WriteAny(writer StreamWriter, data interface{}) error {
	if data == nil {
		return writer.WriteByte(126) // null
	}

	switch v := data.(type) {
	case string: // TYPE 119: STRING
		if err := writer.WriteByte(119); err != nil {
			return err
		}
		return WriteVarString(writer, v)

	case bool: // TYPE 120/121: boolean (true/false)
		if v {
			return writer.WriteByte(120)
		}
		return writer.WriteByte(121)

	case float64: // TYPE 123: FLOAT64
		if err := writer.WriteByte(123); err != nil {
			return err
		}
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], math.Float64bits(v))
		_, err := writer.Write(b[:])
		return err

	case float32: // TYPE 124: FLOAT32
		if err := writer.WriteByte(124); err != nil {
			return err
		}
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], math.Float32bits(v))
		_, err := writer.Write(b[:])
		return err

	case int: // TYPE 125: INTEGER
		if err := writer.WriteByte(125); err != nil {
			return err
		}
		return WriteVarInt(writer, int64(v), nil)

	case int64: // Special case: treat LONG as INTEGER
		if err := writer.WriteByte(125); err != nil {
			return err
		}
		return WriteVarInt(writer, v, nil)

	case []byte: // TYPE 116: ArrayBuffer
		if err := writer.WriteByte(116); err != nil {
			return err
		}
		return WriteVarUint8Array(writer, v)

	case map[string]interface{}: // TYPE 118: object (Dictionary<string, object>)
		if err := writer.WriteByte(118); err != nil {
			return err
		}
		if err := WriteVarUint(writer, uint32(len(v))); err != nil {
			return err
		}
		for key, val := range v {
			if err := WriteVarString(writer, key); err != nil {
				return err
			}
			if err := WriteAny(writer, val); err != nil {
				return err
			}
		}
		return nil

	case []interface{}: // TYPE 117: Array
		if err := writer.WriteByte(117); err != nil {
			return err
		}
		if err := WriteVarUint(writer, uint32(len(v))); err != nil {
			return err
		}
		for _, item := range v {
			if err := WriteAny(writer, item); err != nil {
				return err
			}
		}
		return nil

	default:
		// Handle other types via reflection
		val := reflect.ValueOf(data)
		switch val.Kind() {
		case reflect.Slice, reflect.Array:
			if err := writer.WriteByte(117); err != nil {
				return err
			}
			length := val.Len()
			if err := WriteVarUint(writer, uint32(length)); err != nil {
				return err
			}
			for i := 0; i < length; i++ {
				if err := WriteAny(writer, val.Index(i).Interface()); err != nil {
					return err
				}
			}
			return nil

		case reflect.Map:
			if err := writer.WriteByte(118); err != nil {
				return err
			}
			keys := val.MapKeys()
			if err := WriteVarUint(writer, uint32(len(keys))); err != nil {
				return err
			}
			for _, key := range keys {
				strKey, ok := key.Interface().(string)
				if !ok {
					return ErrUnsupportedType
				}
				if err := WriteVarString(writer, strKey); err != nil {
					return err
				}
				if err := WriteAny(writer, val.MapIndex(key).Interface()); err != nil {
					return err
				}
			}
			return nil

		default:
			return ErrUnsupportedType
		}
	}
}


