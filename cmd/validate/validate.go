// Package validate provides the validate subcommand.
package validate

import (
	"context"
	"fmt"

	ufcli "github.com/urfave/cli/v3"

	"github.com/mholtzscher/github-janitor/cmd/common"
	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/github"
)

// NewCommand creates the validate command.
func NewCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "validate",
		Usage: "Validate configuration file and authentication",
		Action: func(_ context.Context, cmd *ufcli.Command) error {
			return runValidate(cmd)
		},
	}
}

func runValidate(cmd *ufcli.Command) error {
	configPath := cmd.String(common.FlagConfig)
	token := cmd.String(common.FlagToken)

	fmt.Println(common.Cyan("Validating configuration...")) //nolint:forbidigo // CLI output

	// Load and validate config
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	fmt.Printf( //nolint:forbidigo // CLI output
		"Configuration valid: %s repositories configured\n",
		common.Green(len(cfg.Repositories)),
	)

	// Validate authentication
	fmt.Println("\n" + common.Cyan("Validating GitHub authentication...")) //nolint:forbidigo // CLI output
	client, err := github.NewClient(token)
	if err != nil {
		return fmt.Errorf("authentication error: %w", err)
	}

	if authErr := client.ValidateAuth(); authErr != nil {
		return authErr
	}

	user, err := client.GetAuthenticatedUser()
	if err != nil {
		return err
	}

	fmt.Printf( //nolint:forbidigo // CLI output
		"Authentication valid: authenticated as %s (token from: %s)\n",
		common.Cyan(user),
		common.Cyan(client.TokenSource),
	)
	fmt.Println("\n" + common.Green("All validations passed!")) //nolint:forbidigo // CLI output

	return nil
}
