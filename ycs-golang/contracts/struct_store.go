package contracts

// IStructStore represents a struct store interface
type IStructStore interface {
	GetClients() map[int64][]IStructItem
	AddStruct(str IStructItem)
	CleanupPendingStructs()
	Find(id StructID) (IStructItem, error)
	FindIndexCleanStart(transaction ITransaction, structs []IStructItem, clock int64) int
	FollowRedone(id StructID) (IStructItem, int)
	GetItemCleanEnd(transaction ITransaction, id StructID) IStructItem
	GetItemCleanStart(transaction ITransaction, id StructID) IStructItem
	GetState(clientID int64) int64
	GetStateVector() map[int64]int64
	IntegrityCheck()
	IterateStructs(transaction ITransaction, structs []IStructItem, clockStart int64, length int64, fun func(IStructItem) bool)
	MergeReadStructsIntoPendingReads(clientStructRefs map[int64][]IStructItem)
	ReadAndApplyDeleteSet(decoder IDSDecoder, transaction ITransaction) error
	ReplaceStruct(oldStruct IStructItem, newStruct IStructItem) error
	ResumeStructIntegration(transaction ITransaction)
	TryResumePendingDeleteReaders(transaction ITransaction)
}
