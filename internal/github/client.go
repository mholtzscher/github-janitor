package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v82/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client
type Client struct {
	client *github.Client
	ctx    context.Context
}

func derefBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// NewClient creates a new GitHub client with the given token
// If token is empty, it attempts to auto-detect from gh CLI or GITHUB_TOKEN env var
func NewClient(token string) (*Client, error) {
	ctx := context.Background()

	// If no token provided, try to auto-detect
	if token == "" {
		var err error
		token, err = detectToken()
		if err != nil {
			return nil, err
		}
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

// detectToken attempts to find a GitHub token from various sources
func detectToken() (string, error) {
	// First, try GITHUB_TOKEN environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// Second, try to get token from gh CLI
	if token, err := getGhCliToken(); err == nil && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no GitHub token found. Set GITHUB_TOKEN environment variable or authenticate with 'gh auth login'")
}

// getGhCliToken attempts to get a token from the GitHub CLI
func getGhCliToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ValidateAuth checks if the client can authenticate with GitHub
func (c *Client) ValidateAuth() error {
	_, resp, err := c.client.Users.Get(c.ctx, "")
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}
	return nil
}

// GetAuthenticatedUser returns the currently authenticated user
func (c *Client) GetAuthenticatedUser() (string, error) {
	user, _, err := c.client.Users.Get(c.ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}
	if user == nil || user.Login == nil {
		return "", fmt.Errorf("failed to get authenticated user: missing login")
	}
	return *user.Login, nil
}

// RepositoryInfo holds information about a repository
type RepositoryInfo struct {
	Owner            string
	Name             string
	AllowMergeCommit bool
	AllowSquashMerge bool
	AllowRebaseMerge bool
	Private          bool
	Exists           bool

	// Repository metadata
	Description string
	Homepage    string
	Topics      []string

	// Repository settings
	DefaultBranch      string
	AllowAutoMerge     bool
	GitHubPagesEnabled bool

	// New repository settings
	DeleteBranchOnMerge      bool
	SquashMergeCommitTitle   string
	SquashMergeCommitMessage string
	MergeCommitTitle         string
	MergeCommitMessage       string
	HasIssues                bool
	HasProjects              bool
	HasWiki                  bool
	HasDiscussions           bool
	Archived                 bool
	AllowUpdateBranch        bool
	WebCommitSignoffRequired bool
	AllowForking             bool
}

// Repository settings updates use go-github's *github.Repository directly.
// Nil pointer fields are not sent to the GitHub API.

// GetRepository fetches information about a repository
func (c *Client) GetRepository(owner, name string) (*RepositoryInfo, error) {
	repo, resp, err := c.client.Repositories.Get(c.ctx, owner, name)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return &RepositoryInfo{
				Owner:  owner,
				Name:   name,
				Exists: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, name, err)
	}

	info := &RepositoryInfo{
		Owner:            owner,
		Name:             name,
		AllowMergeCommit: derefBool(repo.AllowMergeCommit),
		AllowSquashMerge: derefBool(repo.AllowSquashMerge),
		AllowRebaseMerge: derefBool(repo.AllowRebaseMerge),
		Private:          derefBool(repo.Private),
		Exists:           true,
	}

	// Set repository metadata fields
	if repo.Description != nil {
		info.Description = *repo.Description
	}
	if repo.Homepage != nil {
		info.Homepage = *repo.Homepage
	}
	if repo.Topics != nil {
		info.Topics = repo.Topics
	}

	// Set repository settings fields
	if repo.DefaultBranch != nil {
		info.DefaultBranch = *repo.DefaultBranch
	}
	if repo.AllowAutoMerge != nil {
		info.AllowAutoMerge = *repo.AllowAutoMerge
	}

	// Set GitHub Pages status
	if repo.HasPages != nil {
		info.GitHubPagesEnabled = *repo.HasPages
	}

	// Set new fields if they exist in the API response
	if repo.DeleteBranchOnMerge != nil {
		info.DeleteBranchOnMerge = *repo.DeleteBranchOnMerge
	}
	if repo.SquashMergeCommitTitle != nil {
		info.SquashMergeCommitTitle = *repo.SquashMergeCommitTitle
	}
	if repo.SquashMergeCommitMessage != nil {
		info.SquashMergeCommitMessage = *repo.SquashMergeCommitMessage
	}
	if repo.MergeCommitTitle != nil {
		info.MergeCommitTitle = *repo.MergeCommitTitle
	}
	if repo.MergeCommitMessage != nil {
		info.MergeCommitMessage = *repo.MergeCommitMessage
	}
	if repo.HasIssues != nil {
		info.HasIssues = *repo.HasIssues
	}
	if repo.HasProjects != nil {
		info.HasProjects = *repo.HasProjects
	}
	if repo.HasWiki != nil {
		info.HasWiki = *repo.HasWiki
	}
	if repo.HasDiscussions != nil {
		info.HasDiscussions = *repo.HasDiscussions
	}
	if repo.Archived != nil {
		info.Archived = *repo.Archived
	}
	if repo.AllowUpdateBranch != nil {
		info.AllowUpdateBranch = *repo.AllowUpdateBranch
	}
	if repo.WebCommitSignoffRequired != nil {
		info.WebCommitSignoffRequired = *repo.WebCommitSignoffRequired
	}
	if repo.AllowForking != nil {
		info.AllowForking = *repo.AllowForking
	}

	return info, nil
}

// UpdateRepositorySettings updates repository settings.
// Only non-nil pointer fields in patch are sent to the GitHub API.
func (c *Client) UpdateRepositorySettings(owner, name string, patch *github.Repository) error {
	if patch == nil {
		patch = &github.Repository{}
	}

	_, _, err := c.client.Repositories.Edit(c.ctx, owner, name, patch)
	if err != nil {
		return fmt.Errorf("failed to update repository %s/%s: %w", owner, name, err)
	}

	return nil
}

// BranchProtectionInfo holds branch protection settings
type BranchProtectionInfo struct {
	Enabled bool
	Pattern string

	PullRequestReviewsEnabled bool
	RequiredReviews           int
	DismissStaleReviews       bool
	RequireCodeOwnerReviews   bool

	StatusChecksEnabled     bool
	RequireBranchesUpToDate bool
	StatusCheckContexts     []string
	StatusCheckChecks       []*github.RequiredStatusCheck

	RestrictionsEnabled bool
	RestrictionsUsers   []string
	RestrictionsTeams   []string
	RestrictionsApps    []string

	// Enhanced branch protection settings
	IncludeAdmins                 bool
	RequireLinearHistory          bool
	RequireSignedCommits          bool
	RequireConversationResolution bool
	AllowForcePushes              bool
	AllowDeletions                bool
}

// GetBranchProtection fetches branch protection settings
func (c *Client) GetBranchProtection(owner, name, pattern string) (*BranchProtectionInfo, error) {
	protection, resp, err := c.client.Repositories.GetBranchProtection(c.ctx, owner, name, pattern)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			// No protection exists
			return &BranchProtectionInfo{
				Enabled: false,
				Pattern: pattern,
			}, nil
		}
		return nil, fmt.Errorf("failed to get branch protection: %w", err)
	}

	info := &BranchProtectionInfo{
		Enabled: true,
		Pattern: pattern,
	}

	if protection.RequiredPullRequestReviews != nil {
		info.PullRequestReviewsEnabled = true
		info.RequiredReviews = protection.RequiredPullRequestReviews.RequiredApprovingReviewCount
		info.DismissStaleReviews = protection.RequiredPullRequestReviews.DismissStaleReviews
		info.RequireCodeOwnerReviews = protection.RequiredPullRequestReviews.RequireCodeOwnerReviews
	}

	if protection.RequiredStatusChecks != nil {
		info.StatusChecksEnabled = true
		info.RequireBranchesUpToDate = protection.RequiredStatusChecks.Strict
		if protection.RequiredStatusChecks.Contexts != nil {
			info.StatusCheckContexts = *protection.RequiredStatusChecks.Contexts
		}
		if protection.RequiredStatusChecks.Checks != nil {
			info.StatusCheckChecks = *protection.RequiredStatusChecks.Checks
		}
	}

	// Set enhanced protection fields
	if protection.EnforceAdmins != nil {
		info.IncludeAdmins = protection.EnforceAdmins.Enabled
	}
	if protection.Restrictions != nil {
		info.RestrictionsEnabled = true
		for _, u := range protection.Restrictions.Users {
			if u != nil && u.Login != nil {
				info.RestrictionsUsers = append(info.RestrictionsUsers, *u.Login)
			}
		}
		for _, team := range protection.Restrictions.Teams {
			if team != nil && team.Slug != nil {
				info.RestrictionsTeams = append(info.RestrictionsTeams, *team.Slug)
			}
		}
		for _, app := range protection.Restrictions.Apps {
			if app != nil && app.Slug != nil {
				info.RestrictionsApps = append(info.RestrictionsApps, *app.Slug)
			}
		}
	}
	if protection.RequireLinearHistory != nil {
		info.RequireLinearHistory = protection.RequireLinearHistory.Enabled
	}
	if protection.RequiredSignatures != nil && protection.RequiredSignatures.Enabled != nil {
		info.RequireSignedCommits = *protection.RequiredSignatures.Enabled
	}
	if protection.RequiredConversationResolution != nil {
		info.RequireConversationResolution = protection.RequiredConversationResolution.Enabled
	}
	if protection.AllowForcePushes != nil {
		info.AllowForcePushes = protection.AllowForcePushes.Enabled
	}
	if protection.AllowDeletions != nil {
		info.AllowDeletions = protection.AllowDeletions.Enabled
	}

	return info, nil
}

// UpdateBranchProtection updates branch protection settings
func (c *Client) UpdateBranchProtection(owner, name string, protection *BranchProtectionInfo) error {
	if !protection.Enabled {
		// Remove protection if disabled
		_, err := c.client.Repositories.RemoveBranchProtection(c.ctx, owner, name, protection.Pattern)
		if err != nil {
			return err
		}
		return nil
	}

	req := buildProtectionRequest(protection)

	_, _, err := c.client.Repositories.UpdateBranchProtection(c.ctx, owner, name, protection.Pattern, req)
	if err != nil {
		return fmt.Errorf("failed to update branch protection: %w", err)
	}

	if err := c.updateRequiredSignatures(owner, name, protection.Pattern, protection.RequireSignedCommits); err != nil {
		return fmt.Errorf("failed to update required signatures: %w", err)
	}

	return nil
}

func buildProtectionRequest(protection *BranchProtectionInfo) *github.ProtectionRequest {
	if protection == nil {
		return &github.ProtectionRequest{}
	}

	var reqReviews *github.PullRequestReviewsEnforcementRequest
	if protection.PullRequestReviewsEnabled {
		reqReviews = &github.PullRequestReviewsEnforcementRequest{
			RequiredApprovingReviewCount: protection.RequiredReviews,
			DismissStaleReviews:          protection.DismissStaleReviews,
			RequireCodeOwnerReviews:      protection.RequireCodeOwnerReviews,
		}
	}

	var reqChecks *github.RequiredStatusChecks
	if protection.StatusChecksEnabled {
		var contexts *[]string
		var checks *[]*github.RequiredStatusCheck

		hasContexts := len(protection.StatusCheckContexts) > 0
		hasChecks := len(protection.StatusCheckChecks) > 0

		switch {
		case hasContexts && !hasChecks:
			// If we set contexts explicitly, also clear any existing checks.
			contexts = &protection.StatusCheckContexts
			emptyChecks := []*github.RequiredStatusCheck{}
			checks = &emptyChecks
		case hasChecks && !hasContexts:
			// If we set checks explicitly, also clear any existing contexts.
			checks = &protection.StatusCheckChecks
			emptyContexts := []string{}
			contexts = &emptyContexts
		case hasContexts && hasChecks:
			contexts = &protection.StatusCheckContexts
			checks = &protection.StatusCheckChecks
		}
		reqChecks = &github.RequiredStatusChecks{
			Strict:   protection.RequireBranchesUpToDate,
			Contexts: contexts,
			Checks:   checks,
		}
	}

	var restrictions *github.BranchRestrictionsRequest
	if protection.RestrictionsEnabled {
		users := protection.RestrictionsUsers
		teams := protection.RestrictionsTeams
		apps := protection.RestrictionsApps
		if users == nil {
			users = []string{}
		}
		if teams == nil {
			teams = []string{}
		}
		if apps == nil {
			apps = []string{}
		}
		restrictions = &github.BranchRestrictionsRequest{Users: users, Teams: teams, Apps: apps}
	}

	return &github.ProtectionRequest{
		RequiredStatusChecks:           reqChecks,
		RequiredPullRequestReviews:     reqReviews,
		EnforceAdmins:                  protection.IncludeAdmins,
		Restrictions:                   restrictions,
		RequireLinearHistory:           &protection.RequireLinearHistory,
		AllowForcePushes:               &protection.AllowForcePushes,
		AllowDeletions:                 &protection.AllowDeletions,
		RequiredConversationResolution: &protection.RequireConversationResolution,
	}
}

func (c *Client) updateRequiredSignatures(owner, name, pattern string, required bool) error {
	if required {
		_, _, err := c.client.Repositories.RequireSignaturesOnProtectedBranch(c.ctx, owner, name, pattern)
		if err != nil {
			return fmt.Errorf("failed to require signatures on protected branch %s/%s:%s: %w", owner, name, pattern, err)
		}
		return nil
	}
	_, err := c.client.Repositories.OptionalSignaturesOnProtectedBranch(c.ctx, owner, name, pattern)
	if err != nil {
		return fmt.Errorf("failed to make signatures optional on protected branch %s/%s:%s: %w", owner, name, pattern, err)
	}
	return nil
}
