// Package common provides shared utilities for CLI commands.
package common

import (
	"strings"

	"github.com/fatih/color"
)

// SeparatorWidth is the width of the separator line in CLI output.
const SeparatorWidth = 60

// Color functions.
var (
	BoldWhite = color.New(color.FgWhite, color.Bold).SprintFunc() //nolint:gochecknoglobals // CLI color helper
	Cyan      = color.New(color.FgCyan).SprintFunc()              //nolint:gochecknoglobals // CLI color helper
	Green     = color.New(color.FgGreen).SprintFunc()             //nolint:gochecknoglobals // CLI color helper
	Red       = color.New(color.FgRed).SprintFunc()               //nolint:gochecknoglobals // CLI color helper
	Yellow    = color.New(color.FgYellow).SprintFunc()            //nolint:gochecknoglobals // CLI color helper
)

// Repeat returns a new string consisting of count copies of the string s.
func Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}
