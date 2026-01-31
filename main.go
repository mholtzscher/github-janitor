// github-janitor is a CLI tool.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mholtzscher/github-janitor/cmd"
)

func main() {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
