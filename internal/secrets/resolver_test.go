package secrets_test

import (
	"context"
	"testing"

	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/secrets"
)

func TestEnvVarSource_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		envName   string
		envValue  string
		setEnv    bool
		wantValue string
		wantErr   bool
	}{
		{
			name:      "success",
			envName:   "TEST_SECRET",
			envValue:  "secret-value",
			setEnv:    true,
			wantValue: "secret-value",
			wantErr:   false,
		},
		{
			name:    "not set",
			envName: "TEST_SECRET_NOT_SET",
			setEnv:  false,
			wantErr: true,
		},
		{
			name:     "empty value",
			envName:  "TEST_SECRET_EMPTY",
			envValue: "",
			setEnv:   true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envName, tt.envValue)
			}

			source := &secrets.EnvVarSource{Name: tt.envName}
			got, err := source.Resolve(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("Resolve() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestEnvVarSource_Describe(t *testing.T) {
	source := &secrets.EnvVarSource{Name: "MY_SECRET"}
	want := "env:MY_SECRET"

	if got := source.Describe(); got != want {
		t.Errorf("Describe() = %v, want %v", got, want)
	}
}

func TestCommandSource_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		argv      []string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "echo success",
			argv:      []string{"echo", "hello"},
			wantValue: "hello",
			wantErr:   false,
		},
		{
			name:    "empty command",
			argv:    []string{},
			wantErr: true,
		},
		{
			name:    "command not found",
			argv:    []string{"/nonexistent/command"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &secrets.CommandSource{Argv: tt.argv}
			got, err := source.Resolve(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("Resolve() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestCommandSource_Resolve_TrimsNewline(t *testing.T) {
	source := &secrets.CommandSource{Argv: []string{"echo", "hello"}}
	got, err := source.Resolve(context.Background())

	if err != nil {
		t.Errorf("Resolve() unexpected error = %v", err)
		return
	}

	// echo adds a newline, which should be trimmed
	want := "hello"
	if got != want {
		t.Errorf("Resolve() = %q, want %q", got, want)
	}
}

func TestCommandSource_Describe(t *testing.T) {
	tests := []struct {
		name string
		argv []string
		want string
	}{
		{
			name: "with args",
			argv: []string{"op", "read", "op://vault/item/field"},
			want: "command:op",
		},
		{
			name: "empty",
			argv: []string{},
			want: "command:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &secrets.CommandSource{Argv: tt.argv}
			if got := source.Describe(); got != tt.want {
				t.Errorf("Describe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewResolver(t *testing.T) {
	envName := "TEST_VAR"
	configSecrets := []config.ActionsSecret{
		{
			Name: "SECRET1",
			Env:  &envName,
		},
		{
			Name:    "SECRET2",
			Command: []string{"echo", "test"},
		},
	}

	resolver := secrets.NewResolver(configSecrets)

	// Test GetSource for env secret
	source1, ok := resolver.GetSource("SECRET1")
	if !ok {
		t.Error("expected to find SECRET1")
	}
	if source1.Describe() != "env:TEST_VAR" {
		t.Errorf("expected env source, got %s", source1.Describe())
	}

	// Test GetSource for command secret
	source2, ok := resolver.GetSource("SECRET2")
	if !ok {
		t.Error("expected to find SECRET2")
	}
	if source2.Describe() != "command:echo" {
		t.Errorf("expected command source, got %s", source2.Describe())
	}

	// Test GetSource for missing secret
	_, ok = resolver.GetSource("MISSING")
	if ok {
		t.Error("expected not to find MISSING")
	}
}

func TestResolver_ResolveAll(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		wantErr  bool
	}{
		{
			name:     "success",
			envName:  "RESOLVER_TEST_VAR",
			envValue: "test-value",
			wantErr:  false,
		},
		{
			name:    "missing env var",
			envName: "RESOLVER_TEST_VAR_MISSING",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envName, tt.envValue)
			}

			configSecrets := []config.ActionsSecret{
				{
					Name: "TEST_SECRET",
					Env:  &tt.envName,
				},
			}

			resolver := secrets.NewResolver(configSecrets)
			values, err := resolver.ResolveAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if values["TEST_SECRET"] != tt.envValue {
					t.Errorf("ResolveAll() = %v, want %v", values["TEST_SECRET"], tt.envValue)
				}
			}
		})
	}
}
