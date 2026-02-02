package config

import (
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration file
type Config struct {
	Repositories []Repository `yaml:"repositories"`
	Settings     Settings     `yaml:"settings"`
}

// Repository represents a target repository
type Repository struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

// FullName returns the full repository name (owner/name)
func (r Repository) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// Settings represents the settings to apply to all repositories
// Use pointers to distinguish between "not set" (nil) and "set to false"
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
}

// GitHubPages represents GitHub Pages configuration
type GitHubPages struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// BranchProtection represents branch protection settings
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

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Repositories) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	for i, repo := range c.Repositories {
		if repo.Owner == "" {
			return fmt.Errorf("repository %d: owner is required", i)
		}
		if repo.Name == "" {
			return fmt.Errorf("repository %d: name is required", i)
		}
	}

	if c.Settings.Visibility != nil && *c.Settings.Visibility != "public" && *c.Settings.Visibility != "private" {
		return fmt.Errorf("invalid visibility: must be 'public' or 'private'")
	}

	if c.Settings.BranchProtection != nil {
		bp := c.Settings.BranchProtection
		if bp.Enabled && bp.Pattern == "" {
			return fmt.Errorf("branch_protection: pattern is required when enabled")
		}
		if bp.RequiredReviews != nil {
			if *bp.RequiredReviews < 0 || *bp.RequiredReviews > 6 {
				return fmt.Errorf("branch_protection: required_reviews must be between 0 and 6")
			}
		}
	}

	// Validate squash merge commit title
	if c.Settings.SquashMergeCommitTitle != nil {
		valid := []string{"PR_TITLE", "COMMIT_OR_PR_TITLE"}
		if !contains(valid, *c.Settings.SquashMergeCommitTitle) {
			return fmt.Errorf("invalid squash_merge_commit_title: must be one of %v", valid)
		}
	}

	// Validate squash merge commit message
	if c.Settings.SquashMergeCommitMessage != nil {
		valid := []string{"PR_BODY", "COMMIT_MESSAGES", "BLANK"}
		if !contains(valid, *c.Settings.SquashMergeCommitMessage) {
			return fmt.Errorf("invalid squash_merge_commit_message: must be one of %v", valid)
		}
	}

	// Validate merge commit title
	if c.Settings.MergeCommitTitle != nil {
		valid := []string{"PR_TITLE", "MERGE_MESSAGE"}
		if !contains(valid, *c.Settings.MergeCommitTitle) {
			return fmt.Errorf("invalid merge_commit_title: must be one of %v", valid)
		}
	}

	// Validate merge commit message
	if c.Settings.MergeCommitMessage != nil {
		valid := []string{"PR_BODY", "PR_TITLE", "BLANK"}
		if !contains(valid, *c.Settings.MergeCommitMessage) {
			return fmt.Errorf("invalid merge_commit_message: must be one of %v", valid)
		}
	}

	return nil
}

// contains checks if a string slice contains a value
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// ExampleConfig returns an example configuration as a string
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
`
}
