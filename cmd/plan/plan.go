// Package plan provides the plan subcommand.
package plan

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	ufcli "github.com/urfave/cli/v3"

	cliutil "github.com/mholtzscher/github-janitor/cmd/common"
	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/github"
	reposync "github.com/mholtzscher/github-janitor/internal/sync"
)

// NewCommand creates the plan command (dry-run mode).
func NewCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:   "plan",
		Usage:  "Preview what changes would be made (dry-run mode)",
		Action: runPlan,
	}
}

func runPlan(ctx context.Context, cmd *ufcli.Command) error {
	configPath := cmd.String(cliutil.FlagConfig)
	token := cmd.String(cliutil.FlagToken)

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
	if authErr := client.ValidateAuth(); authErr != nil {
		return authErr
	}

	user, err := client.GetAuthenticatedUser()
	if err != nil {
		return err
	}
	fmt.Printf( //nolint:forbidigo // CLI output
		"Authenticated as: %s (token from: %s)\n\n",
		cliutil.Cyan(user),
		cliutil.Cyan(client.TokenSource),
	)

	// Create syncer
	syncer := reposync.NewSyncer(client, cfg)

	mode := cliutil.Yellow("DRY-RUN (preview only)")
	modeColor := cliutil.Yellow

	fmt.Printf("Mode: %s\n", mode)                                       //nolint:forbidigo // CLI output
	fmt.Printf("Repositories: %s\n\n", modeColor(len(cfg.Repositories))) //nolint:forbidigo // CLI output

	// Execute apply in dry-run mode
	results, err := syncer.SyncAll(ctx, true)
	if err != nil {
		return fmt.Errorf("plan failed: %w", err)
	}

	// Print results
	printResults(results)

	return nil
}

func printResults(results []reposync.Result) {
	fmt.Println("\n" + cliutil.BoldWhite(cliutil.Repeat("=", cliutil.SeparatorWidth))) //nolint:forbidigo // CLI output
	fmt.Println(cliutil.BoldWhite("PLAN RESULTS"))                                     //nolint:forbidigo // CLI output
	fmt.Println(cliutil.BoldWhite(cliutil.Repeat("=", cliutil.SeparatorWidth)))        //nolint:forbidigo // CLI output

	for _, result := range results {
		status := cliutil.Green("✓")
		if result.Error != nil {
			status = cliutil.Red("✗")
		}

		fmt.Printf("\n%s %s\n", status, result.Repository) //nolint:forbidigo // CLI output

		if result.Error != nil {
			fmt.Printf("   %s: %s\n", cliutil.Red("Error"), result.Error) //nolint:forbidigo // CLI output
			continue
		}

		if !result.Exists {
			fmt.Println("   " + cliutil.Yellow("Skipped: repository does not exist")) //nolint:forbidigo // CLI output
			continue
		}

		for _, change := range result.Changes {
			if strings.HasPrefix(change.Field, "actions_secret.") {
				fmt.Printf( //nolint:forbidigo // CLI output
					"   %s: %s (%v)\n",
					cliutil.Cyan(change.Field),
					cliutil.Yellow("write-only on GitHub; diff unavailable"),
					change.Desired,
				)
				continue
			}

			arrow := cliutil.Yellow("→ ")
			if reflect.DeepEqual(change.Current, change.Desired) {
				arrow = "="
			}
			fmt.Printf( //nolint:forbidigo // CLI output
				"   %s: %v %s %v\n",
				cliutil.Cyan(change.Field),
				change.Current,
				arrow,
				change.Desired,
			)
		}
	}

	fmt.Println("\n" + cliutil.BoldWhite(cliutil.Repeat("=", cliutil.SeparatorWidth))) //nolint:forbidigo // CLI output
}
