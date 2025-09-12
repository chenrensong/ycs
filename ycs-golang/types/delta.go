package types

// Delta represents a delta operation
type Delta struct {
	Insert    interface{}            `json:"insert,omitempty"`
	Delete    *int                   `json:"delete,omitempty"`
	Retain    *int                   `json:"retain,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}