package adapters

import (
	"context"
	"fmt"

	"github.com/StatusCakeDev/statuscake-go"

	"github.com/StatusCakeDev/cakectl/internal/spec"
)

// ListContactGroups lists every contact group.
func (c *Client) ListContactGroups(ctx context.Context) ([]statuscake.ContactGroup, error) {
	items, err := collect(func(page, limit int32) ([]statuscake.ContactGroup, statuscake.Pagination, error) {
		res, err := c.client.ListContactGroups(ctx).Page(page).Limit(limit).Execute()
		return res.Data, res.Metadata, err
	})
	return items, translate("list contact groups", err)
}

// GetContactGroup reads a contact group.
func (c *Client) GetContactGroup(ctx context.Context, id string) (statuscake.ContactGroup, error) {
	res, err := c.client.GetContactGroup(ctx, id).Execute()
	return res.Data, translate("get contact group "+id, err)
}

// CreateContactGroup creates a contact group and returns its ID.
func (c *Client) CreateContactGroup(ctx context.Context, s spec.ContactGroup) (string, error) {
	if err := s.ValidateCreate(); err != nil {
		return "", fmt.Errorf("invalid contact group: %w", err)
	}

	r := c.client.CreateContactGroup(ctx)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.EmailAddresses, func(v []string) { r = r.EmailAddresses(v) })
	apply(s.Integrations, func(v []string) { r = r.Integrations(v) })
	apply(s.MobileNumbers, func(v []string) { r = r.MobileNumbers(v) })
	apply(s.PingURL, func(v string) { r = r.PingURL(v) })

	res, err := r.Execute()
	return res.Data.NewID, translate("create contact group", err)
}

// UpdateContactGroup changes the fields the spec sets on a contact group.
func (c *Client) UpdateContactGroup(ctx context.Context, id string, s spec.ContactGroup) error {
	if spec.IsEmpty(s) {
		return ErrNothingToUpdate
	}
	if err := s.ValidateUpdate(); err != nil {
		return fmt.Errorf("invalid contact group: %w", err)
	}

	r := c.client.UpdateContactGroup(ctx, id)
	apply(s.Name, func(v string) { r = r.Name(v) })
	apply(s.EmailAddresses, func(v []string) { r = r.EmailAddresses(v) })
	apply(s.Integrations, func(v []string) { r = r.Integrations(v) })
	apply(s.MobileNumbers, func(v []string) { r = r.MobileNumbers(v) })
	apply(s.PingURL, func(v string) { r = r.PingURL(v) })

	return translate("update contact group "+id, r.Execute())
}

// DeleteContactGroup deletes a contact group.
func (c *Client) DeleteContactGroup(ctx context.Context, id string) error {
	return translate("delete contact group "+id, c.client.DeleteContactGroup(ctx, id).Execute())
}
