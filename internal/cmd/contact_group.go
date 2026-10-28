package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/adapter"
	"github.com/StatusCakeDev/cakectl/internal/spec"
	"github.com/StatusCakeDev/cakectl/internal/view"
)

func newContactGroupCommand(o *RootOptions) *cobra.Command {
	cmd := newResourceCommand(o, resource[spec.ContactGroup]{
		use:           "contact-group",
		aliases:       []string{"contact-groups", "cg"},
		noun:          "contact group",
		short:         "Manage contact groups",
		flags:         contactGroupFlags,
		createExample: `cakectl contact-group create --name "On call" --email-addresses oncall@example.com`,
		create: func(ctx context.Context, c *adapter.Client, s spec.ContactGroup) (string, error) {
			return c.CreateContactGroup(ctx, s)
		},
		update: func(ctx context.Context, c *adapter.Client, id string, s spec.ContactGroup) error {
			return c.UpdateContactGroup(ctx, id, s)
		},
		remove: func(ctx context.Context, c *adapter.Client, id string) error { return c.DeleteContactGroup(ctx, id) },
		describe: func(ctx context.Context, c *adapter.Client, id string) (any, error) {
			g, err := c.GetContactGroup(ctx, id)
			return view.NewContactGroupDescription(g), err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List contact groups",
		Long: templates.LongDesc(`
      List contact groups. Email addresses, phone numbers and ping URLs are not
      shown in listings; use describe to see them.`),
		Example: templates.Examples(`cakectl contact-group list`),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			groups, err := api.ListContactGroups(commandContext(cmd))
			if err != nil {
				return err
			}
			return o.print(view.NewContactGroupList(groups))
		},
	})

	return cmd
}

func contactGroupFlags(fs *pflag.FlagSet, _ bool) *binder[spec.ContactGroup] {
	b := newBinder[spec.ContactGroup](fs)

	bindString(b, "name", "Name of the group", func(s *spec.ContactGroup) *gofp.Option[string] { return &s.Name })
	bindStrings(b, "email-addresses", "Email addresses to notify", func(s *spec.ContactGroup) *gofp.Option[[]string] { return &s.EmailAddresses })
	bindStrings(b, "integrations", "IDs of integrations to notify", func(s *spec.ContactGroup) *gofp.Option[[]string] { return &s.Integrations })
	bindStrings(b, "mobile-numbers", "Mobile numbers to notify, in international format", func(s *spec.ContactGroup) *gofp.Option[[]string] { return &s.MobileNumbers })
	bindString(b, "ping-url", "URL to request when an alert is raised", func(s *spec.ContactGroup) *gofp.Option[string] { return &s.PingURL })

	return b
}
