package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"

	"gopkg.in/yaml.v3"
)

const (
	DefaultFilename = "github-janitor.yaml"
	DefaultFileMode = 0644

	VisibilityPublic  = "public"
	VisibilityPrivate = "private"

	SquashTitlePRTitle         = "PR_TITLE"
	SquashTitleCommitOrPRTitle = "COMMIT_OR_PR_TITLE"

	SquashMessagePRBody         = "PR_BODY"
	SquashMessageCommitMessages = "COMMIT_MESSAGES"
	SquashMessageBlank          = "BLANK"

	MergeTitlePRTitle      = "PR_TITLE"
	MergeTitleMergeMessage = "MERGE_MESSAGE"

	MergeMessagePRBody  = "PR_BODY"
	MergeMessagePRTitle = "PR_TITLE"
	MergeMessageBlank   = "BLANK"
)

// Config represents the complete configuration file.
type Config struct {
	Repositories []Repository `yaml:"repositories"`
	Settings     Settings     `yaml:"settings"`
}

// Repository represents a target repository.
type Repository struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

// FullName returns the full repository name (owner/name).
func (r Repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// Settings represents the settings to apply to all repositories
// Use pointers to distinguish between "not set" (nil) and "set to false".
type Settings struct {
	// Merge methods
	AllowMergeCommit    *bool `yaml:"allow_merge_commit,omitempty"`
	AllowSquashMerge    *bool `yaml:"allow_squash_merge,omitempty"`
	AllowRebaseMerge    *bool `yaml:"allow_rebase_merge,omitempty"`
	DeleteBranchOnMerge *bool `yaml:"delete_branch_on_merge,omitempty"`

	// Merge commit messages
	SquashMergeCommitTitle   *string `yaml:"squash_merge_commit_title,omitempty"`
	SquashMergeCommitMessage *string `yaml:"squash_merge_commit_message,omitempty"`
	MergeCommitTitle         *string `yaml:"merge_commit_title,omitempty"`
	MergeCommitMessage       *string `yaml:"merge_commit_message,omitempty"`

	// Repository visibility and features
	Visibility     *string `yaml:"visibility,omitempty"`
	HasIssues      *bool   `yaml:"has_issues,omitempty"`
	HasProjects    *bool   `yaml:"has_projects,omitempty"`
	HasWiki        *bool   `yaml:"has_wiki,omitempty"`
	HasDiscussions *bool   `yaml:"has_discussions,omitempty"`
	Archived       *bool   `yaml:"archived,omitempty"`

	// Additional repository settings
	AllowUpdateBranch        *bool `yaml:"allow_update_branch,omitempty"`
	WebCommitSignoffRequired *bool `yaml:"web_commit_signoff_required,omitempty"`
	AllowForking             *bool `yaml:"allow_forking,omitempty"`

	// Repository metadata
	Description *string  `yaml:"description,omitempty"`
	Homepage    *string  `yaml:"homepage,omitempty"`
	Topics      []string `yaml:"topics,omitempty"`

	// Repository settings
	DefaultBranch  *string `yaml:"default_branch,omitempty"`
	AllowAutoMerge *bool   `yaml:"allow_auto_merge,omitempty"`

	// GitHub Pages
	GitHubPages *GitHubPages `yaml:"github_pages,omitempty"`

	BranchProtection *BranchProtection `yaml:"branch_protection,omitempty"`

	// ActionsSecrets configures GitHub Actions secrets for all repositories.
	ActionsSecrets []ActionsSecret `yaml:"actions_secrets,omitempty"`

	// Security configures GitHub security tooling toggles.
	Security *SecuritySettings `yaml:"security,omitempty"`
}

// SecuritySettings represents GitHub security tooling settings.
type SecuritySettings struct {
	DependabotAlerts          *bool `yaml:"dependabot_alerts,omitempty"`
	DependabotSecurityUpdates *bool `yaml:"dependabot_security_updates,omitempty"`
}

// GitHubPages represents GitHub Pages configuration.
type GitHubPages struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// BranchProtection represents branch protection settings.
type BranchProtection struct {
	Enabled bool   `yaml:"enabled"`
	Pattern string `yaml:"pattern"`

	RequiredReviews     *int  `yaml:"required_reviews,omitempty"`
	RequireStatusChecks *bool `yaml:"require_status_checks,omitempty"`
	DismissStaleReviews *bool `yaml:"dismiss_stale_reviews,omitempty"`

	// StatusCheckContexts controls which status check contexts are required.
	// When omitted, existing required contexts (if any) are preserved.
	StatusCheckContexts []string `yaml:"status_check_contexts,omitempty"`

	// Enhanced branch protection settings
	RequireCodeOwnerReviews       *bool `yaml:"require_code_owner_reviews,omitempty"`
	RequireBranchesUpToDate       *bool `yaml:"require_branches_up_to_date,omitempty"`
	IncludeAdmins                 *bool `yaml:"include_admins,omitempty"`
	RequireLinearHistory          *bool `yaml:"require_linear_history,omitempty"`
	RequireSignedCommits          *bool `yaml:"require_signed_commits,omitempty"`
	RequireConversationResolution *bool `yaml:"require_conversation_resolution,omitempty"`
	AllowForcePushes              *bool `yaml:"allow_force_pushes,omitempty"`
	AllowDeletions                *bool `yaml:"allow_deletions,omitempty"`
}

// ActionsSecret represents a GitHub Actions secret configuration.
type ActionsSecret struct {
	Name    string   `yaml:"name"`
	Env     *string  `yaml:"env,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

// SecretNamePattern validates secret names: must start with letter, contain only uppercase letters, digits, underscores.
var secretNamePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// Validate checks if the secret configuration is valid.
func (s ActionsSecret) Validate() error {
	if s.Name == "" {
		return errors.New("actions_secret: name is required")
	}

	if len(s.Name) > MaxSecretNameLength {
		return fmt.Errorf("actions_secret %s: name exceeds %d characters", s.Name, MaxSecretNameLength)
	}

	if !secretNamePattern.MatchString(s.Name) {
		return fmt.Errorf("actions_secret %s: name must match pattern ^[A-Z][A-Z0-9_]*$", s.Name)
	}

	// Exactly one source must be specified
	hasEnv := s.Env != nil && *s.Env != ""
	hasCommand := len(s.Command) > 0

	if !hasEnv && !hasCommand {
		return fmt.Errorf("actions_secret %s: exactly one of 'env' or 'command' is required", s.Name)
	}

	if hasEnv && hasCommand {
		return fmt.Errorf("actions_secret %s: only one of 'env' or 'command' can be specified", s.Name)
	}

	if hasCommand && s.Command[0] == "" {
		return fmt.Errorf("actions_secret %s: command must have a non-empty executable", s.Name)
	}

	return nil
}

// SourceType returns the type of source for this secret.
func (s ActionsSecret) SourceType() string {
	if s.Env != nil && *s.Env != "" {
		return "env"
	}

	if len(s.Command) > 0 {
		return "command"
	}

	return "unknown"
}

// SourceDescription returns a safe description of the source (no values).
func (s ActionsSecret) SourceDescription() string {
	if s.Env != nil && *s.Env != "" {
		return fmt.Sprintf("env:%s", *s.Env)
	}

	if len(s.Command) > 0 {
		return fmt.Sprintf("command:%s", s.Command[0])
	}

	return "unknown"
}

// Load reads and parses the configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if parseErr := yaml.Unmarshal(data, &cfg); parseErr != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", parseErr)
	}

	if validateErr := cfg.Validate(); validateErr != nil {
		return nil, validateErr
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error { //nolint:gocognit // Validation logic is inherently branching
	if len(c.Repositories) == 0 {
		return errors.New("no repositories configured")
	}

	for i, repo := range c.Repositories {
		if repo.Owner == "" {
			return fmt.Errorf("repository %d: owner is required", i)
		}
		if repo.Name == "" {
			return fmt.Errorf("repository %d: name is required", i)
		}
	}

	if c.Settings.Visibility != nil && *c.Settings.Visibility != VisibilityPublic &&
		*c.Settings.Visibility != VisibilityPrivate {
		return errors.New("invalid visibility: must be 'public' or 'private'")
	}

	if c.Settings.BranchProtection != nil {
		bp := c.Settings.BranchProtection
		if bp.Enabled && bp.Pattern == "" {
			return errors.New("branch_protection: pattern is required when enabled")
		}
		if bp.RequiredReviews != nil {
			if *bp.RequiredReviews < 0 || *bp.RequiredReviews > 6 {
				return errors.New("branch_protection: required_reviews must be between 0 and 6")
			}
		}
	}

	// Validate squash merge commit title
	if c.Settings.SquashMergeCommitTitle != nil {
		valid := []string{SquashTitlePRTitle, SquashTitleCommitOrPRTitle}
		if !contains(valid, *c.Settings.SquashMergeCommitTitle) {
			return fmt.Errorf("invalid squash_merge_commit_title: must be one of %v", valid)
		}
	}

	// Validate squash merge commit message
	if c.Settings.SquashMergeCommitMessage != nil {
		valid := []string{SquashMessagePRBody, SquashMessageCommitMessages, SquashMessageBlank}
		if !contains(valid, *c.Settings.SquashMergeCommitMessage) {
			return fmt.Errorf("invalid squash_merge_commit_message: must be one of %v", valid)
		}
	}

	// Validate merge commit title
	if c.Settings.MergeCommitTitle != nil {
		valid := []string{MergeTitlePRTitle, MergeTitleMergeMessage}
		if !contains(valid, *c.Settings.MergeCommitTitle) {
			return fmt.Errorf("invalid merge_commit_title: must be one of %v", valid)
		}
	}

	// Validate merge commit message
	if c.Settings.MergeCommitMessage != nil {
		valid := []string{MergeMessagePRBody, MergeMessagePRTitle, MergeMessageBlank}
		if !contains(valid, *c.Settings.MergeCommitMessage) {
			return fmt.Errorf("invalid merge_commit_message: must be one of %v", valid)
		}
	}

	// Validate actions secrets
	seenNames := make(map[string]bool)
	for _, secret := range c.Settings.ActionsSecrets {
		if err := secret.Validate(); err != nil {
			return err
		}

		if seenNames[secret.Name] {
			return fmt.Errorf("actions_secret %s: duplicate name", secret.Name)
		}

		seenNames[secret.Name] = true
	}

	if c.Settings.Security != nil &&
		c.Settings.Security.DependabotSecurityUpdates != nil &&
		*c.Settings.Security.DependabotSecurityUpdates {
		if c.Settings.Security.DependabotAlerts == nil || !*c.Settings.Security.DependabotAlerts {
			return errors.New("security: dependabot_alerts must be true when dependabot_security_updates is true")
		}
	}

	return nil
}

// MaxSecretNameLength is the maximum length for a GitHub Actions secret name.
const MaxSecretNameLength = 64

// contains checks if a string slice contains a value.
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// ExampleConfig returns an example configuration as a string.
func ExampleConfig() string {
	return `repositories:
  - owner: mholtzscher
    name: repo1
  - owner: mholtzscher
    name: repo2

settings:
  # Merge methods
  allow_merge_commit: false
  allow_squash_merge: true
  allow_rebase_merge: true
  delete_branch_on_merge: true

  # Merge commit messages (GitHub API values)
  squash_merge_commit_title: PR_TITLE
  squash_merge_commit_message: PR_BODY
  merge_commit_title: PR_TITLE
  merge_commit_message: PR_BODY

  # Repository visibility
  visibility: public

  # Repository features
  has_issues: true
  has_projects: false
  has_wiki: false
  has_discussions: true
  archived: false

  # Additional settings
  allow_update_branch: true
  web_commit_signoff_required: false
  allow_forking: true

  # Repository metadata
  description: "A brief description of the repository"
  homepage: "https://example.com"
  topics: ["go", "cli", "automation"]

  # Repository settings
  default_branch: "main"
  allow_auto_merge: false

  # GitHub Pages (note: enabling requires manual configuration)
  github_pages:
    enabled: false

  # Branch protection (applied to all repos)
  branch_protection:
    enabled: true
    pattern: "main"
    required_reviews: 1
    require_status_checks: true
    status_check_contexts: ["ci/test"]
    dismiss_stale_reviews: true
    # Enhanced protection settings
    require_code_owner_reviews: false
    require_branches_up_to_date: true
    include_admins: false
    require_linear_history: false
    require_signed_commits: false
    require_conversation_resolution: true
    allow_force_pushes: false
    allow_deletions: false

  # GitHub Actions secrets (applied to all repos)
  # Values are resolved at runtime - never stored in config
  actions_secrets:
    # From environment variable
    # - name: NPM_TOKEN
    #   env: NPM_TOKEN
    # From command (e.g., 1Password CLI)
    # - name: AWS_ACCESS_KEY_ID
    #   command: ["op", "read", "op://vault/item/access_key_id"]

  # Security tooling (applied to all repos)
  # security:
  #   dependabot_alerts: true
  #   dependabot_security_updates: true
`
}
