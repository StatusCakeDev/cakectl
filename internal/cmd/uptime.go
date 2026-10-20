package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/adapters"
	"github.com/StatusCakeDev/cakectl/internal/spec"
	"github.com/StatusCakeDev/cakectl/internal/view"
)

func newUptimeCommand(o *RootOptions) *cobra.Command {
	cmd := newResourceCommand(o, resource[spec.Uptime]{
		use:           "uptime",
		aliases:       []string{"uptime-check"},
		noun:          "uptime check",
		short:         "Manage uptime checks",
		flags:         uptimeFlags,
		createExample: `cakectl uptime create --name "Home page" --test-type HTTP --website-url https://example.com --check-rate 300`,
		create: func(ctx context.Context, c *adapters.Client, s spec.Uptime) (string, error) {
			return c.CreateUptime(ctx, s)
		},
		update: func(ctx context.Context, c *adapters.Client, id string, s spec.Uptime) error {
			return c.UpdateUptime(ctx, id, s)
		},
		remove: func(ctx context.Context, c *adapters.Client, id string) error { return c.DeleteUptime(ctx, id) },
		describe: func(ctx context.Context, c *adapters.Client, id string) (any, error) {
			t, err := c.GetUptime(ctx, id)
			return view.NewUptimeDescription(t), err
		},
	})

	cmd.AddCommand(newUptimeListCommand(o))
	cmd.AddCommand(newUptimeHistoryCommand(o, "alerts", "List the alerts a check has raised",
		func(ctx context.Context, c *adapters.Client, id string, w adapters.Window) (any, error) {
			r, err := c.ListUptimeAlerts(ctx, id, w)
			return view.NewAlertList(r), err
		}))
	cmd.AddCommand(newUptimeHistoryCommand(o, "history", "List the tests a check has run",
		func(ctx context.Context, c *adapters.Client, id string, w adapters.Window) (any, error) {
			r, err := c.ListUptimeHistory(ctx, id, w)
			return view.NewHistoryList(r), err
		}))
	cmd.AddCommand(newUptimeHistoryCommand(o, "periods", "List the periods a check spent up or down",
		func(ctx context.Context, c *adapters.Client, id string, w adapters.Window) (any, error) {
			r, err := c.ListUptimePeriods(ctx, id, w)
			return view.NewPeriodList(r), err
		}))

	return cmd
}

func uptimeFlags(fs *pflag.FlagSet, create bool) *binder[spec.Uptime] {
	b := newBinder[spec.Uptime](fs)

	bindString(b, "name", "Name of the check", func(s *spec.Uptime) *gofp.Option[string] { return &s.Name })
	if create {
		bindString(b, "test-type", "Kind of check (DNS, HEAD, HTTP, PING, SMTP, SSH, TCP)", func(s *spec.Uptime) *gofp.Option[string] { return &s.TestType })
	}
	bindString(b, "website-url", "URL or address to check", func(s *spec.Uptime) *gofp.Option[string] { return &s.WebsiteURL })
	bindInt32(b, "check-rate", "Seconds between checks (0, 30, 60, 300, 900, 1800, 3600, 86400)", func(s *spec.Uptime) *gofp.Option[int32] { return &s.CheckRate })
	bindString(b, "basic-username", "Username for HTTP basic authentication", func(s *spec.Uptime) *gofp.Option[string] { return &s.BasicUsername })
	bindString(b, "basic-password", "Password for HTTP basic authentication", func(s *spec.Uptime) *gofp.Option[string] { return &s.BasicPassword })
	bindInt32(b, "confirmation", "Confirmation servers that must agree before an alert", func(s *spec.Uptime) *gofp.Option[int32] { return &s.Confirmation })
	bindStrings(b, "contact-groups", "IDs of contact groups to alert", func(s *spec.Uptime) *gofp.Option[[]string] { return &s.ContactGroups })
	bindString(b, "custom-header", "Custom request headers as a JSON object", func(s *spec.Uptime) *gofp.Option[string] { return &s.CustomHeader })
	bindStrings(b, "dns-ips", "Expected IP addresses of a DNS check", func(s *spec.Uptime) *gofp.Option[[]string] { return &s.DNSIPs })
	bindString(b, "dns-server", "DNS server a DNS check queries", func(s *spec.Uptime) *gofp.Option[string] { return &s.DNSServer })
	bindBool(b, "do-not-find", "Alert if the find string is present rather than absent", func(s *spec.Uptime) *gofp.Option[bool] { return &s.DoNotFind })
	bindBool(b, "enable-ssl-alert", "Alert when the certificate is about to expire", func(s *spec.Uptime) *gofp.Option[bool] { return &s.EnableSSLAlert })
	bindString(b, "final-endpoint", "URL the check must end up at after redirects", func(s *spec.Uptime) *gofp.Option[string] { return &s.FinalEndpoint })
	bindString(b, "find-string", "Text the response must contain", func(s *spec.Uptime) *gofp.Option[string] { return &s.FindString })
	bindBool(b, "follow-redirects", "Follow redirects", func(s *spec.Uptime) *gofp.Option[bool] { return &s.FollowRedirects })
	bindString(b, "host", "Name of the hosting provider", func(s *spec.Uptime) *gofp.Option[string] { return &s.Host })
	bindBool(b, "include-header", "Include response headers when searching for the find string", func(s *spec.Uptime) *gofp.Option[bool] { return &s.IncludeHeader })
	bindBool(b, "paused", "Pause the check", func(s *spec.Uptime) *gofp.Option[bool] { return &s.Paused })
	bindInt32(b, "port", "Port a TCP, SMTP or SSH check connects to", func(s *spec.Uptime) *gofp.Option[int32] { return &s.Port })
	bindString(b, "post-body", "Form data to send with the request, as JSON", func(s *spec.Uptime) *gofp.Option[string] { return &s.PostBody })
	bindString(b, "post-raw", "Raw body to send with the request", func(s *spec.Uptime) *gofp.Option[string] { return &s.PostRaw })
	bindStrings(b, "regions", "Region codes to check from (see: cakectl locations uptime)", func(s *spec.Uptime) *gofp.Option[[]string] { return &s.Regions })
	bindStrings(b, "status-codes", "HTTP status codes that count as down", func(s *spec.Uptime) *gofp.Option[[]string] { return &s.StatusCodes })
	bindStrings(b, "tags", "Tags", func(s *spec.Uptime) *gofp.Option[[]string] { return &s.Tags })
	bindInt32(b, "timeout", "Seconds to wait for a response", func(s *spec.Uptime) *gofp.Option[int32] { return &s.Timeout })
	bindInt32(b, "trigger-rate", "Failed checks, in a row, before an alert", func(s *spec.Uptime) *gofp.Option[int32] { return &s.TriggerRate })
	bindBool(b, "use-jar", "Keep cookies between redirects", func(s *spec.Uptime) *gofp.Option[bool] { return &s.UseJAR })
	bindString(b, "user-agent", "User agent to send", func(s *spec.Uptime) *gofp.Option[string] { return &s.UserAgent })

	return b
}

func newUptimeListCommand(o *RootOptions) *cobra.Command {
	var filter adapters.UptimeFilter

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List uptime checks",
		Example: templates.Examples(`
      cakectl uptime list
      cakectl uptime list --status down
      cakectl uptime list --tags prod,web --match-any`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			tests, err := api.ListUptime(commandContext(cmd), filter)
			if err != nil {
				return err
			}
			return o.print(view.NewUptimeList(tests))
		},
	}

	addCheckFilterFlags(cmd.Flags(), &filter)
	return cmd
}

// addCheckFilterFlags registers the filters shared by uptime and heartbeat
// listings.
func addCheckFilterFlags(fs *pflag.FlagSet, filter *adapters.UptimeFilter) {
	fs.StringVar(&filter.Status, "status", "", "Only checks with this status (up or down)")
	fs.StringSliceVar(&filter.Tags, "tags", nil, "Only checks with these tags")
	fs.BoolVar(&filter.MatchAny, "match-any", false, "Match checks with any of the tags rather than all")
}

func newUptimeHistoryCommand(
	o *RootOptions,
	use, short string,
	fetch func(context.Context, *adapters.Client, string, adapters.Window) (any, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use + " ID",
		Short: short,
		Example: templates.Examples(`
      cakectl uptime ` + use + ` 123
      cakectl uptime ` + use + ` 123 --limit 100 --after 2026-10-01T00:00:00Z`),
		Args: cobra.ExactArgs(1),
	}

	window := windowFlags(cmd.Flags())

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		w, err := window()
		if err != nil {
			return err
		}

		api, err := o.client(cmd)
		if err != nil {
			return err
		}

		v, err := fetch(commandContext(cmd), api, args[0], w)
		if err != nil {
			return err
		}
		return o.print(v)
	}
	return cmd
}
