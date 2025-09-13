// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

package structs

// GC represents a garbage collection marker
// This is the Go implementation of the C# GC class
type GC struct {
	AbstractStruct
}

// StructGCRefNumber is a constant reference number for GC
const (
	StructGCRefNumber byte = 0
)

// NewGC creates a new instance of GC
// id is the identifier for this struct
// length is the length of the GC range
func NewGC(id *ID, length int) (*GC, error) {
	abstractStruct, err := NewAbstractStruct(id, length)
	if err != nil {
		return nil, err
	}

	return &GC{
		AbstractStruct: *abstractStruct,
	}, nil
}

// Deleted returns true as GC structs represent deleted content
func (g *GC) Deleted() bool {
	return true
}

// MergeWith merges this GC struct with another struct
// Returns true if the merge was successful
func (g *GC) MergeWith(right *AbstractStruct) bool {
	// We expect to merge with another GC
	g.Length += right.Length
	return true
}

// Delete does nothing for GC
func (g *GC) Delete(transaction *Transaction) {
	// Do nothing (implementation as in C#)
}

// Integrate integrates the GC with a transaction
func (g *GC) Integrate(transaction *Transaction, offset int) {
	if offset > 0 {
		// Update the ID and length based on offset
		newId, _ := NewID(g.Id.Client(), g.Id.Clock()+offset)
		g.Id = newId
		g.Length -= offset
	}

	// Add this struct to the document's store
	transaction.Doc.Store.AddStruct(g)
}

// GetMissing returns nil as there is no missing data for GC
func (g *GC) GetMissing(transaction *Transaction, store *StructStore) *int {
	return nil
}

// Write writes the GC to an update encoder
func (g *GC) Write(encoder IUpdateEncoder, offset int) {
	// Write the GC reference number and length
	encoder.WriteInfo(StructGCRefNumber)
	encoder.WriteLength(g.Length - offset)
}

// CheckDisposed is a placeholder for debug checks
func (g *GC) CheckDisposed() {
	// In debug mode, this would check if the encoder has been disposed
	// and panic if it has. For now, we'll keep it simple.
}
