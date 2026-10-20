package view

import "fmt"

// Outcome reports a change that was made to a resource.
type Outcome struct {
	Resource string `json:"resource"`
	ID       string `json:"id"`
	Action   string `json:"action"`
}

// Created reports that a resource was created.
func Created(resource, id string) Outcome {
	return Outcome{Resource: resource, ID: id, Action: "created"}
}

// Updated reports that a resource was updated.
func Updated(resource, id string) Outcome {
	return Outcome{Resource: resource, ID: id, Action: "updated"}
}

// Deleted reports that a resource was deleted.
func Deleted(resource, id string) Outcome {
	return Outcome{Resource: resource, ID: id, Action: "deleted"}
}

// FormatText renders the outcome as a sentence.
func (o Outcome) FormatText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s %s %s\n", o.Resource, o.ID, o.Action)), nil
}
