package content

import (
	"encoding/json"

	"github.com/chenrensong/ygo/contracts"
)

const ContentJsonRef = 2

type ContentJson struct {
	content []interface{}
}

func NewContentJson(data []interface{}) *ContentJson {
	content := make([]interface{}, len(data))
	copy(content, data)
	return &ContentJson{
		content: content,
	}
}

func (c *ContentJson) GetRef() int {
	return ContentJsonRef
}

func (c *ContentJson) GetCountable() bool {
	return true
}

func (c *ContentJson) GetLength() int {
	if c.content == nil {
		return 0
	}
	return len(c.content)
}

func (c *ContentJson) GetContent() []interface{} {
	result := make([]interface{}, len(c.content))
	copy(result, c.content)
	return result
}

func (c *ContentJson) Copy() contracts.IContent {
	return &ContentJson{content: c.content}
}

func (c *ContentJson) Splice(offset int) contracts.IContent {
	right := &ContentJson{
		content: make([]interface{}, len(c.content)-offset),
	}
	copy(right.content, c.content[offset:])
	c.content = c.content[:offset]
	return right
}

func (c *ContentJson) MergeWith(right contracts.IContent) bool {
	rightJson, ok := right.(*ContentJson)
	if !ok {
		return false
	}
	c.content = append(c.content, rightJson.content...)
	return true
}

func (c *ContentJson) Integrate(transaction contracts.ITransaction, item contracts.IStructItem) {
	// Do nothing
}

func (c *ContentJson) Delete(transaction contracts.ITransaction) {
	// Do nothing
}

func (c *ContentJson) Gc(store contracts.IStructStore) {
	// Do nothing
}

func (c *ContentJson) Write(encoder contracts.IUpdateEncoder, offset int) {
	length := len(c.content)
	encoder.WriteLength(length)
	for i := offset; i < length; i++ {
		jsonBytes, _ := json.Marshal(c.content[i])
		jsonStr := string(jsonBytes)
		encoder.WriteString(jsonStr)
	}
}

func ReadContentJson(decoder contracts.IUpdateDecoder) *ContentJson {
	length := decoder.ReadLength()
	content := make([]interface{}, length)

	for i := 0; i < length; i++ {
		jsonStr := decoder.ReadString()
		var jsonObj interface{}
		if jsonStr == "undefined" {
			jsonObj = nil
		} else {
			json.Unmarshal([]byte(jsonStr), &jsonObj)
		}
		content[i] = jsonObj
	}

	return &ContentJson{content: content}
}
