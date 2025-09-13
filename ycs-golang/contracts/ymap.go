package contracts

// IYMap represents a Y map interface
type IYMap interface {
	GetCount() int
	CallObserver(transaction ITransaction, parentSubs map[string]struct{})
	Clone() IYMap
	ContainsKey(key string) bool
	Delete(key string)
	Get(key string) interface{}
	GetEnumerator() map[string]interface{} // Go doesn't have IEnumerator, using map
	Integrate(doc IYDoc, item IStructItem)
	InternalClone() IAbstractType
	InternalCopy() IAbstractType
	Keys() []string
	Set(key string, value interface{})
	Values() []interface{}
	Write(encoder IUpdateEncoder)
}
