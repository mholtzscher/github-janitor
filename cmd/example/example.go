// Package example provides the example subcommand.
package example

import (
	"context"
	"fmt"

	"github.com/mholtzscher/github-janitor/internal/cli"
	ufcli "github.com/urfave/cli/v3"
)

// NewCommand creates the example command.
func NewCommand() *ufcli.Command {
	return &ufcli.Command{
		Name:  "example",
		Usage: "An example subcommand",
		Flags: []ufcli.Flag{
			&ufcli.StringFlag{
				Name:  "message",
				Value: "Hello, World!",
				Usage: "Message to print",
			},
		},
		Action: func(ctx context.Context, cmd *ufcli.Command) error {
			opts := cli.GlobalOptionsFromContext(ctx)

			if opts.Verbose {
				fmt.Println("Running example command in verbose mode")
			}

			message := cmd.String("message")
			fmt.Println(message)

			return nil
		},
	}
}
