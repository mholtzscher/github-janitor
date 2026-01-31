// Package cmd implements the CLI commands for github-janitor.
package cmd

import (
	"context"

	ufcli "github.com/urfave/cli/v3"
	"github.com/mholtzscher/github-janitor/cmd/example"
	"github.com/mholtzscher/github-janitor/internal/cli"
)

// Version is set at build time.
var Version = "0.1.0" // x-release-please-version

// Run is the entry point for the CLI.
func Run(ctx context.Context, args []string) error {
	app := &ufcli.Command{
		Name:    "github-janitor",
		Usage:   "A Go CLI tool built with Nix",
		Version: Version,
		Flags: []ufcli.Flag{
			&ufcli.BoolFlag{
				Name:  cli.FlagVerbose,
				Usage: "Print verbose output",
			},
			&ufcli.BoolFlag{
				Name:  cli.FlagNoColor,
				Usage: "Disable colored output",
			},
		},
		Commands: []*ufcli.Command{
			example.NewCommand(),
		},
	}

	return app.Run(ctx, args)
}
