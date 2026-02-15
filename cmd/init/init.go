// Package initcmd provides the init subcommand.
package initcmd

import (
	"context"
	"fmt"
	"os"

	ufcli "github.com/urfave/cli/v3"

	"github.com/mholtzscher/github-janitor/cmd/common"
	"github.com/mholtzscher/github-janitor/internal/config"
)

// NewCommand creates the init command.
func NewCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "init",
		Usage: "Generate an example configuration file",
		Action: func(_ context.Context, cmd *ufcli.Command) error {
			return runInit(cmd)
		},
	}
}

func runInit(cmd *ufcli.Command) error {
	configPath := cmd.String(common.FlagConfig)

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	// Write example config
	exampleConfig := config.ExampleConfig()
	if err := os.WriteFile(configPath, []byte(exampleConfig), config.DefaultFileMode); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Printf("Created example configuration file: %s\n", common.Green(configPath)) //nolint:forbidigo // CLI output
	fmt.Println("\n" + common.BoldWhite("Next steps:"))                              //nolint:forbidigo // CLI output
	fmt.Println("1. Edit the configuration file to add your repositories")           //nolint:forbidigo // CLI output
	fmt.Println("2. Run 'github-janitor validate' to verify your setup")             //nolint:forbidigo // CLI output
	fmt.Println("3. Run 'github-janitor plan' to preview changes")                   //nolint:forbidigo // CLI output
	fmt.Println("4. Run 'github-janitor sync' to apply changes")                     //nolint:forbidigo // CLI output

	return nil
}
