// Package cli provides shared CLI options and utilities.
package cli

import (
	"context"
)

// Flag names
const (
	FlagVerbose = "verbose"
	FlagNoColor = "no-color"
)

// GlobalOptions holds CLI flags that are shared across commands.
type GlobalOptions struct {
	Verbose bool
	NoColor bool
}

// GlobalOptionsFromContext extracts global options from the CLI context.
func GlobalOptionsFromContext(ctx context.Context) GlobalOptions {
	// In a real implementation, you'd extract these from the context
	// For now, return defaults
	return GlobalOptions{
		Verbose: false,
		NoColor: false,
	}
}
