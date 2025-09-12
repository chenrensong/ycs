package core

import (
	"encoding/binary"
	"errors"
	"io"
)

// StreamDecodingExtensions contains extensions compatible with the lib0 library
// for decoding data from streams.

// ReadUint16 reads two bytes as an unsigned integer.
func ReadUint16(r io.Reader) (uint16, error) {
	var buf [2]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf[:]), nil
}

// ReadUint32 reads four bytes as an unsigned integer.
func ReadUint32(r io.Reader) (uint32, error) {
	var buf [4]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

// ReadVarUint reads unsigned integer (32-bit) with variable length.
// 1/8th of the storage is used as encoding overhead.
// Values < 2^7 are stored in one byte.
// Values < 2^14 are stored in two bytes.
func ReadVarUint(r io.Reader) (uint32, error) {
	var num uint32 = 0
	var len int = 0

	for {
		b, err := readByte(r)
		if err != nil {
			return 0, err
		}

		num |= uint32(b&Bits7) << len
		len += 7

		if b < Bit8 {
			return num, nil
		}

		if len > 35 {
			return 0, errors.New("integer out of range")
		}
	}
}

// ReadVarInt reads a 32-bit variable length signed integer.
// 1/8th of storage is used as encoding overhead.
// Values < 2^7 are stored in one byte.
// Values < 2^14 are stored in two bytes.
func ReadVarInt(r io.Reader) (int32, error) {
	b, err := readByte(r)
	if err != nil {
		return 0, err
	}

	var num uint32 = uint32(b & Bits6)
	var len int = 6
	var sign int32 = 1
	if (b & Bit7) > 0 {
		sign = -1
	}

	if (b & Bit8) == 0 {
		// Don't continue reading.
		return sign * int32(num), nil
	}

	for {
		b, err := readByte(r)
		if err != nil {
			return 0, err
		}

		num |= uint32(b&Bits7) << len
		len += 7

		if b < Bit8 {
			return sign * int32(num), nil
		}

		if len > 41 {
			return 0, errors.New("integer out of range")
		}
	}
}

// ReadVarString reads a variable length string.
// ReadVarUint is used to store the length of the string.
func ReadVarString(r io.Reader) (string, error) {
	length, err := ReadVarUint(r)
	if err != nil {
		return "", err
	}

	if length == 0 {
		return "", nil
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// ReadVarUint8Array reads a variable length byte array.
func ReadVarUint8Array(r io.Reader) ([]byte, error) {
	length, err := ReadVarUint(r)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Helper function to read a single byte
func readByte(r io.Reader) (byte, error) {
	var buf [1]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}