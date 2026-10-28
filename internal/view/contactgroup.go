package view

import (
	"strconv"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/domain"
	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// ContactGroupRow is a contact group as shown in a listing. Email addresses,
// phone numbers and the ping URL are personal data or secrets, so only how
// many of each there are is shown.
type ContactGroupRow struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Emails       int    `json:"email_addresses"`
	Mobiles      int    `json:"mobile_numbers"`
	Integrations int    `json:"integrations"`
}

// ContactGroupList is a listing of contact groups.
type ContactGroupList []ContactGroupRow

// NewContactGroupList provides the listing of the given groups.
func NewContactGroupList(groups []domain.ContactGroup) ContactGroupList {
	rows := make(ContactGroupList, len(groups))
	for i, g := range groups {
		rows[i] = ContactGroupRow{
			ID: g.ID, Name: g.Name, Emails: len(g.EmailAddresses),
			Mobiles: len(g.MobileNumbers), Integrations: len(g.Integrations),
		}
	}
	return rows
}

// FormatText renders the listing as a table.
func (l ContactGroupList) FormatText() ([]byte, error) {
	rows := make([][]string, len(l))
	for i, r := range l {
		rows[i] = []string{r.ID, r.Name, strconv.Itoa(r.Emails), strconv.Itoa(r.Mobiles), strconv.Itoa(r.Integrations)}
	}
	return table([]string{"ID", "NAME", "EMAILS", "MOBILES", "INTEGRATIONS"}, rows, "No contact groups found."), nil
}

// ContactGroupDescription is the full detail of a contact group.
type ContactGroupDescription struct {
	ID   string            `json:"id"`
	Spec spec.ContactGroup `json:"spec"`
}

// NewContactGroupDescription describes a contact group.
func NewContactGroupDescription(g domain.ContactGroup) ContactGroupDescription {
	return ContactGroupDescription{
		ID: g.ID,
		Spec: spec.ContactGroup{
			Name:           gofp.Some(g.Name),
			EmailAddresses: gofp.Some(list(g.EmailAddresses)),
			Integrations:   gofp.Some(list(g.Integrations)),
			MobileNumbers:  gofp.Some(list(g.MobileNumbers)),
			PingURL:        someIf(g.PingURL),
		},
	}
}

// FormatText renders the description as labelled fields.
func (d ContactGroupDescription) FormatText() ([]byte, error) {
	s := d.Spec
	return fields(
		field{"ID", d.ID},
		field{"Name", show(s.Name)},
		field{"Email addresses", show(s.EmailAddresses)},
		field{"Mobile numbers", show(s.MobileNumbers)},
		field{"Integrations", show(s.Integrations)},
		field{"Ping URL", show(s.PingURL)},
	), nil
}
