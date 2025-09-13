package utils

import (
	"fmt"
	"sort"
	"sync"

	"github.com/chenrensong/ygo/structs"
	"github.com/chenrensong/ygo/types"
	"github.com/chenrensong/ygo/utils"
)

// Transaction represents a set of changes to a Yjs document.
type Transaction struct {
	doc                 *Doc
	origin              interface{}
	local               bool
	beforeState         map[uint64]uint64
	afterState          map[uint64]uint64
	changed             map[*types.AbstractType]map[string]struct{}
	changedParentTypes  map[*types.AbstractType][]*types.YEvent
	meta                map[string]interface{}
	deleteSet           *structs.DeleteSet
	mergeStructs        []structs.AbstractStruct
	subdocsAdded        map[*Doc]struct{}
	subdocsRemoved      map[*Doc]struct{}
	subdocsLoaded       map[*Doc]struct{}
	cleanupActions      []func()
	cleanupActionsMutex sync.Mutex
}

// NewTransaction creates a new Transaction instance.
func NewTransaction(doc *Doc, origin interface{}, local bool) *Transaction {
	return &Transaction{
		doc:                doc,
		origin:             origin,
		local:              local,
		beforeState:        doc.Store.GetStateVector(),
		afterState:         make(map[uint64]uint64),
		changed:            make(map[*types.AbstractType]map[string]struct{}),
		changedParentTypes: make(map[*types.AbstractType][]*types.YEvent),
		meta:               make(map[string]interface{}),
		deleteSet:          structs.NewDeleteSet(),
		mergeStructs:       make([]structs.AbstractStruct, 0),
		subdocsAdded:       make(map[*Doc]struct{}),
		subdocsRemoved:     make(map[*Doc]struct{}),
		subdocsLoaded:      make(map[*Doc]struct{}),
	}
}

// Doc returns the document associated with this transaction.
func (t *Transaction) Doc() interface{} {
	return t.doc
}

// Origin returns the origin of this transaction.
func (t *Transaction) Origin() interface{} {
	return t.origin
}

// Local returns whether this transaction is local.
func (t *Transaction) Local() bool {
	return t.local
}

// DeleteSet returns the delete set for this transaction.
func (t *Transaction) DeleteSet() interface{} {
	return t.deleteSet
}

// AddChangedType adds a changed type to the transaction.
func (t *Transaction) AddChangedType(type_ *types.AbstractType, parentSub string) {
	item := type_.Item()
	if item == nil || (t.beforeState[item.ID().Client] > item.ID().Clock && !item.Deleted()) {
		if _, exists := t.changed[type_]; !exists {
			t.changed[type_] = make(map[string]struct{})
		}
		t.changed[type_][parentSub] = struct{}{}
	}
}

// Cleanup cleans up the transaction.
func (t *Transaction) Cleanup(cleanups []*Transaction, i int) {
	if i < len(cleanups) {
		transaction := cleanups[i]
		doc := transaction.doc
		store := doc.store
		ds := transaction.deleteSet
		mergeStructs := transaction.mergeStructs

		defer func() {
			// Replace deleted items with ItemDeleted/GC
			if doc.gc {
				ds.TryGCDeleteSet(store, doc.gcFilter)
			}

			ds.TryMergeDeleteSet(store)

			// On all affected store.clients props, try to merge
			for client, clock := range transaction.afterState {
				beforeClock, exists := transaction.beforeState[client]
				if !exists {
					beforeClock = 0
				}

				if beforeClock != clock {
					structs := store.Clients[client]
					firstChangePos := max(structs.FindIndexSS(beforeClock), 1)
					for j := len(structs) - 1; j >= firstChangePos; j-- {
						structs.DeleteSet.TryToMergeWithLeft(j)
					}
				}
			}

			// Try to merge mergeStructs
			for j := 0; j < len(mergeStructs); j++ {
				client := mergeStructs[j].ID().Client
				clock := mergeStructs[j].ID().Clock
				structs := store.Clients[client]
				replacedStructPos := structs.FindIndexSS(clock)

				if replacedStructPos+1 < len(structs) {
					structs.DeleteSet.TryToMergeWithLeft(replacedStructPos + 1)
				}

				if replacedStructPos > 0 {
					structs.DeleteSet.TryToMergeWithLeft(replacedStructPos)
				}
			}

			if !transaction.local {
				afterClock, afterExists := transaction.afterState[doc.ClientID]
				beforeClock, beforeExists := transaction.beforeState[doc.ClientID]

				if !afterExists {
					afterClock = ^uint64(0)
				}
				if !beforeExists {
					beforeClock = ^uint64(0)
				}

				if afterClock != beforeClock {
					doc.ClientID = GenerateNewClientID()
				}
			}

			doc.onAfterTransactionCleanup(transaction)

			doc.invokeUpdateV2(transaction)

			for subDoc := range transaction.subdocsAdded {
				doc.subdocs[subDoc] = struct{}{}
			}

			for subDoc := range transaction.subdocsRemoved {
				delete(doc.subdocs, subDoc)
			}

			doc.invokeSubdocsChanged(transaction.subdocsLoaded, transaction.subdocsAdded, transaction.subdocsRemoved)

			for subDoc := range transaction.subdocsRemoved {
				subDoc.Destroy()
			}

			if len(cleanups) <= i+1 {
				doc.transactionCleanups = nil
				doc.invokeAfterAllTransactions(cleanups)
			} else {
				CleanupTransactions(cleanups, i+1)
			}
		}()

		ds.SortAndMergeDeleteSet()
		transaction.afterState = store.GetStateVector()
		doc.transaction = nil

		t.cleanupActionsMutex.Lock()
		defer t.cleanupActionsMutex.Unlock()

		t.cleanupActions = append(t.cleanupActions, func() {
			doc.invokeOnBeforeObserverCalls(transaction)
		})

		for type_, subs := range transaction.changed {
			if type_.Item() == nil || !type_.Item().Deleted() {
				t.cleanupActions = append(t.cleanupActions, func() {
					type_.CallObserver(transaction, subs)
				})
			}
		}

		for type_, events := range transaction.changedParentTypes {
			if type_.Item() == nil || !type_.Item().Deleted() {
				sortedEvents := make([]*types.YEvent, len(events))
				copy(sortedEvents, events)
				sort.Slice(sortedEvents, func(i, j int) bool {
					return len(sortedEvents[i].Path) < len(sortedEvents[j].Path)
				})

				t.cleanupActions = append(t.cleanupActions, func() {
					type_.CallDeepEventHandlerListeners(sortedEvents, transaction)
				})
			}
		}

		t.cleanupActions = append(t.cleanupActions, func() {
			doc.invokeOnAfterTransaction(transaction)
		})

		t.callAllCleanupActions()
	}
}

// RedoItem redoes an item in the transaction.
func (t *Transaction) RedoItem(item *structs.Item, redoItems map[*structs.Item]struct{}) structs.AbstractStruct {
	doc := t.doc
	store := doc.store
	ownClientID := doc.ClientID
	redone := item.Redone()

	if redone != nil {
		return store.GetItemCleanStart(t, *redone)
	}

	var parentItem *structs.Item
	if parent, ok := item.Parent().(*types.AbstractType); ok {
		parentItem = parent.Item()
	}

	var left, right structs.AbstractStruct

	if item.ParentSub() == "" {
		// Is an array item, insert at old position
		left = item.Left()
		right = item
	} else {
		// Is a map item, insert at current value
		left = item
		for left != nil && left.Right() != nil {
			left = left.Right()
			if left.ID().Client != ownClientID {
				// Cannot redo due to conflict
				return nil
			}
		}

		if left != nil && left.Right() != nil {
			if parent, ok := item.Parent().(*types.AbstractType); ok {
				left = parent.Map()[item.ParentSub()]
			}
		}

		right = nil
	}

	// Ensure parent is redone
	if parentItem != nil && parentItem.Deleted() && parentItem.Redone() == nil {
		if _, exists := redoItems[parentItem]; !exists || t.RedoItem(parentItem, redoItems) == nil {
			return nil
		}
	}

	if parentItem != nil && parentItem.Redone() != nil {
		for parentItem.Redone() != nil {
			parentItem = store.GetItemCleanStart(t, *parentItem.Redone()).(*structs.Item)
		}

		// Find next cloned redo items
		for left != nil {
			leftTrace := left
			for leftTrace != nil {
				if item, ok := leftTrace.(*structs.Item); ok {
					if parent, ok := item.Parent().(*types.AbstractType); ok {
						if parent.Item() == parentItem {
							left = leftTrace
							break
						}
					}
					leftTrace = item.Redone()
					if leftTrace != nil {
						leftTrace = store.GetItemCleanStart(t, *leftTrace)
					}
				} else {
					break
				}
			}
			if leftTrace != nil {
				break
			}
			left = left.Left()
		}

		for right != nil {
			rightTrace := right
			for rightTrace != nil {
				if item, ok := rightTrace.(*structs.Item); ok {
					if parent, ok := item.Parent().(*types.AbstractType); ok {
						if parent.Item() == parentItem {
							right = rightTrace
							break
						}
					}
					rightTrace = item.Redone()
					if rightTrace != nil {
						rightTrace = store.GetItemCleanStart(t, *rightTrace)
					}
				} else {
					break
				}
			}
			if rightTrace != nil {
				break
			}
			right = right.Right()
		}
	}

	nextClock := store.GetState(ownClientID)
	nextID := utils.NewID(ownClientID, nextClock)

	var parent interface{}
	if parentItem == nil {
		parent = item.Parent()
	} else {
		if content, ok := parentItem.Content().(*types.ContentType); ok {
			parent = content.Type()
		}
	}

	redoneItem := structs.NewItem(
		nextID,
		left,
		left.LastID(),
		right,
		right.ID(),
		parent,
		item.ParentSub(),
		item.Content().Copy(),
	)

	item.SetRedone(&nextID)
	redoneItem.KeepItemAndParents(true)
	redoneItem.Integrate(t, 0)

	return redoneItem
}

// SplitSnapshotAffectedStructs splits structs affected by a snapshot.
func (t *Transaction) SplitSnapshotAffectedStructs(snapshot *utils.Snapshot) {
	metaKey := "splitSnapshotAffectedStructs"
	metaValue, exists := t.meta[metaKey]
	if !exists {
		metaValue = make(map[*utils.Snapshot]struct{})
		t.meta[metaKey] = metaValue
	}

	meta := metaValue.(map[*utils.Snapshot]struct{})
	store := t.doc.store

	if _, alreadySplit := meta[snapshot]; !alreadySplit {
		for client, clock := range snapshot.StateVector {
			if clock < store.GetState(client) {
				store.GetItemCleanStart(t, utils.NewID(client, clock))
			}
		}

		snapshot.DeleteSet.IterateDeletedStructs(t, func(item structs.AbstractStruct) bool {
			return true
		})

		meta[snapshot] = struct{}{}
	}
}

// WriteUpdateMessageFromTransaction writes an update message from the transaction.
func (t *Transaction) WriteUpdateMessageFromTransaction(encoder encoding.Encoder) (bool, error) {
	if len(t.deleteSet.Clients) == 0 {
		allUnchanged := true
		for client, afterClock := range t.afterState {
			if beforeClock, exists := t.beforeState[client]; !exists || beforeClock != afterClock {
				allUnchanged = false
				break
			}
		}
		if allUnchanged {
			return false, nil
		}
	}

	t.deleteSet.SortAndMergeDeleteSet()
	if err := encoding.WriteClientsStructs(encoder, t.doc.store, t.beforeState); err != nil {
		return false, fmt.Errorf("failed to write clients structs: %w", err)
	}

	if err := t.deleteSet.Write(encoder); err != nil {
		return false, fmt.Errorf("failed to write delete set: %w", err)
	}

	return true, nil
}

// callAllCleanupActions calls all cleanup actions.
func (t *Transaction) callAllCleanupActions() {
	t.cleanupActionsMutex.Lock()
	defer t.cleanupActionsMutex.Unlock()

	for _, action := range t.cleanupActions {
		action()
	}
	t.cleanupActions = nil
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
