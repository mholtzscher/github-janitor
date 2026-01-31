package config

import "testing"

func TestValidate_BranchProtectionPatternRequired(t *testing.T) {
	t.Run("disabled_allows_missing_pattern", func(t *testing.T) {
		cfg := &Config{
			Repositories: []Repository{{Owner: "o", Name: "r"}},
			Settings: Settings{
				BranchProtection: &BranchProtection{Enabled: false, Pattern: ""},
			},
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("Validate() error = %v; want nil", err)
		}
	})

	t.Run("enabled_requires_pattern", func(t *testing.T) {
		cfg := &Config{
			Repositories: []Repository{{Owner: "o", Name: "r"}},
			Settings: Settings{
				BranchProtection: &BranchProtection{Enabled: true, Pattern: ""},
			},
		}
		if err := cfg.Validate(); err == nil {
			t.Fatal("Validate() = nil; want error")
		}
	})
}
