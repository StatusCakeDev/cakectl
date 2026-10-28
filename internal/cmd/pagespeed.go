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

func newPagespeedCommand(o *RootOptions) *cobra.Command {
	cmd := newResourceCommand(o, resource[spec.Pagespeed]{
		use:           "pagespeed",
		aliases:       []string{"pagespeed-check"},
		noun:          "pagespeed check",
		short:         "Manage pagespeed checks",
		flags:         pagespeedFlags,
		createExample: `cakectl pagespeed create --name "Landing page" --website-url https://example.com --check-rate 3600 --region UK`,
		create: func(ctx context.Context, c *adapter.Client, s spec.Pagespeed) (string, error) {
			return c.CreatePagespeed(ctx, s)
		},
		update: func(ctx context.Context, c *adapter.Client, id string, s spec.Pagespeed) error {
			return c.UpdatePagespeed(ctx, id, s)
		},
		remove: func(ctx context.Context, c *adapter.Client, id string) error { return c.DeletePagespeed(ctx, id) },
		describe: func(ctx context.Context, c *adapter.Client, id string) (any, error) {
			t, err := c.GetPagespeed(ctx, id)
			return view.NewPagespeedDescription(t), err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List pagespeed checks",
		Example: templates.Examples(`cakectl pagespeed list`),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			tests, err := api.ListPagespeed(commandContext(cmd))
			if err != nil {
				return err
			}
			return o.print(view.NewPagespeedList(tests))
		},
	})

	history := &cobra.Command{
		Use:   "history ID",
		Short: "List the tests a check has run",
		Example: templates.Examples(`
      cakectl pagespeed history 123
      cakectl pagespeed history 123 --limit 100 --after 2026-10-01T00:00:00Z`),
		Args: cobra.ExactArgs(1),
	}
	window := windowFlags(history.Flags())
	history.RunE = func(cmd *cobra.Command, args []string) error {
		w, err := window()
		if err != nil {
			return err
		}

		api, err := o.client(cmd)
		if err != nil {
			return err
		}

		results, err := api.ListPagespeedHistory(commandContext(cmd), args[0], w)
		if err != nil {
			return err
		}
		return o.print(view.NewPagespeedResultList(results))
	}
	cmd.AddCommand(history)

	return cmd
}

func pagespeedFlags(fs *pflag.FlagSet, create bool) *binder[spec.Pagespeed] {
	b := newBinder[spec.Pagespeed](fs)

	bindString(b, "name", "Name of the check", func(s *spec.Pagespeed) *gofp.Option[string] { return &s.Name })
	bindString(b, "website-url", "URL of the page to test", func(s *spec.Pagespeed) *gofp.Option[string] { return &s.WebsiteURL })
	bindInt32(b, "check-rate", "Seconds between checks (60, 300, 600, 900, 1800, 3600, 86400)", func(s *spec.Pagespeed) *gofp.Option[int32] { return &s.CheckRate })
	bindString(b, "region", "Region to test from (AU, CA, DE, FR, IN, JP, NL, SG, UK, US, USW)", func(s *spec.Pagespeed) *gofp.Option[string] { return &s.Region })
	bindInt32(b, "alert-bigger", "Alert if the page is bigger than this many kilobytes (0 disables)", func(s *spec.Pagespeed) *gofp.Option[int32] { return &s.AlertBigger })
	bindInt64(b, "alert-slower", "Alert if the page loads slower than this many milliseconds (0 disables)", func(s *spec.Pagespeed) *gofp.Option[int64] { return &s.AlertSlower })
	bindInt32(b, "alert-smaller", "Alert if the page makes fewer than this many requests (0 disables)", func(s *spec.Pagespeed) *gofp.Option[int32] { return &s.AlertSmaller })
	bindStrings(b, "contact-groups", "IDs of contact groups to alert", func(s *spec.Pagespeed) *gofp.Option[[]string] { return &s.ContactGroups })
	bindBool(b, "paused", "Pause the check", func(s *spec.Pagespeed) *gofp.Option[bool] { return &s.Paused })

	return b
}
