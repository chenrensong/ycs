package utils

import (
	"bytes"
	"encoding/binary"
)

// ID represents a client ID and clock
type ID struct {
	Client int64
	Clock  int64
}

// NewID creates a new ID
func NewID(client, clock int64) *ID {
	// In Go, we don't have Debug.Assert, but we could add checks if needed:
	// if client < 0 {
	//     panic("Client should not be negative")
	// }
	// if clock < 0 {
	//     panic("Clock should not be negative")
	// }
	
	return &ID{
		Client: client,
		Clock:  clock,
	}
}

// Equals checks if two IDs are equal
func (id *ID) Equals(other *ID) bool {
	if id == nil && other == nil {
		return true
	}
	if id == nil || other == nil {
		return false
	}
	return id.Client == other.Client && id.Clock == other.Clock
}

// Write writes the ID to a writer
func (id *ID) Write(writer *bytes.Buffer) error {
	// In a real implementation, you would need to use the varuint encoding
	// For now, we'll use binary encoding as a placeholder
	err := binary.Write(writer, binary.LittleEndian, uint64(id.Client))
	if err != nil {
		return err
	}
	return binary.Write(writer, binary.LittleEndian, uint64(id.Clock))
}

// Read reads an ID from a reader
func ReadID(reader *bytes.Buffer) (*ID, error) {
	// In a real implementation, you would need to use the varuint decoding
	// For now, we'll use binary decoding as a placeholder
	var client, clock uint64
	err := binary.Read(reader, binary.LittleEndian, &client)
	if err != nil {
		return nil, err
	}
	err = binary.Read(reader, binary.LittleEndian, &clock)
	if err != nil {
		return nil, err
	}
	
	// In a real implementation, you would need to add debug assertions
	// if client < 0 || clock < 0 {
	//     panic("Client and clock should not be negative")
	// }
	
	return NewID(int64(client), int64(clock)), nil
}