package cmd

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cliruntime "github.com/tomasbasham/cli-runtime"
	cliflag "github.com/tomasbasham/cli-runtime/flag"
	"github.com/tomasbasham/cli-runtime/iooption"
	"github.com/tomasbasham/cli-runtime/printer"
	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/StatusCakeDev/cakectl/internal/adapters"
	"github.com/StatusCakeDev/cakectl/internal/cmd/flags"
	"github.com/StatusCakeDev/cakectl/internal/factories"
)

var (
	rootLong = templates.LongDesc(`
    cakectl creates, inspects and manages StatusCake monitoring from the
    terminal.

    Resources are described with flags, with a YAML or JSON file, or both: a
    file supplies the whole specification and any flag overrides the file.
    Every command prints plain text by default; use --format to print JSON or
    YAML for scripts.`)

	// Build information, set at build time via ldflags.
	version = "dev"
	commit  = ""
)

// RootOptions defines the options for the `cakectl` command, and is what the
// subcommands share.
type RootOptions struct {
	configFlags  *flags.ConfigFlags
	printerFlags *cliflag.PrinterFlags

	// api is created on first use, once configuration has been loaded, so that
	// commands which do not talk to StatusCake need no API token. Tests set it
	// to a client of their own.
	api *adapters.Client

	// fsys opens the files named on the command line.
	fsys fs.FS

	// isTerminal reports whether a stream is an interactive terminal. It is a
	// field so that tests can stand in for one.
	isTerminal func(io.Reader) bool

	iooption.IOStreams
}

// NewRootOptions provides an initialised [RootOptions] instance.
func NewRootOptions(streams iooption.IOStreams) *RootOptions {
	return &RootOptions{
		configFlags: &flags.ConfigFlags{},
		printerFlags: cliflag.NewPrinterFlags(
			cliflag.FormatTextFlag|cliflag.FormatJSONFlag|cliflag.FormatPrettyJSONFlag|cliflag.FormatYAMLFlag,
			cliflag.FormatText,
		),
		fsys:       osFS{},
		isTerminal: isTerminal,
		IOStreams:  streams,
	}
}

// NewRootCommand creates the `cakectl` command with default arguments.
func NewRootCommand() *cobra.Command {
	return NewRootCommandWithArgs(NewRootOptions(iooption.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}))
}

// NewRootCommandWithArgs creates the `cakectl` command.
func NewRootCommandWithArgs(o *RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "cakectl",
		Version:               versionInfo(),
		DisableFlagsInUseLine: true,
		Short:                 "Manage StatusCake monitoring from the command line",
		Long:                  rootLong,
		SilenceErrors:         true,
		SilenceUsage:          true,
	}

	cmd.SetOut(o.Out)
	cmd.SetErr(o.ErrOut)
	cmd.SetIn(o.In)

	printerOpts := printer.WarningPrinterOptions{Color: true}
	printer := printer.NewWarningPrinter(o.ErrOut, printerOpts)
	cmd.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc(printer))

	pflags := cmd.PersistentFlags()
	pflags.String("config", "", "Config file (default is $HOME/.config/statuscake/cakectl.yaml)")

	// Add root-specific flags.
	o.configFlags.AddFlags(pflags)
	o.printerFlags.AddFlags(pflags)

	groups := cliruntime.CommandGroups{
		{
			Title: "Monitoring Commands:",
			Commands: []*cobra.Command{
				newHeartbeatCommand(o),
				newPagespeedCommand(o),
				newSSLCommand(o),
				newUptimeCommand(o),
			},
		},
		{
			Title: "Alerting Commands:",
			Commands: []*cobra.Command{
				newContactGroupCommand(o),
				newMaintenanceWindowCommand(o),
			},
		},
		{
			Title:    "Reference Commands:",
			Commands: []*cobra.Command{newLocationsCommand(o)},
		},
	}
	groups.Add(cmd)

	cmd.AddCommand(newAuthCommand(o))
	cmd.AddCommand(newConfigCommand(o))

	// The global normalisation function ensures that all flags specified meet
	// the desired format, changing users' input if necessary.
	cmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc())

	return cmd
}

// client provides the API client, building it from configuration the first
// time it is needed.
func (o *RootOptions) client(cmd *cobra.Command) (*adapters.Client, error) {
	if o.api != nil {
		return o.api, nil
	}

	if err := o.loadConfig(cmd); err != nil {
		return nil, err
	}

	sdk, err := factories.CreateClient(factories.ClientConfig{
		APIToken:       o.configFlags.APIToken,
		RPS:            o.configFlags.RPS,
		Retries:        o.configFlags.Retries,
		MinBackoff:     o.configFlags.MinBackoff,
		MaxBackoff:     o.configFlags.MaxBackoff,
		CustomEndpoint: o.configFlags.CustomEndpoint,

		// These values will form part of the User-Agent header sent to StatusCake.
		// They will appear in its logs and help identify the source of requests.
		Version: version,
		Commit:  commit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create StatusCake client: %w", err)
	}

	o.api = adapters.NewClient(sdk)
	return o.api, nil
}

func (o *RootOptions) loadConfig(cmd *cobra.Command) error {
	configFile, _ := cmd.Flags().GetString("config")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("cakectl")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.config/statuscake")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.AutomaticEnv()
		} else {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(o.configFlags); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// configPath is the file that `auth` and `config` commands change: the one
// named by --config, or the default location.
func configPath(cmd *cobra.Command) (string, error) {
	if path, _ := cmd.Flags().GetString("config"); path != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find the config file: %w", err)
	}
	return filepath.Join(home, ".config", "statuscake", "cakectl.yaml"), nil
}

// print writes a result in the chosen output format.
func (o *RootOptions) print(v any) error {
	p, err := o.printerFlags.ToPrinter()
	if err != nil {
		return err
	}
	return p.Print(o.Out, v)
}

// context is the context a command runs in.
func commandContext(cmd *cobra.Command) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func versionInfo() string {
	if version == "" {
		return "dev"
	}
	return fmt.Sprintf("%s (commit: %s)", version, commit)
}
