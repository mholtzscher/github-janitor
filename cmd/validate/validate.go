// Package validate provides the validate subcommand.
package validate

import (
	"context"
	"fmt"

	"github.com/mholtzscher/github-janitor/cmd/common"
	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/github"
	ufcli "github.com/urfave/cli/v3"
)

// NewCommand creates the validate command.
func NewCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "validate",
		Usage: "Validate configuration file and authentication",
		Action: func(ctx context.Context, cmd *ufcli.Command) error {
			return runValidate(cmd)
		},
	}
}

func runValidate(cmd *ufcli.Command) error {
	configPath := cmd.String("config")
	token := cmd.String("token")

	fmt.Println(common.Cyan("Validating configuration..."))

	// Load and validate config
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	fmt.Printf("Configuration valid: %s repositories configured\n", common.Green(len(cfg.Repositories)))

	// Validate authentication
	fmt.Println("\n" + common.Cyan("Validating GitHub authentication..."))
	client, err := github.NewClient(token)
	if err != nil {
		return fmt.Errorf("authentication error: %w", err)
	}

	if err := client.ValidateAuth(); err != nil {
		return err
	}

	user, err := client.GetAuthenticatedUser()
	if err != nil {
		return err
	}

	fmt.Printf("Authentication valid: authenticated as %s\n", common.Cyan(user))
	fmt.Println("\n" + common.Green("All validations passed!"))

	return nil
}
