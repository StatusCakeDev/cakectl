// Package cmd wires the application together behind a command line interface.
// The root command creates the configuration store and the API adapter, and
// passes them to a command for each resource. Each command accepts the user's
// flags or a resource file, calls the adapter and renders the result.
package cmd
