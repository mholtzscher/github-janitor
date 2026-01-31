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

type globalOptionsKey struct{}

// WithGlobalOptions attaches global CLI options to the context.
func WithGlobalOptions(ctx context.Context, opts GlobalOptions) context.Context {
	return context.WithValue(ctx, globalOptionsKey{}, opts)
}

// GlobalOptionsFromContext extracts global options from the CLI context.
func GlobalOptionsFromContext(ctx context.Context) GlobalOptions {
	if v := ctx.Value(globalOptionsKey{}); v != nil {
		if opts, ok := v.(GlobalOptions); ok {
			return opts
		}
	}
	return GlobalOptions{}
}
