package core

import (
	"encoding/binary"
	"io"
	"math"
)

// StreamEncodingExtensions contains extensions compatible with the lib0 library
// for encoding data to streams.

// WriteUint16 writes two bytes as an unsigned integer.
func WriteUint16(w io.Writer, num uint16) error {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], num)
	_, err := w.Write(buf[:])
	return err
}

// WriteUint32 writes four bytes as an unsigned integer.
func WriteUint32(w io.Writer, num uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], num)
	_, err := w.Write(buf[:])
	return err
}

// WriteVarUint writes a variable length unsigned integer.
// Encodes integers in the range [0, 4294967295] / [0, 0xFFFFFFFF].
func WriteVarUint(w io.Writer, num uint32) error {
	for num > Bits7 {
		err := writeByte(w, byte(Bit8|(Bits7&num)))
		if err != nil {
			return err
		}
		num >>= 7
	}

	return writeByte(w, byte(Bits7&num))
}

// WriteVarInt writes a variable length integer.
// Encodes integers in the range [-2147483648, -2147483647].
// We don't use zig-zag encoding because we want to keep the option open
// to use the same function for BigInt and 53-bit integers (doubles).
// We use the 7-th bit instead for signalling that this is a negative number.
func WriteVarInt(w io.Writer, num int32) error {
	isNegative := num < 0
	if isNegative {
		num = -num
	}

	//                      |   whether to continue reading   |         is negative         | value.
	err := writeByte(w, byte((func() uint32 {
		if num > int32(Bits6) {
			return Bit8
		}
		return 0
	}())|(func() uint32 {
		if isNegative {
			return Bit7
		}
		return 0
	}())|(Bits6&uint32(num))))
	if err != nil {
		return err
	}

	num >>= 6

	// We don't need to consider the case of num == 0 so we can use a different pattern here than above.
	for num > 0 {
		err := writeByte(w, byte((func() uint32 {
			if num > int32(Bits7) {
				return Bit8
			}
			return 0
		}())|(Bits7&uint32(num))))
		if err != nil {
			return err
		}
		num >>= 7
	}

	return nil
}

// WriteVarString writes a variable length string.
func WriteVarString(w io.Writer, str string) error {
	data := []byte(str)
	return WriteVarUint8Array(w, data)
}

// WriteVarUint8Array appends a byte array to the stream.
func WriteVarUint8Array(w io.Writer, array []byte) error {
	err := WriteVarUint(w, uint32(len(array)))
	if err != nil {
		return err
	}

	_, err = w.Write(array)
	return err
}

// Helper function to write a single byte
func writeByte(w io.Writer, b byte) error {
	_, err := w.Write([]byte{b})
	return err
}

// WriteAny encodes data with efficient binary format.
func WriteAny(w io.Writer, o interface{}) error {
	switch v := o.(type) {
	case string: // TYPE 119: STRING
		err := writeByte(w, 119)
		if err != nil {
			return err
		}
		return WriteVarString(w, v)
	case bool: // TYPE 120/121: boolean (true/false)
		var b byte
		if v {
			b = 120
		} else {
			b = 121
		}
		return writeByte(w, b)
	case float64: // TYPE 123: FLOAT64
		err := writeByte(w, 123)
		if err != nil {
			return err
		}
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(v))
		_, err = w.Write(buf[:])
		return err
	case float32: // TYPE 124: FLOAT32
		err := writeByte(w, 124)
		if err != nil {
			return err
		}
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], math.Float32bits(v))
		_, err = w.Write(buf[:])
		return err
	case int32: // TYPE 125: INTEGER
		err := writeByte(w, 125)
		if err != nil {
			return err
		}
		return WriteVarInt(w, v)
	case int: // Special case: treat int as int32
		err := writeByte(w, 125)
		if err != nil {
			return err
		}
		return WriteVarInt(w, int32(v))
	case nil: // TYPE 126: null
		// TYPE 127: undefined
		return writeByte(w, 126)
	case []byte: // TYPE 116: ArrayBuffer
		err := writeByte(w, 116)
		if err != nil {
			return err
		}
		return WriteVarUint8Array(w, v)
	case map[string]interface{}: // TYPE 118: object (Dictionary<string, object>)
		err := writeByte(w, 118)
		if err != nil {
			return err
		}
		err = WriteVarUint(w, uint32(len(v)))
		if err != nil {
			return err
		}
		for key, value := range v {
			err = WriteVarString(w, key)
			if err != nil {
				return err
			}
			err = WriteAny(w, value)
			if err != nil {
				return err
			}
		}
		return nil
	case []interface{}: // TYPE 117: Array
		err := writeByte(w, 117)
		if err != nil {
			return err
		}
		err = WriteVarUint(w, uint32(len(v)))
		if err != nil {
			return err
		}
		for _, item := range v {
			err = WriteAny(w, item)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		// For unsupported types, write as undefined
		return writeByte(w, 127)
	}
}