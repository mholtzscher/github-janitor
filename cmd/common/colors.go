// Package common provides shared utilities for CLI commands.
package common

import (
	"strings"

	"github.com/fatih/color"
)

// Color functions
var (
	BoldWhite = color.New(color.FgWhite, color.Bold).SprintFunc()
	Cyan      = color.New(color.FgCyan).SprintFunc()
	Green     = color.New(color.FgGreen).SprintFunc()
	Red       = color.New(color.FgRed).SprintFunc()
	Yellow    = color.New(color.FgYellow).SprintFunc()
)

// Repeat returns a new string consisting of count copies of the string s.
func Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}
