package contracts

// ISnapshot represents a snapshot interface
type ISnapshot interface {
	GetDeleteSet() IDeleteSet
	GetStateVector() map[int64]int64
	EncodeSnapshotV2() []byte
	RestoreDocument(originDoc IYDoc, opts *YDocOptions) IYDoc
}
