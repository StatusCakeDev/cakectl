package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/tomasbasham/cli-runtime/iooption"
	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/StatusCakeDev/cakectl/internal/adapters"
	"github.com/StatusCakeDev/cakectl/internal/spec"
	"github.com/StatusCakeDev/cakectl/internal/view"
)

// resource describes one kind of StatusCake resource to the generic command
// builder: how to create, update, delete and describe it, and the flags that
// override its spec.
type resource[S any] struct {
	// use is the command name, which also names the resource in output.
	use     string
	aliases []string
	// noun is how the resource is named in prose, such as "uptime check".
	noun  string
	short string
	// flags registers the flags that override the spec. They differ between
	// creating and updating: some fields cannot be changed once set.
	flags func(fs *pflag.FlagSet, create bool) *binder[S]
	// createExample is a typical invocation, shown in the help.
	createExample string

	create   func(context.Context, *adapters.Client, S) (string, error)
	update   func(context.Context, *adapters.Client, string, S) error
	remove   func(context.Context, *adapters.Client, string) error
	describe func(context.Context, *adapters.Client, string) (any, error)
}

// newResourceCommand builds the command group for a resource, with the
// create, update, delete and describe subcommands every resource has. The
// caller adds the listing and anything specific to the resource.
// withArticle puts "a" or "an" before a noun.
func withArticle(noun string) string {
	if strings.ContainsRune("aeiouAEIOU", rune(noun[0])) || strings.HasPrefix(noun, "SSL") {
		return "an " + noun
	}
	return "a " + noun
}

func newResourceCommand[S any](o *RootOptions, r resource[S]) *cobra.Command {
	cmd := &cobra.Command{
		Use:     r.use,
		Aliases: r.aliases,
		Short:   r.short,
	}

	cmd.AddCommand(newCreateCommand(o, r))
	cmd.AddCommand(newUpdateCommand(o, r))
	cmd.AddCommand(newDeleteCommand(o, r))
	cmd.AddCommand(newDescribeCommand(o, r))

	return cmd
}

func newCreateCommand[S any](o *RootOptions, r resource[S]) *cobra.Command {
	var filename string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create " + withArticle(r.noun),
		Long: templates.LongDesc(fmt.Sprintf(`
      Create %[1]s.

      The configuration comes from a YAML or JSON file given with --filename
      (use - to read standard input), from flags, or from both. When both are
      given, a flag overrides the same field in the file.`, withArticle(r.noun))),
		Example: templates.Examples(fmt.Sprintf(`
      # Create %[1]s from flags
      %[2]s

      # Create it from a file
      cakectl %[3]s create --filename %[3]s.yaml

      # Create it from a file, overriding one field
      cakectl %[3]s create --filename %[3]s.yaml --name "Staging"`, withArticle(r.noun), r.createExample, r.use)),
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVar(&filename, "filename", "", "File containing the specification (YAML or JSON; - for standard input)")
	b := r.flags(cmd.Flags(), true)

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		s, err := readSpec[S](o, filename, b)
		if err != nil {
			return err
		}

		api, err := o.client(cmd)
		if err != nil {
			return err
		}

		id, err := r.create(commandContext(cmd), api, s)
		if err != nil {
			return err
		}
		return o.print(view.Created(r.use, id))
	}
	return cmd
}

func newUpdateCommand[S any](o *RootOptions, r resource[S]) *cobra.Command {
	var filename string

	cmd := &cobra.Command{
		Use:   "update ID",
		Short: "Update " + withArticle(r.noun),
		Long: templates.LongDesc(fmt.Sprintf(`
      Update %[1]s.

      Only the fields you supply change; everything else is left as it is. The
      new values come from a file given with --filename, from flags, or from
      both, where a flag overrides the same field in the file.

      The output of "describe --format yaml" can be edited and given back with
      --filename.`, withArticle(r.noun))),
		Example: templates.Examples(fmt.Sprintf(`
      # Change one field
      cakectl %[1]s update 123 --name "Renamed"

      # Edit a resource as a file
      cakectl %[1]s describe 123 --format yaml > %[1]s.yaml
      cakectl %[1]s update 123 --filename %[1]s.yaml`, r.use)),
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&filename, "filename", "", "File containing the new values (YAML or JSON; - for standard input)")
	b := r.flags(cmd.Flags(), false)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		s, err := readSpec[S](o, filename, b)
		if err != nil {
			return err
		}

		api, err := o.client(cmd)
		if err != nil {
			return err
		}

		if err := r.update(commandContext(cmd), api, args[0], s); err != nil {
			if errors.Is(err, adapters.ErrNothingToUpdate) {
				return fmt.Errorf("%w; give --filename or a flag such as --name", err)
			}
			return err
		}
		return o.print(view.Updated(r.use, args[0]))
	}
	return cmd
}

func newDeleteCommand[S any](o *RootOptions, r resource[S]) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete " + withArticle(r.noun),
		Long: templates.LongDesc(fmt.Sprintf(`
      Delete %s. This cannot be undone.

      You are asked to confirm first. When input is not a terminal, as in a
      script, the command refuses to run unless --yes is given.`, withArticle(r.noun))),
		Example: templates.Examples(fmt.Sprintf(`
      # Delete, after confirming
      cakectl %[1]s delete 123

      # Delete without asking
      cakectl %[1]s delete 123 --yes`, r.use)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.confirm(fmt.Sprintf("Delete %s %s?", r.noun, args[0]), yes); err != nil {
				return err
			}

			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			if err := r.remove(commandContext(cmd), api, args[0]); err != nil {
				return err
			}
			return o.print(view.Deleted(r.use, args[0]))
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Do not ask for confirmation")
	return cmd
}

func newDescribeCommand[S any](o *RootOptions, r resource[S]) *cobra.Command {
	return &cobra.Command{
		Use:     "describe ID",
		Aliases: []string{"get"},
		Short:   "Show the details of " + withArticle(r.noun),
		Long: templates.LongDesc(fmt.Sprintf(`
      Show everything about %s.

      With --format yaml or --format json the output has the same shape that --filename
      reads, so it can be edited and applied back with the update command.`, withArticle(r.noun))),
		Example: templates.Examples(fmt.Sprintf(`
      cakectl %[1]s describe 123
      cakectl %[1]s describe 123 --format yaml`, r.use)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			api, err := o.client(cmd)
			if err != nil {
				return err
			}

			v, err := r.describe(commandContext(cmd), api, args[0])
			if err != nil {
				return err
			}
			return o.print(v)
		},
	}
}

// readSpec builds the spec a command acts on: the file's content, if a file
// was named, with the flags that were passed applied over it.
func readSpec[S any](o *RootOptions, filename string, b *binder[S]) (S, error) {
	var s S

	if filename != "" {
		var err error
		if s, err = decodeFile[S](o, filename); err != nil {
			return s, err
		}
	}

	if err := b.apply(&s); err != nil {
		return s, err
	}
	return s, nil
}

// decodeFile reads a spec from the named file, or from standard input for "-".
func decodeFile[S any](o *RootOptions, filename string) (S, error) {
	var in = o.In
	if filename != "-" {
		f, err := iooption.OpenFile(o.fsys, filename)
		if err != nil {
			var zero S
			return zero, fmt.Errorf("open %s: %w", filename, err)
		}
		defer f.Close()
		in = f
	}

	s, err := spec.Decode[S](in)
	if err != nil {
		return s, fmt.Errorf("%s: %w", filename, err)
	}
	return s, nil
}
