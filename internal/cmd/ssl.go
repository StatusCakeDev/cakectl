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

func newSSLCommand(o *RootOptions) *cobra.Command {
	cmd := newResourceCommand(o, resource[spec.SSL]{
		use:           "ssl",
		aliases:       []string{"ssl-check"},
		noun:          "SSL check",
		short:         "Manage SSL certificate checks",
		flags:         sslFlags,
		createExample: `cakectl ssl create --website-url https://example.com --check-rate 86400 --alert-at 7,30`,
		create:        func(ctx context.Context, c *adapter.Client, s spec.SSL) (string, error) { return c.CreateSSL(ctx, s) },
		update: func(ctx context.Context, c *adapter.Client, id string, s spec.SSL) error {
			return c.UpdateSSL(ctx, id, s)
		},
		remove: func(ctx context.Context, c *adapter.Client, id string) error { return c.DeleteSSL(ctx, id) },
		describe: func(ctx context.Context, c *adapter.Client, id string) (any, error) {
			t, err := c.GetSSL(ctx, id)
			return view.NewSSLDescription(t), err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List SSL checks",
		Example: templates.Examples(`cakectl ssl list`),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			tests, err := api.ListSSL(commandContext(cmd))
			if err != nil {
				return err
			}
			return o.print(view.NewSSLList(tests))
		},
	})

	return cmd
}

func sslFlags(fs *pflag.FlagSet, _ bool) *binder[spec.SSL] {
	b := newBinder[spec.SSL](fs)

	bindString(b, "website-url", "HTTPS URL whose certificate is checked", func(s *spec.SSL) *gofp.Option[string] { return &s.WebsiteURL })
	bindInt32(b, "check-rate", "Seconds between checks (300, 600, 1800, 3600, 86400, 2073600)", func(s *spec.SSL) *gofp.Option[int32] { return &s.CheckRate })
	bindInt32s(b, "alert-at", "Days before expiry at which to alert", func(s *spec.SSL) *gofp.Option[[]int32] { return &s.AlertAt })
	bindBool(b, "alert-broken", "Alert when the certificate is broken", func(s *spec.SSL) *gofp.Option[bool] { return &s.AlertBroken })
	bindBool(b, "alert-expiry", "Alert when the certificate expires", func(s *spec.SSL) *gofp.Option[bool] { return &s.AlertExpiry })
	bindBool(b, "alert-mixed", "Alert on mixed content", func(s *spec.SSL) *gofp.Option[bool] { return &s.AlertMixed })
	bindBool(b, "alert-reminder", "Send reminders as expiry approaches", func(s *spec.SSL) *gofp.Option[bool] { return &s.AlertReminder })
	bindStrings(b, "contact-groups", "IDs of contact groups to alert", func(s *spec.SSL) *gofp.Option[[]string] { return &s.ContactGroups })
	bindBool(b, "follow-redirects", "Follow redirects", func(s *spec.SSL) *gofp.Option[bool] { return &s.FollowRedirects })
	bindString(b, "hostname", "Hostname to present during the TLS handshake", func(s *spec.SSL) *gofp.Option[string] { return &s.Hostname })
	bindBool(b, "paused", "Pause the check", func(s *spec.SSL) *gofp.Option[bool] { return &s.Paused })
	bindString(b, "user-agent", "User agent to send", func(s *spec.SSL) *gofp.Option[string] { return &s.UserAgent })

	return b
}
