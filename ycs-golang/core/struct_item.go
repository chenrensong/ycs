package core

import (
	"ycs/contracts"
)

// InfoFlags represents the bit flags for struct item properties
type InfoFlags int

const (
	InfoKeep      InfoFlags = 1 << 0
	InfoCountable InfoFlags = 1 << 1
	InfoDeleted   InfoFlags = 1 << 2
	InfoMarker    InfoFlags = 1 << 3
)

// StructItem represents the main struct item implementation
type StructItem struct {
	id          contracts.StructID
	length      int
	leftOrigin  *contracts.StructID
	left        contracts.IStructItem
	rightOrigin *contracts.StructID
	right       contracts.IStructItem
	parent      interface{}
	parentSub   *string
	redone      *contracts.StructID
	content     contracts.IContentEx
	info        InfoFlags
}

// NewStructItem creates a new StructItem
func NewStructItem(id contracts.StructID, left contracts.IStructItem, leftOrigin *contracts.StructID, right contracts.IStructItem, rightOrigin *contracts.StructID, parent interface{}, parentSub *string, content contracts.IContentEx) *StructItem {
	item := &StructItem{
		id:          id,
		length:      content.GetLength(),
		leftOrigin:  leftOrigin,
		left:        left,
		right:       right,
		rightOrigin: rightOrigin,
		parent:      parent,
		parentSub:   parentSub,
		redone:      nil,
		content:     content,
		info:        0,
	}

	if content.GetCountable() {
		item.info |= InfoCountable
	}

	return item
}

// GetID returns the struct ID
func (si *StructItem) GetID() contracts.StructID {
	return si.id
}

// SetID sets the struct ID
func (si *StructItem) SetID(id contracts.StructID) {
	si.id = id
}

// GetLength returns the length
func (si *StructItem) GetLength() int {
	return si.length
}

// SetLength sets the length
func (si *StructItem) SetLength(length int) {
	si.length = length
}

// GetLeftOrigin returns the left origin
func (si *StructItem) GetLeftOrigin() *contracts.StructID {
	return si.leftOrigin
}

// SetLeftOrigin sets the left origin
func (si *StructItem) SetLeftOrigin(leftOrigin *contracts.StructID) {
	si.leftOrigin = leftOrigin
}

// GetLeft returns the left item
func (si *StructItem) GetLeft() contracts.IStructItem {
	return si.left
}

// SetLeft sets the left item
func (si *StructItem) SetLeft(left contracts.IStructItem) {
	si.left = left
}

// GetRightOrigin returns the right origin
func (si *StructItem) GetRightOrigin() *contracts.StructID {
	return si.rightOrigin
}

// SetRightOrigin sets the right origin
func (si *StructItem) SetRightOrigin(rightOrigin *contracts.StructID) {
	if rightOrigin == nil {
		si.rightOrigin = nil
	} else {
		id := *rightOrigin
		si.rightOrigin = &id
	}
}

// GetRight returns the right item
func (si *StructItem) GetRight() contracts.IStructItem {
	return si.right
}

// SetRight sets the right item
func (si *StructItem) SetRight(right contracts.IStructItem) {
	si.right = right
}

// GetParent returns the parent
func (si *StructItem) GetParent() interface{} {
	return si.parent
}

// SetParent sets the parent
func (si *StructItem) SetParent(parent interface{}) {
	si.parent = parent
}

// GetParentSub returns the parent sub key
func (si *StructItem) GetParentSub() string {
	if si.parentSub == nil {
		return ""
	}
	return *si.parentSub
}

// SetParentSub sets the parent sub key
func (si *StructItem) SetParentSub(parentSub string) {
	if parentSub == "" {
		si.parentSub = nil
	} else {
		si.parentSub = &parentSub
	}
}

// GetRedone returns the redone ID
func (si *StructItem) GetRedone() *contracts.StructID {
	if si.redone == nil {
		return nil
	}
	return si.redone
}

// SetRedone sets the redone ID
func (si *StructItem) SetRedone(redone *contracts.StructID) {
	if redone == nil {
		si.redone = nil
	} else {
		id := *redone
		si.redone = &id
	}
}

// GetContent returns the content
func (si *StructItem) GetContent() contracts.IContentEx {
	return si.content
}

// SetContent sets the content
func (si *StructItem) SetContent(content contracts.IContentEx) {
	si.content = content
}

// IsMarker returns whether this is a marker item
func (si *StructItem) IsMarker() bool {
	return si.info&InfoMarker != 0
}

// GetMarker returns whether this is a marker item
func (si *StructItem) GetMarker() bool {
	return si.IsMarker()
}

// SetMarker sets the marker flag
func (si *StructItem) SetMarker(marker bool) {
	if marker {
		si.info |= InfoMarker
	} else {
		si.info &^= InfoMarker
	}
}

// IsKeep returns whether this item should be kept (not garbage collected)
func (si *StructItem) IsKeep() bool {
	return si.info&InfoKeep != 0
}

// GetKeep returns whether this item should be kept (not garbage collected)
func (si *StructItem) GetKeep() bool {
	return si.IsKeep()
}

// SetKeep sets the keep flag
func (si *StructItem) SetKeep(keep bool) {
	if keep {
		si.info |= InfoKeep
	} else {
		si.info &^= InfoKeep
	}
}

// IsCountable returns whether this item is countable
func (si *StructItem) IsCountable() bool {
	return si.info&InfoCountable != 0
}

// GetCountable returns whether this item is countable
func (si *StructItem) GetCountable() bool {
	return si.IsCountable()
}

// SetCountable sets the countable flag
func (si *StructItem) SetCountable(countable bool) {
	if countable {
		si.info |= InfoCountable
	} else {
		si.info &^= InfoCountable
	}
}

// IsDeleted returns whether this item is deleted
func (si *StructItem) IsDeleted() bool {
	return si.info&InfoDeleted != 0
}

// GetDeleted returns whether this item is deleted
func (si *StructItem) GetDeleted() bool {
	return si.info&InfoDeleted != 0
}

// IsGC returns false since this is not a GC struct
func (si *StructItem) IsGC() bool {
	return false
}

// GetLastID returns the last ID based on this struct's ID and length
func (si *StructItem) GetLastID() contracts.StructID {
	if si.length <= 1 {
		return si.id
	}
	lastID := contracts.StructID{
		Client: si.id.Client,
		Clock:  si.id.Clock + int64(si.length) - 1,
	}
	return lastID
}

// GetNext returns the next non-deleted item
func (si *StructItem) GetNext() contracts.IStructItem {
	n := si.right
	for n != nil && n.GetDeleted() {
		n = n.GetRight()
	}
	return n
}

// GetPrev returns the previous non-deleted item
func (si *StructItem) GetPrev() contracts.IStructItem {
	n := si.left
	for n != nil && n.GetDeleted() {
		n = n.GetLeft()
	}
	return n
}

// First returns the first non-deleted item
func (si *StructItem) First() contracts.IStructItem {
	// Note: This method should be called on AbstractType, not StructItem
	// For StructItem, we'll return the item itself if it's not deleted
	if !si.GetDeleted() {
		return si
	}
	return nil
}

// MarkDeleted marks this item as deleted
func (si *StructItem) MarkDeleted() {
	si.info |= InfoDeleted
}

// MergeWith tries to merge with the right item
func (si *StructItem) MergeWith(right contracts.IStructItem) bool {
	// Check if we can merge with the right item
	if si.id.Client != right.GetID().Client {
		return false
	}

	// Check if items are adjacent
	if si.id.Clock+int64(si.length) != right.GetID().Clock {
		return false
	}

	// Check if both items have the same parent
	if si.parent != right.GetParent() {
		return false
	}

	// Check if both items have the same parentSub
	if si.parentSub != nil && right.GetParentSub() != "" {
		if *si.parentSub != right.GetParentSub() {
			return false
		}
	} else if si.parentSub != nil || right.GetParentSub() != "" {
		return false
	}

	// Check if both items are either deleted or not deleted
	if si.GetDeleted() != right.GetDeleted() {
		return false
	}

	// Try to merge contents
	if si.content.MergeWith(right.GetContent()) {
		// Update length
		si.length += right.GetLength()
		return true
	}

	return false
}

// TryToMergeWithRight tries to merge with the right item
func (si *StructItem) TryToMergeWithRight(right contracts.IStructItem) bool {
	// Check if we can merge with the right item
	if si.id.Client != right.GetID().Client {
		return false
	}

	// Check if items are adjacent
	if si.id.Clock+int64(si.length) != right.GetID().Clock {
		return false
	}

	// Check if both items have the same parent
	if si.parent != right.GetParent() {
		return false
	}

	// Check if both items have the same parentSub
	siParentSub := si.GetParentSub()
	rightParentSub := right.GetParentSub()

	if siParentSub != "" && rightParentSub != "" {
		if siParentSub != rightParentSub {
			return false
		}
	} else if siParentSub != "" || rightParentSub != "" {
		return false
	}

	// Check if both items are either deleted or not deleted
	if si.GetDeleted() != right.GetDeleted() {
		return false
	}

	// Try to merge contents
	if si.content.MergeWith(right.GetContent()) {
		// Update length
		si.length += right.GetLength()
		return true
	}

	return false
}

// Delete marks this item as deleted
func (si *StructItem) Delete(transaction contracts.ITransaction) {
	if !si.GetDeleted() {
		if parent, ok := si.parent.(contracts.IAbstractType); ok {
			if si.IsCountable() && si.parentSub == nil {
				parent.SetLength(parent.GetLength() - si.length)
			}
		}

		si.MarkDeleted()
		transaction.GetDeleteSet().Add(si.id.Client, si.id.Clock, int64(si.length))

		if parent, ok := si.parent.(contracts.IAbstractType); ok {
			parentSub := ""
			if si.parentSub != nil {
				parentSub = *si.parentSub
			}
			transaction.AddChangedTypeToTransaction(parent, parentSub)
		}

		si.content.Delete(transaction)
	}
}

// Integrate integrates this item into the document
func (si *StructItem) Integrate(transaction contracts.ITransaction, offset int) {
	if offset > 0 {
		si.id = contracts.StructID{
			Client: si.id.Client,
			Clock:  si.id.Clock + int64(offset),
		}

		leftID := contracts.StructID{
			Client: si.id.Client,
			Clock:  si.id.Clock - 1,
		}

		si.left = transaction.GetDoc().GetStore().GetItemCleanEnd(transaction, leftID)

		if si.left != nil {
			lastID := si.left.GetLastID()
			si.leftOrigin = &lastID
		}

		si.content = si.content.Splice(offset).(contracts.IContentEx)
		si.length -= offset
	}

	if si.parent != nil {
		si.integrateIntoParent(transaction)
	} else {
		// Parent is not defined. Integrate GC struct instead
		gcStruct := NewStructGC(si.id, si.length)
		gcStruct.Integrate(transaction, 0)
		return
	}

	// Add to store
	transaction.GetDoc().GetStore().AddStruct(si)

	// Integrate content
	si.content.Integrate(transaction, si)

	// Add parent to transaction.changed
	if parent, ok := si.parent.(contracts.IAbstractType); ok {
		parentSub := ""
		if si.parentSub != nil {
			parentSub = *si.parentSub
		}
		transaction.AddChangedTypeToTransaction(parent, parentSub)

		// Delete if parent is deleted or if this is not the current attribute value of parent
		if (parent.GetItem() != nil && parent.GetItem().GetDeleted()) ||
			(si.parentSub != nil && si.right != nil) {
			si.Delete(transaction)
		}
	}
}

// integrateIntoParent handles the complex integration logic into parent
func (si *StructItem) integrateIntoParent(transaction contracts.ITransaction) {
	// Check if we need to find the correct position
	if (si.left == nil && (si.right == nil || si.right.GetLeft() != nil)) ||
		(si.left != nil && si.left.GetRight() != si.right) {

		var left contracts.IStructItem = si.left
		var o contracts.IStructItem

		// Set 'o' to the first conflicting item
		if left != nil {
			o = left.GetRight()
		} else if si.parentSub != nil {
			if parent, ok := si.parent.(contracts.IAbstractType); ok {
				parentMap := parent.GetMap()
				if item, exists := parentMap[*si.parentSub]; exists {
					o = item
					for o != nil && o.GetLeft() != nil {
						o = o.GetLeft()
					}
				}
			}
		} else {
			if parent, ok := si.parent.(contracts.IAbstractType); ok {
				o = parent.GetStart()
			}
		}

		conflictingItems := make(map[contracts.IStructItem]struct{})
		itemsBeforeOrigin := make(map[contracts.IStructItem]struct{})

		for o != nil && o != si.right {
			itemsBeforeOrigin[o] = struct{}{}
			conflictingItems[o] = struct{}{}

			if structIDEquals(si.leftOrigin, o.GetLeftOrigin()) {
				// Case 1
				if o.GetID().Client < si.id.Client {
					left = o
					conflictingItems = make(map[contracts.IStructItem]struct{})
				} else if structIDEquals(si.rightOrigin, o.GetRightOrigin()) {
					// This and 'o' are conflicting and point to the same integration points
					// The id decides which item comes first
					break
				}
			} else if o.GetLeftOrigin() != nil {
				leftOriginItem, err := transaction.GetDoc().GetStore().Find(*o.GetLeftOrigin())
				if err == nil && leftOriginItem != nil {
					if _, exists := itemsBeforeOrigin[leftOriginItem]; exists {
						// Case 2
						if _, conflicts := conflictingItems[leftOriginItem]; !conflicts {
							left = o
							conflictingItems = make(map[contracts.IStructItem]struct{})
						}
					}
				}
			} else {
				break
			}

			o = o.GetRight()
		}

		si.left = left
	}

	// Reconnect left/right + update parent map/start if necessary
	if si.left != nil {
		right := si.left.GetRight()
		si.right = right
		si.left.SetRight(si)
	} else {
		var r contracts.IStructItem

		if si.parentSub != nil {
			if parent, ok := si.parent.(contracts.IAbstractType); ok {
				parentMap := parent.GetMap()
				if item, exists := parentMap[*si.parentSub]; exists {
					r = item
					for r != nil && r.GetLeft() != nil {
						r = r.GetLeft()
					}
				}
			}
		} else {
			if parent, ok := si.parent.(contracts.IAbstractType); ok {
				r = parent.GetStart()
				parent.SetStart(si)
			}
		}

		si.right = r
	}

	if si.right != nil {
		si.right.SetLeft(si)
	} else if si.parentSub != nil {
		// Set as current parent value if right == nil and this is parentSub
		if parent, ok := si.parent.(contracts.IAbstractType); ok {
			parentMap := parent.GetMap()
			parentMap[*si.parentSub] = si
			// This is the current attribute value of parent. Delete left.
			if si.left != nil {
				si.left.Delete(transaction)
			}
		}
	}

	// Adjust length of parent
	if si.parentSub == nil && si.IsCountable() && !si.GetDeleted() {
		if parent, ok := si.parent.(contracts.IAbstractType); ok {
			parent.SetLength(parent.GetLength() + si.length)
		}
	}
}

// GetMissing returns the creator ClientID of the missing OP or defines missing items and returns nil
func (si *StructItem) GetMissing(transaction contracts.ITransaction, store contracts.IStructStore) *int64 {
	if si.leftOrigin != nil && si.leftOrigin.Client != si.id.Client && si.leftOrigin.Clock >= store.GetState(si.leftOrigin.Client) {
		return &si.leftOrigin.Client
	}

	if si.rightOrigin != nil && si.rightOrigin.Client != si.id.Client && si.rightOrigin.Clock >= store.GetState(si.rightOrigin.Client) {
		return &si.rightOrigin.Client
	}

	if parentID, ok := si.parent.(contracts.StructID); ok && si.id.Client != parentID.Client && parentID.Clock >= store.GetState(parentID.Client) {
		return &parentID.Client
	}

	// We have all missing ids, now find the items
	if si.leftOrigin != nil {
		si.left = store.GetItemCleanEnd(transaction, *si.leftOrigin)
		if si.left != nil {
			lastID := si.left.GetLastID()
			si.leftOrigin = &lastID
		}
	}

	if si.rightOrigin != nil {
		si.right = store.GetItemCleanStart(transaction, *si.rightOrigin)
		if si.right != nil {
			rightID := si.right.GetID()
			si.rightOrigin = &rightID
		}
	}

	if parentID, ok := si.parent.(contracts.StructID); ok {
		if parentItem, err := store.Find(parentID); err == nil {
			si.parent = parentItem
		}
	}

	return nil
}

// Write writes this item to an encoder
func (si *StructItem) Write(encoder contracts.IUpdateEncoder, offset int) error {
	hasLeftOrigin := si.leftOrigin != nil
	hasRightOrigin := si.rightOrigin != nil
	hasParentYKey := false
	hasParentSub := si.parentSub != nil

	// Determine parent info
	var parentYKey *string
	if si.parent != nil {
		if _, isStructID := si.parent.(contracts.StructID); !isStructID {
			if parent, ok := si.parent.(contracts.IAbstractType); ok {
				key := parent.FindRootTypeKey()
				parentYKey = &key
				hasParentYKey = true
			}
		}
	}

	// Calculate info byte
	info := si.content.GetRef()
	if hasLeftOrigin {
		info |= 0x80 // Bit8
	}
	if hasRightOrigin {
		info |= 0x40 // Bit7
	}
	if hasParentSub {
		info |= 0x20 // Bit6
	}

	encoder.WriteInfo(byte(info))

	if hasLeftOrigin {
		encoder.WriteLeftID(*si.leftOrigin)
	}
	if hasRightOrigin {
		encoder.WriteRightID(*si.rightOrigin)
	}

	if (info & (0x40 | 0x80)) == 0 {
		encoder.WriteParentInfo(hasParentYKey)
		if hasParentYKey {
			encoder.WriteString(*parentYKey)
		} else if !hasParentYKey {
			if parentID, ok := si.parent.(contracts.StructID); ok {
				encoder.WriteLeftID(parentID)
			}
		}
	}

	if hasParentSub {
		encoder.WriteString(*si.parentSub)
	}

	return si.content.Write(encoder, offset)
}

// IsVisible returns whether this item is visible in a snapshot
func (si *StructItem) IsVisible(snapshot contracts.ISnapshot) bool {
	stateVector := snapshot.GetStateVector()
	clientClock, exists := stateVector[si.id.Client]
	return exists && si.id.Clock < clientClock && !snapshot.GetDeleteSet().IsDeleted(si.id)
}

// KeepItemAndParents marks this item and its parents as kept (not to be garbage collected)
func (si *StructItem) KeepItemAndParents(keep bool) {
	var item contracts.IStructItem = si
	for item != nil && item.GetKeep() != keep {
		item.SetKeep(keep)
		if parent, ok := item.GetParent().(contracts.IAbstractType); ok {
			item = parent.GetItem()
		} else {
			break
		}
	}
}

// SplitItem splits this item at the given difference
func (si *StructItem) SplitItem(transaction contracts.ITransaction, diff int) contracts.IStructItem {
	if diff == 0 {
		return si
	}

	rightID := contracts.StructID{
		Client: si.id.Client,
		Clock:  si.id.Clock + int64(diff),
	}

	rightContent := si.content.Splice(diff)
	rightItem := NewStructItem(
		rightID,
		si,
		func() *contracts.StructID { id := si.GetLastID(); return &id }(),
		si.right,
		si.rightOrigin,
		si.parent,
		si.parentSub,
		rightContent.(contracts.IContentEx),
	)

	if si.GetDeleted() {
		rightItem.MarkDeleted()
	}

	if si.IsKeep() {
		rightItem.SetKeep(true)
	}

	if si.redone != nil {
		rightRedone := contracts.StructID{
			Client: si.redone.Client,
			Clock:  si.redone.Clock + int64(diff),
		}
		rightItem.redone = &rightRedone
	}

	// Update connections
	si.right = rightItem
	if rightItem.right != nil {
		rightItem.right.SetLeft(rightItem)
	}
	si.rightOrigin = &rightID
	si.length = diff

	return rightItem
}

// Gc performs garbage collection on this item
func (si *StructItem) Gc(store contracts.IStructStore, parentGCd bool) {
	if !si.GetDeleted() && !si.IsKeep() {
		si.content.Gc(store)
	}
}

// Helper function to compare StructID pointers
func structIDEquals(a, b *contracts.StructID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Client == b.Client && a.Clock == b.Clock
}
