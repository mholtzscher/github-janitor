package config //nolint:testpackage // Tests internal implementation details

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

func TestValidate_SecurityDependabotUpdatesRequireAlerts(t *testing.T) {
	t.Run("security_updates_true_requires_alerts_true", func(t *testing.T) {
		cfg := &Config{
			Repositories: []Repository{{Owner: "o", Name: "r"}},
			Settings: Settings{
				Security: &SecuritySettings{
					DependabotSecurityUpdates: boolPtr(true),
				},
			},
		}

		if err := cfg.Validate(); err == nil {
			t.Fatal("Validate() = nil; want error")
		}
	})

	t.Run("security_updates_true_allows_alerts_true", func(t *testing.T) {
		cfg := &Config{
			Repositories: []Repository{{Owner: "o", Name: "r"}},
			Settings: Settings{
				Security: &SecuritySettings{
					DependabotAlerts:          boolPtr(true),
					DependabotSecurityUpdates: boolPtr(true),
				},
			},
		}

		if err := cfg.Validate(); err != nil {
			t.Fatalf("Validate() error = %v; want nil", err)
		}
	})
}

func boolPtr(v bool) *bool {
	return &v
}
