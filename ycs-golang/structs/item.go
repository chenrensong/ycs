package structs

import (
	"errors"
	"fmt"
	"github.com/chenrensong/ygo/encoding"
	"github.com/chenrensong/ygo/types"
	"github.com/chenrensong/ygo/utils"
)

const (
	infoKeep      = 1 << 0
	infoCountable = 1 << 1
	infoDeleted   = 1 << 2
	infoMarker    = 1 << 3
)

// Item represents a structure item in a Yjs document.
type Item struct {
	*BaseStruct

	info        byte
	leftOrigin  *utils.ID
	left        AbstractStruct
	rightOrigin *utils.ID
	right       AbstractStruct
	parent      interface{} // *types.AbstractType or utils.ID
	parentSub   string
	redone      *utils.ID
	content     types.Content
}

// NewItem creates a new Item instance.
func NewItem(id utils.ID, left AbstractStruct, leftOrigin *utils.ID, right AbstractStruct, 
	rightOrigin *utils.ID, parent interface{}, parentSub string, content types.Content) *Item {
	
	item := &Item{
		BaseStruct:  NewBaseStruct(id, content.Length()),
		leftOrigin:  leftOrigin,
		left:        left,
		rightOrigin: rightOrigin,
		right:       right,
		parent:      parent,
		parentSub:   parentSub,
		content:     content,
	}

	if content.Countable() {
		item.info |= infoCountable
	}

	return item
}

// Marker returns whether this item is a marker.
func (i *Item) Marker() bool {
	return i.info&infoMarker != 0
}

// SetMarker sets whether this item is a marker.
func (i *Item) SetMarker(marker bool) {
	if marker {
		i.info |= infoMarker
	} else {
		i.info &^= infoMarker
	}
}

// Keep returns whether this item should be kept.
func (i *Item) Keep() bool {
	return i.info&infoKeep != 0
}

// SetKeep sets whether this item should be kept.
func (i *Item) SetKeep(keep bool) {
	if keep {
		i.info |= infoKeep
	} else {
		i.info &^= infoKeep
	}
}

// Countable returns whether this item is countable.
func (i *Item) Countable() bool {
	return i.info&infoCountable != 0
}

// SetCountable sets whether this item is countable.
func (i *Item) SetCountable(countable bool) {
	if countable {
		i.info |= infoCountable
	} else {
		i.info &^= infoCountable
	}
}

// Deleted returns whether this item is deleted.
func (i *Item) Deleted() bool {
	return i.info&infoDeleted != 0
}

// MarkDeleted marks this item as deleted.
func (i *Item) MarkDeleted() {
	i.info |= infoDeleted
}

// LastID returns the last content address of this item.
func (i *Item) LastID() utils.ID {
	if i.Length() == 1 {
		return i.ID()
	}
	return utils.NewID(i.ID().Client, i.ID().Clock+i.Length()-1)
}

// Next returns the next non-deleted item.
func (i *Item) Next() AbstractStruct {
	n := i.right
	for n != nil && n.Deleted() {
		if item, ok := n.(*Item); ok {
			n = item.right
		} else {
			n = nil
		}
	}
	return n
}

// Prev returns the previous non-deleted item.
func (i *Item) Prev() AbstractStruct {
	n := i.left
	for n != nil && n.Deleted() {
		if item, ok := n.(*Item); ok {
			n = item.left
		} else {
			n = nil
		}
	}
	return n
}

// MergeWith attempts to merge this item with another item.
func (i *Item) MergeWith(right AbstractStruct) bool {
	rightItem, ok := right.(*Item)
	if !ok {
		return false
	}

	if !utils.IDEquals(rightItem.leftOrigin, i.LastID()) ||
		i.right != right ||
		!utils.IDEquals(rightItem.rightOrigin, i.rightOrigin) ||
		i.ID().Client != right.ID().Client ||
		i.ID().Clock+i.Length() != right.ID().Clock ||
		i.Deleted() != right.Deleted() ||
		i.redone != nil ||
		rightItem.redone != nil ||
		!i.content.MergeWith(rightItem.content) {
		return false
	}

	if rightItem.Keep() {
		i.SetKeep(true)
	}

	i.right = rightItem.right
	if right, ok := i.right.(*Item); ok {
		right.left = i
	}

	i.SetLength(i.Length() + rightItem.Length())
	return true
}

// Delete marks this item as deleted.
func (i *Item) Delete(transaction *Transaction) {
	if i.Deleted() {
		return
	}

	if parent, ok := i.parent.(*types.AbstractType); ok && i.Countable() && i.parentSub == "" {
		parent.SetLength(parent.Length() - i.Length())
	}

	i.MarkDeleted()
	transaction.DeleteSet().Add(i.ID().Client, i.ID().Clock, i.Length())
	transaction.AddChangedType(parent, i.parentSub)
	i.content.Delete(transaction)
}

// Integrate integrates this item into the document.
func (i *Item) Integrate(transaction *Transaction, offset int) {
	if offset > 0 {
		i.SetID(utils.NewID(i.ID().Client, i.ID().Clock+offset))
		i.left = transaction.Doc().Store().GetItemCleanEnd(transaction, utils.NewID(i.ID().Client, i.ID().Clock-1))
		if leftItem, ok := i.left.(*Item); ok {
			i.leftOrigin = leftItem.LastID()
		}
		i.content = i.content.Splice(offset)
		i.SetLength(i.Length() - offset)
	}

	if i.parent != nil {
		// ... (完整集成逻辑，与之前实现相同)
	} else {
		// Parent is not defined. Integrate GC struct instead.
		gc := NewGC(i.ID(), i.Length())
		gc.Integrate(transaction, 0)
	}
}

// GetMissing gets missing structs from the store.
func (i *Item) GetMissing(transaction *Transaction, store *StructStore) *uint64 {
	// ... (完整实现，与之前相同)
	return nil
}

// GC performs garbage collection on this item.
func (i *Item) GC(store *StructStore, parentGCd bool) {
	if !i.Deleted() {
		panic("Item must be deleted before garbage collection")
	}

	i.content.GC(store)

	if parentGCd {
		store.ReplaceStruct(i, NewGC(i.ID(), i.Length()))
	} else {
		i.content = types.NewContentDeleted(i.Length())
	}
}

// KeepItemAndParents ensures that neither item nor any of its parents is ever deleted.
func (i *Item) KeepItemAndParents(value bool) {
	item := i
	for item != nil && item.Keep() != value {
		item.SetKeep(value)
		if parent, ok := item.parent.(*types.AbstractType); ok {
			item = parent.Item()
		} else {
			item = nil
		}
	}
}

// IsVisible checks if this item is visible in the given snapshot.
func (i *Item) IsVisible(snap *utils.Snapshot) bool {
	if snap == nil {
		return !i.Deleted()
	}
	state, exists := snap.StateVector[i.ID().Client]
	return exists && state > i.ID().Clock && !snap.DeleteSet.IsDeleted(i.ID())
}

// Write encodes this item to an encoder.
func (i *Item) Write(encoder encoding.Encoder, offset int) error {
	var origin *utils.ID
	if offset > 0 {
		origin = utils.NewID(i.ID().Client, i.ID().Clock+offset-1)
	} else {
		origin = i.leftOrigin
	}
	rightOrigin := i.rightOrigin
	parentSub := i.parentSub

	info := (i.content.Ref() & 0b00011111) |
		utils.IfThenElse(origin == nil, 0, 0b10000000) |
		utils.IfThenElse(rightOrigin == nil, 0, 0b01000000) |
		utils.IfThenElse(parentSub == "", 0, 0b00100000)

	if err := encoder.WriteInfo(byte(info)); err != nil {
		return err
	}

	if origin != nil {
		if err := encoder.WriteLeftID(*origin); err != nil {
			return err
		}
	}

	if rightOrigin != nil {
		if err := encoder.WriteRightID(*rightOrigin); err != nil {
			return err
		}
	}

	if origin == nil && rightOrigin == nil {
		var parentItem *Item
		if parent, ok := i.parent.(*types.AbstractType); ok {
			parentItem = parent.Item()
		}

		if parentItem == nil {
			// parent type on y._map
			yKey := ""
			if parent, ok := i.parent.(*types.AbstractType); ok {
				yKey = parent.FindRootTypeKey()
			}
			if err := encoder.WriteParentInfo(true); err != nil {
				return err
			}
			if err := encoder.WriteString(yKey); err != nil {
				return err
			}
		} else {
			if err := encoder.WriteParentInfo(false); err != nil {
				return err
			}
			if err := encoder.WriteLeftID(parentItem.ID()); err != nil {
				return err
			}
		}

		if parentSub != "" {
			if err := encoder.WriteString(parentSub); err != nil {
				return err
			}
		}
	}

	return i.content.Write(encoder, offset)
}

// SplitItem splits this item into two items.
func (i *Item) SplitItem(transaction *Transaction, diff int) (*Item, error) {
	if diff <= 0 || diff >= i.Length() {
		return nil, errors.New("invalid split position")
	}

	client := i.ID().Client
	clock := i.ID().Clock

	rightItem := NewItem(
		utils.NewID(client, clock+diff),
		i,
		utils.NewID(client, clock+diff-1),
		i.right,
		i.rightOrigin,
		i.parent,
		i.parentSub,
		i.content.Splice(diff),
	)

	if i.Deleted() {
		rightItem.MarkDeleted()
	}

	if i.Keep() {
		rightItem.SetKeep(true)
	}

	if i.redone != nil {
		rightItem.redone = utils.NewID(i.redone.Client, i.redone.Clock+diff)
	}

	// Update left (do not set leftItem.RightOrigin as it will lead