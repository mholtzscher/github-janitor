// Package cmd implements the CLI commands for github-janitor.
package cmd

import (
	"context"

	initcmd "github.com/mholtzscher/github-janitor/cmd/init"
	"github.com/mholtzscher/github-janitor/cmd/plan"
	"github.com/mholtzscher/github-janitor/cmd/sync"
	"github.com/mholtzscher/github-janitor/cmd/validate"
	"github.com/mholtzscher/github-janitor/internal/cli"
	ufcli "github.com/urfave/cli/v3"
)

// Version is set at build time.
var Version = "0.1.0" // x-release-please-version

// Run is the entry point for the CLI.
func Run(ctx context.Context, args []string) error {
	app := &ufcli.Command{
		Name:    "github-janitor",
		Usage:   "Synchronize GitHub repository settings across multiple repos",
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
			&ufcli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   ".github-janitor.yaml",
				Usage:   "Path to configuration file",
				Sources: ufcli.EnvVars("GITHUB_JANITOR_CONFIG"),
			},
			&ufcli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "GitHub personal access token (overrides auto-detection)",
				Sources: ufcli.EnvVars("GITHUB_TOKEN"),
			},
		},
		Commands: []*ufcli.Command{
			sync.NewCommand(),
			plan.NewCommand(),
			validate.NewCommand(),
			initcmd.NewCommand(),
		},
	}

	return app.Run(ctx, args)
}
