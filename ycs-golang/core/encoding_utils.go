package core

import (
	"bytes"
	"fmt"
	"io"
)

const (
	Bit8  = 0x80
	Bit7  = 0x40
	Bits6 = 0x3F
	Bits7 = 0x7F
	Bits8 = 0xFF
)

// WriteVarUint writes a variable length unsigned integer
// Encodes integers in the range [0, 4294967295] / [0, 0xFFFFFFFF]
func WriteVarUint(w io.Writer, num uint32) error {
	for num > Bits7 {
		if err := writeByte(w, byte(Bit8|(Bits7&num))); err != nil {
			return err
		}
		num >>= 7
	}
	return writeByte(w, byte(Bits7&num))
}

// WriteVarInt writes a variable length integer
// Encodes integers in the range [-2147483648, -2147483647]
// We use the 7th bit for signalling that this is a negative number
func WriteVarInt(w io.Writer, num int32, treatZeroAsNegative bool) error {
	isNegative := (num == 0 && treatZeroAsNegative) || num < 0
	if isNegative {
		num = -num
	}

	// First byte: |continue bit|negative bit|6 value bits|
	firstByte := byte(Bits6 & num)
	if isNegative {
		firstByte |= Bit7
	}
	if num > Bits6 {
		firstByte |= Bit8
	}

	if err := writeByte(w, firstByte); err != nil {
		return err
	}
	num >>= 6

	// Subsequent bytes: |continue bit|7 value bits|
	for num > 0 {
		nextByte := byte(Bits7 & num)
		if num > Bits7 {
			nextByte |= Bit8
		}
		if err := writeByte(w, nextByte); err != nil {
			return err
		}
		num >>= 7
	}

	return nil
}

// WriteVarString writes a variable length string
func WriteVarString(w io.Writer, str string) error {
	data := []byte(str)
	return WriteVarUint8Array(w, data)
}

// WriteVarUint8Array writes a byte array with length prefix
func WriteVarUint8Array(w io.Writer, array []byte) error {
	if err := WriteVarUint(w, uint32(len(array))); err != nil {
		return err
	}
	_, err := w.Write(array)
	return err
}

// ReadVarUint reads an unsigned integer (32-bit) with variable length
func ReadVarUint(r io.Reader) (uint32, error) {
	var num uint32
	var len int

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
			return 0, fmt.Errorf("integer out of range")
		}
	}
}

// ReadVarInt reads a 32-bit variable length signed integer
func ReadVarInt(r io.Reader) (int32, bool, error) {
	b, err := readByte(r)
	if err != nil {
		return 0, false, err
	}

	num := uint32(b & Bits6)
	len := 6
	isNegative := (b & Bit7) > 0

	if (b & Bit8) == 0 {
		// Don't continue reading
		if isNegative {
			return -int32(num), isNegative, nil
		}
		return int32(num), isNegative, nil
	}

	for {
		b, err := readByte(r)
		if err != nil {
			return 0, false, err
		}

		num |= uint32(b&Bits7) << len
		len += 7

		if b < Bit8 {
			if isNegative {
				return -int32(num), isNegative, nil
			}
			return int32(num), isNegative, nil
		}

		if len > 41 {
			return 0, false, fmt.Errorf("integer out of range")
		}
	}
}

// ReadVarString reads a variable length string
func ReadVarString(r io.Reader) (string, error) {
	data, err := ReadVarUint8Array(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadVarUint8Array reads a byte array with length prefix
func ReadVarUint8Array(r io.Reader) ([]byte, error) {
	length, err := ReadVarUint(r)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Helper functions
func writeByte(w io.Writer, b byte) error {
	if bw, ok := w.(*bytes.Buffer); ok {
		return bw.WriteByte(b)
	}
	_, err := w.Write([]byte{b})
	return err
}

func readByte(r io.Reader) (byte, error) {
	if br, ok := r.(*bytes.Buffer); ok {
		return br.ReadByte()
	}
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	return buf[0], err
}
