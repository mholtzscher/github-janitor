// Package sync provides the sync subcommand.
package sync

import (
	"context"
	"fmt"
	"reflect"

	"github.com/mholtzscher/github-janitor/cmd/common"
	"github.com/mholtzscher/github-janitor/internal/cli"
	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/github"
	"github.com/mholtzscher/github-janitor/internal/sync"
	ufcli "github.com/urfave/cli/v3"
)

// NewCommand creates the sync command.
func NewCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "sync",
		Usage: "Apply settings to all configured repositories",
		Flags: []ufcli.Flag{
			&ufcli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview changes without applying them",
			},
		},
		Action: func(ctx context.Context, cmd *ufcli.Command) error {
			return runSync(ctx, cmd, cmd.Bool("dry-run"))
		},
	}
}

func runSync(ctx context.Context, cmd *ufcli.Command, dryRun bool) error {
	configPath := cmd.String("config")
	token := cmd.String("token")

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create GitHub client
	client, err := github.NewClient(token)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Validate authentication
	if err := client.ValidateAuth(); err != nil {
		return err
	}

	user, err := client.GetAuthenticatedUser()
	if err != nil {
		return err
	}
	fmt.Printf("Authenticated as: %s\n\n", common.Cyan(user))

	// Create syncer
	syncer := sync.NewSyncer(client, cfg)

	mode := common.BoldWhite("APPLYING")
	modeColor := common.Cyan
	if dryRun {
		mode = common.Yellow("DRY-RUN (preview only)")
		modeColor = common.Yellow
	}
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Repositories: %s\n\n", modeColor(len(cfg.Repositories)))

	// Execute sync
	results, err := syncer.SyncAll(dryRun)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Print results
	printResults(results, cli.GlobalOptionsFromContext(ctx))

	return nil
}

func printResults(results []sync.Result, opts cli.GlobalOptions) {
	fmt.Println("\n" + common.BoldWhite(common.Repeat("=", 60)))
	fmt.Println(common.BoldWhite("SYNC RESULTS"))
	fmt.Println(common.BoldWhite(common.Repeat("=", 60)))

	for _, result := range results {
		status := common.Green("✓")
		if result.Error != nil {
			status = common.Red("✗")
		}

		fmt.Printf("\n%s %s\n", status, result.Repository)

		if result.Error != nil {
			fmt.Printf("   %s: %s\n", common.Red("Error"), result.Error)
			continue
		}

		if !result.Exists {
			fmt.Println("   " + common.Yellow("Skipped: repository does not exist"))
			continue
		}

		for _, change := range result.Changes {
			arrow := common.Yellow("→")
			if reflect.DeepEqual(change.Current, change.Desired) {
				arrow = "="
			}
			fmt.Printf("   %s: %v %s %v\n", common.Cyan(change.Field), change.Current, arrow, change.Desired)
		}
	}

	fmt.Println("\n" + common.BoldWhite(common.Repeat("=", 60)))
}
