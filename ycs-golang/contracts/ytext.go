package contracts

// YTextChangeType represents the type of text change
type YTextChangeType int

const (
	YTextChangeTypeAdded YTextChangeType = iota
	YTextChangeTypeRemoved
)

// String returns string representation of YTextChangeType
func (y YTextChangeType) String() string {
	switch y {
	case YTextChangeTypeAdded:
		return "Added"
	case YTextChangeTypeRemoved:
		return "Removed"
	default:
		return "Unknown"
	}
}

// YTextChangeAttributes represents attributes for text changes
type YTextChangeAttributes struct {
	Type  YTextChangeType
	User  int
	State YTextChangeType
}

// IYText represents a Y text interface
type IYText interface {
	ApplyDelta(delta []Delta, sanitize ...bool) // sanitize defaults to true
	CallObserver(transaction ITransaction, parentSubs map[string]struct{})
	Clone() IYText
	Delete(index int, length int)
	Format(index int, length int, attributes map[string]interface{})
	GetAttribute(name string) interface{}
	GetAttributes() map[string]interface{}
	Insert(index int, text string, attributes ...map[string]interface{})            // attributes is optional
	InsertEmbed(index int, embed interface{}, attributes ...map[string]interface{}) // attributes is optional
	Integrate(doc IYDoc, item IStructItem)
	InternalClone() IAbstractType
	RemoveAttribute(name string)
	SetAttribute(name string, value interface{})
	ToDelta(snapshot ISnapshot, prevSnapshot ISnapshot, computeYChange func(YTextChangeType, StructID, YTextChangeAttributes) interface{}) []Delta
	ToString() string
	Write(encoder IUpdateEncoder)
}
