package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/StatusCakeDev/cakectl/internal/configstore"
)

func newConfigCommand(o *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long: templates.LongDesc(fmt.Sprintf(`
      View and change the settings cakectl keeps in its config file.

      The settings are: %s.

      Each setting can also be given as a flag or as an environment variable
      such as STATUSCAKE_API_TOKEN; a flag wins over the environment, and the
      environment wins over the file.`, strings.Join(configstore.Keys(), ", "))),
	}

	cmd.AddCommand(newConfigGetCommand(o))
	cmd.AddCommand(newConfigSetCommand(o))
	cmd.AddCommand(newConfigUnsetCommand(o))
	cmd.AddCommand(newConfigViewCommand(o))

	return cmd
}

func (o *RootOptions) store(cmd *cobra.Command) (*configstore.Store, error) {
	path, err := configPath(cmd)
	if err != nil {
		return nil, err
	}
	return configstore.New(path), nil
}

func newConfigGetCommand(o *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "get SETTING",
		Short:   "Print the value of a setting",
		Example: templates.Examples(`cakectl config get rps`),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := o.store(cmd)
			if err != nil {
				return err
			}

			v, err := store.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(o.Out, v)
			return nil
		},
	}
}

func newConfigSetCommand(o *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "set SETTING VALUE",
		Short: "Change the value of a setting",
		Example: templates.Examples(`
      cakectl config set rps 8
      cakectl config set min_backoff 500ms`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := o.store(cmd)
			if err != nil {
				return err
			}
			return store.Set(args[0], args[1])
		},
	}
}

func newConfigUnsetCommand(o *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "unset SETTING",
		Short:   "Remove a setting, restoring its default",
		Example: templates.Examples(`cakectl config unset rps`),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := o.store(cmd)
			if err != nil {
				return err
			}
			return store.Unset(args[0])
		},
	}
}

func newConfigViewCommand(o *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show the settings in the config file",
		Long: templates.LongDesc(`
      Show the settings in the config file. The API token is partly hidden; use
      "config get api_token" to print it in full.`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := o.store(cmd)
			if err != nil {
				return err
			}

			values, err := store.All()
			if err != nil {
				return err
			}
			if len(values) == 0 {
				if _, err := os.Stat(store.Path()); errors.Is(err, os.ErrNotExist) {
					fmt.Fprintf(o.ErrOut, "No config file at %s\n", store.Path())
				}
				return nil
			}

			for _, key := range configstore.Keys() {
				v, ok := values[key]
				if !ok {
					continue
				}
				if configstore.Secret(key) {
					v = configstore.Mask(fmt.Sprint(v))
				}
				fmt.Fprintf(o.Out, "%s: %v\n", key, v)
			}
			return nil
		},
	}
}
