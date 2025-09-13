package contracts

import (
	"encoding/binary"
	"fmt"
	"io"
)

// StructID represents a unique identifier for a struct
type StructID struct {
	// Client id
	Client int64
	// Unique per client id, continuous number
	Clock int64
}

// NewStructID creates a new StructID
func NewStructID(client, clock int64) StructID {
	if client < 0 {
		panic("Client should not be negative, as it causes client encoder to fail")
	}
	if clock < 0 {
		panic("Clock should not be negative")
	}
	return StructID{Client: client, Clock: clock}
}

// Equals checks if two StructIDs are equal
func (s StructID) Equals(other StructID) bool {
	return s.Client == other.Client && s.Clock == other.Clock
}

// EqualsPtr checks if two StructID pointers are equal (handles nil cases)
func EqualsPtr(a, b *StructID) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return a.Equals(*b)
	}
	return false
}

// Write writes the StructID to a writer using variable-length encoding
func (s StructID) Write(writer io.Writer) error {
	if err := writeVarUint(writer, uint64(s.Client)); err != nil {
		return err
	}
	return writeVarUint(writer, uint64(s.Clock))
}

// ReadStructID reads a StructID from a reader
func ReadStructID(reader io.Reader) (StructID, error) {
	client, err := readVarUint(reader)
	if err != nil {
		return StructID{}, err
	}
	clock, err := readVarUint(reader)
	if err != nil {
		return StructID{}, err
	}
	if client < 0 || clock < 0 {
		return StructID{}, fmt.Errorf("client and clock should be non-negative")
	}
	return StructID{Client: int64(client), Clock: int64(clock)}, nil
}

// String returns string representation of StructID
func (s StructID) String() string {
	return fmt.Sprintf("StructID{Client: %d, Clock: %d}", s.Client, s.Clock)
}

// Helper functions for variable-length encoding (simplified implementation)
func writeVarUint(writer io.Writer, value uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, value)
	_, err := writer.Write(buf[:n])
	return err
}

func readVarUint(reader io.Reader) (uint64, error) {
	return binary.ReadUvarint(&readerWrapper{reader})
}

// readerWrapper implements io.ByteReader for binary.ReadUvarint
type readerWrapper struct {
	io.Reader
}

func (r *readerWrapper) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	_, err := r.Reader.Read(buf)
	return buf[0], err
}
