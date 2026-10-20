package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/StatusCakeDev/cakectl/internal/configstore"
)

func newAuthCommand(o *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}

	cmd.AddCommand(newLoginCommand(o))

	return cmd
}

func newLoginCommand(o *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Save the API token used to talk to StatusCake",
		Long: templates.LongDesc(`
      Save an API token in the config file so later commands can use it.

      The token comes from --api-token, from the STATUSCAKE_API_TOKEN
      environment variable, or from standard input. The config file is only
      readable by you.`),
		Example: templates.Examples(`
      cakectl auth login --api-token "$TOKEN"
      echo "$TOKEN" | cakectl auth login`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := o.token(cmd)
			if err != nil {
				return err
			}

			path, err := configPath(cmd)
			if err != nil {
				return err
			}

			store := configstore.New(path)
			if err := store.Set("api_token", token); err != nil {
				return err
			}

			fmt.Fprintf(o.ErrOut, "Saved API token to %s\n", path)
			return nil
		},
	}
}

// token finds the token to save: the flag or environment first, then a line on
// standard input.
func (o *RootOptions) token(cmd *cobra.Command) (string, error) {
	if token, _ := cmd.Flags().GetString("api-token"); token != "" {
		return token, nil
	}

	if o.isTerminal(o.In) {
		fmt.Fprint(o.ErrOut, "API token: ")
	}

	line, err := bufio.NewReader(o.In).ReadString('\n')
	if token := strings.TrimSpace(line); token != "" {
		return token, nil
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read API token: %w", err)
	}
	return "", errors.New("an API token is required: pass --api-token or provide it on standard input")
}
