package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/tomasbasham/gofp"

	"github.com/StatusCakeDev/cakectl/internal/adapters"
	"github.com/StatusCakeDev/cakectl/internal/spec"
	"github.com/StatusCakeDev/cakectl/internal/view"
)

func newMaintenanceWindowCommand(o *RootOptions) *cobra.Command {
	cmd := newResourceCommand(o, resource[spec.Maintenance]{
		use:           "maintenance-window",
		aliases:       []string{"maintenance-windows", "mw"},
		noun:          "maintenance window",
		short:         "Manage maintenance windows",
		flags:         maintenanceFlags,
		createExample: `cakectl maintenance-window create --name "Deploy" --start-at 2026-11-01T22:00:00Z --end-at 2026-11-02T01:00:00Z --timezone Europe/London --tests 123,456`,
		create: func(ctx context.Context, c *adapters.Client, s spec.Maintenance) (string, error) {
			return c.CreateMaintenance(ctx, s)
		},
		update: func(ctx context.Context, c *adapters.Client, id string, s spec.Maintenance) error {
			return c.UpdateMaintenance(ctx, id, s)
		},
		remove: func(ctx context.Context, c *adapters.Client, id string) error { return c.DeleteMaintenance(ctx, id) },
		describe: func(ctx context.Context, c *adapters.Client, id string) (any, error) {
			w, err := c.GetMaintenance(ctx, id)
			return view.NewMaintenanceDescription(w), err
		},
	})

	var state string
	list := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List maintenance windows",
		Example: templates.Examples(`
      cakectl maintenance-window list
      cakectl maintenance-window list --state active`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			windows, err := api.ListMaintenance(commandContext(cmd), state)
			if err != nil {
				return err
			}
			return o.print(view.NewMaintenanceList(windows))
		},
	}
	list.Flags().StringVar(&state, "state", "", "Only windows in this state (active, paused or pending)")
	cmd.AddCommand(list)

	return cmd
}

func maintenanceFlags(fs *pflag.FlagSet, _ bool) *binder[spec.Maintenance] {
	b := newBinder[spec.Maintenance](fs)

	bindString(b, "name", "Name of the window", func(s *spec.Maintenance) *gofp.Option[string] { return &s.Name })
	bindTime(b, "start-at", "When the window starts", func(s *spec.Maintenance) *gofp.Option[time.Time] { return &s.StartAt })
	bindTime(b, "end-at", "When the window ends", func(s *spec.Maintenance) *gofp.Option[time.Time] { return &s.EndAt })
	bindString(b, "timezone", "IANA time zone of the window, such as Europe/London", func(s *spec.Maintenance) *gofp.Option[string] { return &s.Timezone })
	bindString(b, "repeat-interval", "How often the window recurs (never, 1d, 1w, 2w, 1m)", func(s *spec.Maintenance) *gofp.Option[string] { return &s.RepeatInterval })
	bindStrings(b, "tags", "Silence checks carrying these tags", func(s *spec.Maintenance) *gofp.Option[[]string] { return &s.Tags })
	bindStrings(b, "tests", "IDs of the checks to silence", func(s *spec.Maintenance) *gofp.Option[[]string] { return &s.Tests })

	return b
}
