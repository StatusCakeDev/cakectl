package spec

import "github.com/tomasbasham/gofp"

// ContactGroup is the desired state of a contact group.
type ContactGroup struct {
	Name           gofp.Option[string]   `json:"name,omitzero"`
	EmailAddresses gofp.Option[[]string] `json:"email_addresses,omitzero"`
	Integrations   gofp.Option[[]string] `json:"integrations,omitzero"`
	MobileNumbers  gofp.Option[[]string] `json:"mobile_numbers,omitzero"`
	PingURL        gofp.Option[string]   `json:"ping_url,omitzero"`
}

// ValidateCreate reports every reason the spec cannot create a group.
func (c ContactGroup) ValidateCreate() error {
	var v validator
	required(&v, "name", c.Name)
	return v.err()
}

// ValidateUpdate reports every reason the spec cannot update a group. Any
// subset of fields is acceptable.
func (c ContactGroup) ValidateUpdate() error {
	return nil
}
