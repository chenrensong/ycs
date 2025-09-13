package contracts

// IDeleteSet represents a delete set interface
type IDeleteSet interface {
	GetClients() map[int64][]DeleteItem
	Add(client int64, clock int64, length int64)
	FindIndexSS(dis []DeleteItem, clock int64) *int
	IsDeleted(id StructID) bool
	IterateDeletedStructs(transaction ITransaction, fun func(IStructItem) bool)
	SortAndMergeDeleteSet()
	TryGc(store IStructStore, gcFilter func(IStructItem) bool)
	TryGcDeleteSet(store IStructStore, gcFilter func(IStructItem) bool)
	TryMergeDeleteSet(store IStructStore)
	Write(encoder IDSEncoder)
}
