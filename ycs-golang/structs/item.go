package structs

// Item represents an item in the Yjs document structure
type Item struct {
	*AbstractStruct
	LeftOrigin  *ID
	Left        *AbstractStruct
	RightOrigin *ID
	Right       *AbstractStruct
	Parent      interface{}
	ParentSub   string
	Redone      *ID
	Content     ContentEx
	info        int // Using int to represent the InfoEnum flags
}

// NewItem creates a new Item
func NewItem(id *ID, left *AbstractStruct, leftOrigin *ID, right *AbstractStruct, rightOrigin *ID, parent interface{}, parentSub string, content Content) *Item {
	item := &Item{
		AbstractStruct: NewAbstractStruct(id, content.Length()),
		LeftOrigin:     leftOrigin,
		Left:           left,
		RightOrigin:    rightOrigin,
		Right:          right,
		Parent:         parent,
		ParentSub:      parentSub,
		Redone:         nil,
		Content:        content.(ContentEx), // Assuming content implements ContentEx
		info:           0,
	}

	// Set countable flag based on content
	if content.Countable() {
		item.info |= 1 << 1 // Countable flag
	}

	return item
}

// Marker gets or sets the marker flag
func (i *Item) Marker() bool {
	return (i.info & (1 << 3)) != 0
}

func (i *Item) SetMarker(value bool) {
	if value {
		i.info |= 1 << 3
	} else {
		i.info &^= 1 << 3
	}
}

// Keep gets or sets the keep flag
func (i *Item) Keep() bool {
	return (i.info & (1 << 0)) != 0
}

func (i *Item) SetKeep(value bool) {
	if value {
		i.info |= 1 << 0
	} else {
		i.info &^= 1 << 0
	}
}

// Countable gets or sets the countable flag
func (i *Item) Countable() bool {
	return (i.info & (1 << 1)) != 0
}

func (i *Item) SetCountable(value bool) {
	if value {
		i.info |= 1 << 1
	} else {
		i.info &^= 1 << 1
	}
}

// Deleted returns whether the item is deleted
func (i *Item) Deleted() bool {
	return (i.info & (1 << 2)) != 0
}

// LastId computes the last content address of this Item
func (i *Item) LastId() *ID {
	if i.Length == 1 {
		return i.ID
	}
	// Return new ID(Id.Client, Id.Clock + Length - 1)
	return &ID{} // Placeholder
}

// Next returns the next non-deleted item
func (i *Item) Next() *AbstractStruct {
	n := i.Right
	for n != nil && n.Deleted() {
		// In Go, we need to type assert to access Item-specific fields
		if item, ok := n.(*Item); ok {
			n = item.Right
		} else {
			break
		}
	}
	return n
}

// Prev returns the previous non-deleted item
func (i *Item) Prev() *AbstractStruct {
	n := i.Left
	for n != nil && n.Deleted() {
		// In Go, we need to type assert to access Item-specific fields
		if item, ok := n.(*Item); ok {
			n = item.Left
		} else {
			break
		}
	}
	return n
}

// MarkDeleted marks this item as deleted
func (i *Item) MarkDeleted() {
	i.info |= 1 << 2
}

// MergeWith tries to merge two items
func (i *Item) MergeWith(right *AbstractStruct) bool {
	rightItem, ok := right.(*Item)
	if !ok {
		return false
	}

	// Simplified implementation - in a real implementation, you would need to check all the conditions
	// as in the C# version
	
	if i.Right == right && i.Content.Ref() == rightItem.Content.Ref() {
		// Merge the content
		if i.Content.MergeWith(rightItem.Content) {
			if rightItem.Keep() {
				i.SetKeep(true)
			}

			i.Right = rightItem.Right
			if rightItem.Right != nil {
				if rightItemRight, ok := rightItem.Right.(*Item); ok {
					rightItemRight.Left = i
				}
			}

			i.Length += rightItem.Length
			return true
		}
	}

	return false
}

// Delete marks this item as deleted
func (i *Item) Delete(transaction *Transaction) {
	if !i.Deleted() {
		// In a real implementation, you would need to handle the parent type
		// and update the parent's length if countable
		
		i.MarkDeleted()
		
		// In a real implementation, you would need to add to the delete set
		// transaction.DeleteSet.Add(i.Id.Client, i.Id.Clock, i.Length)
		
		// In a real implementation, you would need to add the changed type to the transaction
		// transaction.AddChangedTypeToTransaction(parent, i.ParentSub)
		
		i.Content.Delete(transaction)
	}
}

// Integrate integrates this item
func (i *Item) Integrate(transaction *Transaction, offset int) {
	// This is a simplified implementation
	// A full implementation would need to handle all the logic from the C# version
	
	if offset > 0 {
		// In a real implementation, you would need to update the ID
		// i.Id = new ID(i.Id.Client, i.Id.Clock + offset)
		
		// In a real implementation, you would need to get the left item
		// i.Left = transaction.Doc.Store.GetItemCleanEnd(transaction, new ID(i.Id.Client, i.Id.Clock - 1))
		
		// In a real implementation, you would need to update the left origin
		// i.LeftOrigin = i.Left.LastId()
		
		// In a real implementation, you would need to splice the content
		// i.Content = i.Content.Splice(offset)
		
		i.Length -= offset
	}

	// In a real implementation, you would need to handle the parent integration logic
	// This is complex and would require implementing much of the logic from the C# version
	
	if i.Parent != nil {
		// Add to store
		// transaction.Doc.Store.AddStruct(i)
		
		// Integrate content
		i.Content.Integrate(transaction, i)
		
		// Add parent to transaction.changed
		// transaction.AddChangedTypeToTransaction(i.Parent, i.ParentSub)
	} else {
		// Parent is not defined. Integrate GC struct instead.
		// new GC(i.Id, i.Length).Integrate(transaction, 0)
	}
}

// GetMissing returns the creator ClientID of the missing OP or define missing items and return null
func (i *Item) GetMissing(transaction *Transaction, store *StructStore) *int64 {
	// This is a simplified implementation
	// A full implementation would need to handle all the logic from the C# version
	
	// In a real implementation, you would need to check the left and right origins
	// and the parent to see if any are missing
	
	return nil
}

// Write writes this item to an encoder
func (i *Item) Write(encoder IUpdateEncoder, offset int) {
	// This is a simplified implementation
	// A full implementation would need to handle all the logic from the C# version
	
	// In a real implementation, you would need to write the item data to the encoder
}

// SplitItem splits this item into two items
func (i *Item) SplitItem(transaction *Transaction, diff int) *Item {
	// This is a simplified implementation
	// A full implementation would need to handle all the logic from the C# version
	
	// In a real implementation, you would need to create a new item with the split content
	// and update the links between items
	
	return nil
}