package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/StatusCakeDev/cakectl/internal/view"
)

func newLocationsCommand(o *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "locations",
		Aliases: []string{"location"},
		Short:   "List the places checks can run from",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "uptime",
		Short: "List the locations uptime checks run from",
		Long: templates.LongDesc(`
      List the locations uptime checks run from. Use the region codes with the
      --regions flag of "uptime create".`),
		Example: templates.Examples(`cakectl locations uptime`),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			locations, err := api.ListUptimeLocations(commandContext(cmd))
			if err != nil {
				return err
			}
			return o.print(view.NewLocationList(locations))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "pagespeed",
		Short: "List the locations pagespeed checks run from",
		Long: templates.LongDesc(`
      List the locations pagespeed checks run from. Use the region codes with
      the --region flag of "pagespeed create".`),
		Example: templates.Examples(`cakectl locations pagespeed`),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			locations, err := api.ListPagespeedLocations(commandContext(cmd))
			if err != nil {
				return err
			}
			return o.print(view.NewLocationList(locations))
		},
	})

	return cmd
}
