package contracts

// IYArray represents a Y array interface
type IYArray interface {
	IAbstractType
	GetLength() int
	Add(content []interface{})
	CallObserver(transaction ITransaction, parentSubs map[string]struct{})
	Clone() IYArray
	Delete(index int, length ...int) // length defaults to 1
	Get(index int) interface{}
	Insert(index int, content []interface{})
	Integrate(doc IYDoc, item IStructItem)
	InternalClone() IAbstractType
	InternalCopy() IAbstractType
	Slice(start ...int) []interface{} // start defaults to 0, optional end parameter
	ToArray() []interface{}
	Unshift(content []interface{})
	Write(encoder IUpdateEncoder)
}

// IYArrayBase represents the base Y array interface
type IYArrayBase interface {
	ClearSearchMarkers()
}
