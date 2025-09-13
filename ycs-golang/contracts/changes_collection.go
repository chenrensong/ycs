package contracts

// ChangesCollection represents a collection of changes
type ChangesCollection struct {
	Added   map[IStructItem]struct{}
	Deleted map[IStructItem]struct{}
	Delta   []Delta
	Keys    map[string]ChangeKey
}

// NewChangesCollection creates a new ChangesCollection
func NewChangesCollection() *ChangesCollection {
	return &ChangesCollection{
		Added:   make(map[IStructItem]struct{}),
		Deleted: make(map[IStructItem]struct{}),
		Delta:   make([]Delta, 0),
		Keys:    make(map[string]ChangeKey),
	}
}

// Delta represents a delta change
type Delta struct {
	Insert     interface{}
	Delete     *int
	Retain     *int
	Attributes map[string]interface{}
}

// NewDelta creates a new Delta
func NewDelta() *Delta {
	return &Delta{
		Attributes: make(map[string]interface{}),
	}
}

// ChangeAction represents the type of change
type ChangeAction int

const (
	ChangeActionAdd ChangeAction = iota
	ChangeActionUpdate
	ChangeActionDelete
)

// String returns string representation of ChangeAction
func (c ChangeAction) String() string {
	switch c {
	case ChangeActionAdd:
		return "Add"
	case ChangeActionUpdate:
		return "Update"
	case ChangeActionDelete:
		return "Delete"
	default:
		return "Unknown"
	}
}

// ChangeKey represents a key change
type ChangeKey struct {
	Action   ChangeAction
	OldValue interface{}
}

// NewChangeKey creates a new ChangeKey
func NewChangeKey(action ChangeAction, oldValue interface{}) ChangeKey {
	return ChangeKey{
		Action:   action,
		OldValue: oldValue,
	}
}
