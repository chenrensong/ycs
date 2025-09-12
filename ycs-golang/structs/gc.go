package structs

// GC represents a garbage collected struct
type GC struct {
	*AbstractStruct
}

// NewGC creates a new GC
func NewGC(id *ID, length int) *GC {
	return &GC{
		AbstractStruct: NewAbstractStruct(id, length),
	}
}

// StructGCRefNumber is the reference number for GC structs
const StructGCRefNumber byte = 0

// Deleted returns whether this struct is deleted
func (g *GC) Deleted() bool {
	return true
}

// MergeWith merges this struct with the right struct
func (g *GC) MergeWith(right *AbstractStruct) bool {
	// In Go, we use type assertion to check the type
	if _, ok := right.(*GC); ok {
		g.Length += right.Length
		return true
	}
	return false
}

// Delete deletes this struct
func (g *GC) Delete(transaction *Transaction) {
	// Do nothing
}

// Integrate integrates this struct
func (g *GC) Integrate(transaction *Transaction, offset int) {
	if offset > 0 {
		// In a real implementation, you would need to update the ID
		// g.Id = new ID(g.Id.Client, g.Id.Clock + offset)
		g.Length -= offset
	}

	// In a real implementation, you would need to add to the store
	// transaction.Doc.Store.AddStruct(g)
}

// GetMissing returns the creator ClientID of the missing OP or define missing items and return null
func (g *GC) GetMissing(transaction *Transaction, store *StructStore) *int64 {
	return nil
}

// Write writes this struct to an encoder
func (g *GC) Write(encoder IUpdateEncoder, offset int) {
	encoder.WriteInfo(StructGCRefNumber)
	encoder.WriteLength(g.Length - offset)
}