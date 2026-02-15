// Package cmd implements the CLI commands for github-janitor.
package cmd

import (
	"context"

	"github.com/fatih/color"
	ufcli "github.com/urfave/cli/v3"

	"github.com/mholtzscher/github-janitor/cmd/common"
	initcmd "github.com/mholtzscher/github-janitor/cmd/init"
	"github.com/mholtzscher/github-janitor/cmd/plan"
	"github.com/mholtzscher/github-janitor/cmd/sync"
	"github.com/mholtzscher/github-janitor/cmd/validate"
	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/github"
)

// Version is set at build time.
var Version = "0.1.3" //nolint:gochecknoglobals // Version is set at build time

// Run is the entry point for the CLI.
func Run(ctx context.Context, args []string) error {
	app := &ufcli.Command{
		Name:    "github-janitor",
		Usage:   "Synchronize GitHub repository settings across multiple repos",
		Version: Version,
		Before: func(ctx context.Context, cmd *ufcli.Command) (context.Context, error) {
			if cmd.Bool(common.FlagNoColor) {
				color.NoColor = true //nolint:reassign // Setting color.NoColor is the intended way to disable colors
			}
			return ctx, nil
		},
		Flags: []ufcli.Flag{
			&ufcli.BoolFlag{
				Name:  common.FlagNoColor,
				Usage: "Disable colored output",
			},
			&ufcli.StringFlag{
				Name:    common.FlagConfig,
				Aliases: []string{"c"},
				Value:   config.DefaultFilename,
				Usage:   "Path to configuration file",
				Sources: ufcli.EnvVars("GITHUB_JANITOR_CONFIG"),
			},
			&ufcli.StringFlag{
				Name:    common.FlagToken,
				Aliases: []string{"t"},
				Usage:   "GitHub personal access token (overrides auto-detection)",
				Sources: ufcli.EnvVars(github.EnvToken),
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
