package secrets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mholtzscher/github-janitor/internal/config"
)

// Source is the interface for secret value resolution.
type Source interface {
	// Resolve returns the secret value.
	Resolve(ctx context.Context) (string, error)
	// Describe returns a safe description of the source (no values).
	Describe() string
}

// EnvVarSource reads a secret from an environment variable.
type EnvVarSource struct {
	Name string
}

// Resolve reads the environment variable value.
func (s *EnvVarSource) Resolve(_ context.Context) (string, error) {
	value, ok := os.LookupEnv(s.Name)
	if !ok {
		return "", fmt.Errorf("environment variable %s not set", s.Name)
	}

	if value == "" {
		return "", fmt.Errorf("environment variable %s is empty", s.Name)
	}

	return strings.TrimSpace(value), nil
}

// Describe returns a safe description of the source.
func (s *EnvVarSource) Describe() string {
	return fmt.Sprintf("env:%s", s.Name)
}

// CommandSource executes a command and captures stdout as the secret value.
type CommandSource struct {
	Argv []string
}

// Resolve executes the command and returns stdout (trimmed of trailing newline).
func (s *CommandSource) Resolve(ctx context.Context) (string, error) {
	if len(s.Argv) == 0 {
		return "", errors.New("command is empty")
	}

	//nolint:gosec // CommandSource.Argv is user-configured; validation happens at config load time
	cmd := exec.CommandContext(ctx, s.Argv[0], s.Argv[1:]...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil // Discard stderr to avoid polluting output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command %s failed: %w", s.Argv[0], err)
	}

	value := strings.TrimSpace(stdout.String())
	if value == "" {
		return "", fmt.Errorf("command %s returned empty value", s.Argv[0])
	}

	return value, nil
}

// Describe returns a safe description of the source (executable name only).
func (s *CommandSource) Describe() string {
	if len(s.Argv) == 0 {
		return "command:"
	}

	return fmt.Sprintf("command:%s", s.Argv[0])
}

// Resolver resolves secrets from configuration.
type Resolver struct {
	secrets map[string]Source
}

// NewResolver creates a new resolver from config secrets.
func NewResolver(configSecrets []config.ActionsSecret) *Resolver {
	sources := make(map[string]Source, len(configSecrets))

	for _, secret := range configSecrets {
		if secret.Env != nil && *secret.Env != "" {
			sources[secret.Name] = &EnvVarSource{Name: *secret.Env}
		} else if len(secret.Command) > 0 {
			sources[secret.Name] = &CommandSource{Argv: secret.Command}
		}
	}

	return &Resolver{secrets: sources}
}

// ResolveAll resolves all configured secrets.
// Returns a map of secret name to value, or error if any resolution fails.
func (r *Resolver) ResolveAll(ctx context.Context) (map[string]string, error) {
	values := make(map[string]string, len(r.secrets))

	for name, source := range r.secrets {
		value, err := source.Resolve(ctx)
		if err != nil {
			return nil, fmt.Errorf("secret %s (%s): %w", name, source.Describe(), err)
		}

		values[name] = value
	}

	return values, nil
}

// GetSource returns the source for a secret name (for dry-run descriptions).
func (r *Resolver) GetSource(name string) (Source, bool) {
	source, ok := r.secrets[name]
	return source, ok
}
