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

func newHeartbeatCommand(o *RootOptions) *cobra.Command {
	cmd := newResourceCommand(o, resource[spec.Heartbeat]{
		use:           "heartbeat",
		aliases:       []string{"heartbeat-check"},
		noun:          "heartbeat check",
		short:         "Manage heartbeat checks",
		flags:         heartbeatFlags,
		createExample: `cakectl heartbeat create --name "Nightly backup" --period 86400`,
		create: func(ctx context.Context, c *adapters.Client, s spec.Heartbeat) (string, error) {
			return c.CreateHeartbeat(ctx, s)
		},
		update: func(ctx context.Context, c *adapters.Client, id string, s spec.Heartbeat) error {
			return c.UpdateHeartbeat(ctx, id, s)
		},
		remove: func(ctx context.Context, c *adapters.Client, id string) error { return c.DeleteHeartbeat(ctx, id) },
		describe: func(ctx context.Context, c *adapters.Client, id string) (any, error) {
			t, err := c.GetHeartbeat(ctx, id)
			return view.NewHeartbeatDescription(t), err
		},
	})

	var filter adapters.HeartbeatFilter
	list := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List heartbeat checks",
		Example: templates.Examples(`
      cakectl heartbeat list
      cakectl heartbeat list --status down`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			tests, err := api.ListHeartbeat(commandContext(cmd), filter)
			if err != nil {
				return err
			}
			return o.print(view.NewHeartbeatList(tests))
		},
	}
	addCheckFilterFlags(list.Flags(), &filter)
	cmd.AddCommand(list)

	return cmd
}

func heartbeatFlags(fs *pflag.FlagSet, _ bool) *binder[spec.Heartbeat] {
	b := newBinder[spec.Heartbeat](fs)

	bindString(b, "name", "Name of the check", func(s *spec.Heartbeat) *gofp.Option[string] { return &s.Name })
	bindInt32(b, "period", "Seconds to wait for a ping before alerting", func(s *spec.Heartbeat) *gofp.Option[int32] { return &s.Period })
	bindStrings(b, "contact-groups", "IDs of contact groups to alert", func(s *spec.Heartbeat) *gofp.Option[[]string] { return &s.ContactGroups })
	bindString(b, "host", "Name of the hosting provider", func(s *spec.Heartbeat) *gofp.Option[string] { return &s.Host })
	bindBool(b, "paused", "Pause the check", func(s *spec.Heartbeat) *gofp.Option[bool] { return &s.Paused })
	bindStrings(b, "tags", "Tags", func(s *spec.Heartbeat) *gofp.Option[[]string] { return &s.Tags })

	return b
}
