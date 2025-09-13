package contracts

// DeleteItem represents an item to be deleted
type DeleteItem struct {
	Clock  int64
	Length int64
}

// NewDeleteItem creates a new DeleteItem
func NewDeleteItem(clock, length int64) DeleteItem {
	return DeleteItem{
		Clock:  clock,
		Length: length,
	}
}
