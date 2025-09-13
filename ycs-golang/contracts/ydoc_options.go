package contracts

import (
	"crypto/rand"
	"fmt"
)

// generateGUID creates a simple GUID-like string
func generateGUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// YDocOptions represents options for YDoc
type YDocOptions struct {
	Gc       bool
	GcFilter func(IStructItem) bool
	Guid     string
	Meta     map[string]string
	AutoLoad bool
}

// NewYDocOptions creates new YDocOptions with default values
func NewYDocOptions() *YDocOptions {
	defaultPredicate := func(item IStructItem) bool { return true }

	return &YDocOptions{
		Gc:       true,
		GcFilter: defaultPredicate,
		Guid:     generateGUID(),
		Meta:     nil,
		AutoLoad: false,
	}
}

// Clone creates a copy of YDocOptions
func (opts *YDocOptions) Clone() *YDocOptions {
	newOpts := &YDocOptions{
		Gc:       opts.Gc,
		GcFilter: opts.GcFilter,
		Guid:     opts.Guid,
		AutoLoad: opts.AutoLoad,
	}

	if opts.Meta != nil {
		newOpts.Meta = make(map[string]string)
		for k, v := range opts.Meta {
			newOpts.Meta[k] = v
		}
	}

	return newOpts
}

// Write writes the options using an encoder
func (opts *YDocOptions) Write(encoder IUpdateEncoder, offset int) {
	dict := make(map[string]interface{})
	dict["gc"] = opts.Gc
	dict["guid"] = opts.Guid
	dict["autoLoad"] = opts.AutoLoad

	if opts.Meta != nil {
		dict["meta"] = opts.Meta
	}

	encoder.WriteAny(dict)
}

// ReadYDocOptions reads YDocOptions from a decoder
func ReadYDocOptions(decoder IUpdateDecoder) *YDocOptions {
	dict := decoder.ReadAny().(map[string]interface{})

	result := NewYDocOptions()

	if gc, ok := dict["gc"]; ok {
		result.Gc = gc.(bool)
	}

	if guid, ok := dict["guid"]; ok {
		result.Guid = guid.(string)
	} else {
		result.Guid = generateGUID()
	}

	if meta, ok := dict["meta"]; ok {
		result.Meta = meta.(map[string]string)
	}

	if autoLoad, ok := dict["autoLoad"]; ok {
		result.AutoLoad = autoLoad.(bool)
	}

	return result
}
