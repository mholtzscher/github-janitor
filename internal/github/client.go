package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v82/github"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

const (
	EnvToken          = "GITHUB_TOKEN" //nolint:gosec // Not a credential, just env var name
	TokenSourceFlag   = "--token flag"
	TokenSourceEnvVar = "GITHUB_TOKEN env var" //nolint:gosec // Not a credential, just source name
	TokenSourceGhCLI  = "gh CLI"
)

// Client wraps the GitHub API client.
type Client struct {
	client      *github.Client
	ctx         context.Context
	TokenSource string
}

func derefBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// NewClient creates a new GitHub client with the given token
// If token is empty, it attempts to auto-detect from gh CLI or GITHUB_TOKEN env var.
func NewClient(token string) (*Client, error) {
	ctx := context.Background()
	tokenSource := TokenSourceFlag

	// If no token provided, try to auto-detect
	if token == "" {
		var err error
		token, tokenSource, err = detectToken()
		if err != nil {
			return nil, err
		}
	}

	token = strings.TrimSpace(token)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{
		client:      client,
		ctx:         ctx,
		TokenSource: tokenSource,
	}, nil
}

// detectToken attempts to find a GitHub token from various sources.
func detectToken() (string, string, error) {
	// First, try GITHUB_TOKEN environment variable
	if token := os.Getenv(EnvToken); token != "" {
		return strings.TrimSpace(token), TokenSourceEnvVar, nil
	}

	// Second, try to get token from gh CLI
	if token, err := getGhCliToken(); err == nil && token != "" {
		return token, TokenSourceGhCLI, nil
	}

	return "", "", fmt.Errorf(
		"no GitHub token found. Set %s environment variable or authenticate with 'gh auth login'",
		EnvToken,
	)
}

// getGhCliToken attempts to get a token from the GitHub CLI.
func getGhCliToken() (string, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ValidateAuth checks if the client can authenticate with GitHub.
func (c *Client) ValidateAuth() error {
	_, resp, err := c.client.Users.Get(c.ctx, "")
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}
	return nil
}

// GetAuthenticatedUser returns the currently authenticated user.
func (c *Client) GetAuthenticatedUser() (string, error) {
	user, _, err := c.client.Users.Get(c.ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}
	if user == nil || user.Login == nil {
		return "", errors.New("failed to get authenticated user: missing login")
	}
	return *user.Login, nil
}

// RepositoryInfo holds information about a repository.
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

// GetRepository fetches information about a repository.
func (c *Client) GetRepository( //nolint:gocognit // Field mapping is straightforward
	owner, name string,
) (*RepositoryInfo, error) {
	repo, resp, err := c.client.Repositories.Get(c.ctx, owner, name)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
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

// BranchProtectionInfo holds branch protection settings.
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

// SecuritySettingsInfo holds repository security tooling settings.
type SecuritySettingsInfo struct {
	DependabotAlerts                bool
	DependabotSecurityUpdates       bool
	DependabotSecurityUpdatesPaused bool
}

// GetBranchProtection fetches branch protection settings.
func (c *Client) GetBranchProtection( //nolint:gocognit // Field mapping is straightforward
	owner, name, pattern string,
) (*BranchProtectionInfo, error) {
	protection, resp, err := c.client.Repositories.GetBranchProtection(c.ctx, owner, name, pattern)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
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

// UpdateBranchProtection updates branch protection settings.
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

	if sigErr := c.updateRequiredSignatures(
		owner,
		name,
		protection.Pattern,
		protection.RequireSignedCommits,
	); sigErr != nil {
		return fmt.Errorf("failed to update required signatures: %w", sigErr)
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
			return fmt.Errorf(
				"failed to require signatures on protected branch %s/%s:%s: %w",
				owner,
				name,
				pattern,
				err,
			)
		}
		return nil
	}
	_, err := c.client.Repositories.OptionalSignaturesOnProtectedBranch(c.ctx, owner, name, pattern)
	if err != nil {
		return fmt.Errorf(
			"failed to make signatures optional on protected branch %s/%s:%s: %w",
			owner,
			name,
			pattern,
			err,
		)
	}
	return nil
}

// GetSecuritySettings fetches repository security tooling settings.
func (c *Client) GetSecuritySettings(owner, name string) (*SecuritySettingsInfo, error) {
	alertsEnabled, _, err := c.client.Repositories.GetVulnerabilityAlerts(c.ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependabot alerts for %s/%s: %w", owner, name, err)
	}

	fixes, _, err := c.client.Repositories.GetAutomatedSecurityFixes(c.ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependabot security updates for %s/%s: %w", owner, name, err)
	}

	info := &SecuritySettingsInfo{DependabotAlerts: alertsEnabled}
	if fixes != nil {
		info.DependabotSecurityUpdates = fixes.GetEnabled()
		info.DependabotSecurityUpdatesPaused = fixes.GetPaused()
	}

	return info, nil
}

// SetDependabotAlerts enables or disables dependabot alerts.
func (c *Client) SetDependabotAlerts(owner, name string, enabled bool) error {
	var err error

	if enabled {
		_, err = c.client.Repositories.EnableVulnerabilityAlerts(c.ctx, owner, name)
	} else {
		_, err = c.client.Repositories.DisableVulnerabilityAlerts(c.ctx, owner, name)
	}

	if err != nil {
		return fmt.Errorf("failed to set dependabot alerts=%t for %s/%s: %w", enabled, owner, name, err)
	}

	return nil
}

// SetDependabotSecurityUpdates enables or disables dependabot security updates.
func (c *Client) SetDependabotSecurityUpdates(owner, name string, enabled bool) error {
	var err error

	if enabled {
		_, err = c.client.Repositories.EnableAutomatedSecurityFixes(c.ctx, owner, name)
	} else {
		_, err = c.client.Repositories.DisableAutomatedSecurityFixes(c.ctx, owner, name)
	}

	if err != nil {
		return fmt.Errorf("failed to set dependabot security updates=%t for %s/%s: %w", enabled, owner, name, err)
	}

	return nil
}

// ActionsSecretPublicKey represents the public key for encrypting secrets.
type ActionsSecretPublicKey struct {
	KeyID string `json:"key_id"`
	Key   string `json:"key"`
}

// GetActionsSecretPublicKey fetches the public key for encrypting repository secrets.
func (c *Client) GetActionsSecretPublicKey(owner, name string) (*ActionsSecretPublicKey, error) {
	url := fmt.Sprintf("repos/%s/%s/actions/secrets/public-key", owner, name)
	req, err := c.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var key ActionsSecretPublicKey
	resp, err := c.client.Do(c.ctx, req, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}
	defer resp.Body.Close()

	return &key, nil
}

// EncryptSecret encrypts a secret value using the repository's public key.
// Uses libsodium sealed box encryption (box.SealAnonymous).
func (c *Client) EncryptSecret(publicKeyB64, secretValue string) (string, error) {
	// Decode the Base64 public key
	decodedPubKey, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	// NaCl public keys are 32 bytes
	const publicKeySize = 32
	if len(decodedPubKey) != publicKeySize {
		return "", fmt.Errorf("invalid public key length: expected %d bytes, got %d", publicKeySize, len(decodedPubKey))
	}

	// Copy into [32]byte array (NaCl requires fixed-size keys)
	var peersPubKey [publicKeySize]byte
	copy(peersPubKey[:], decodedPubKey)

	// Encrypt using sealed box (anonymous encryption)
	encrypted, err := box.SealAnonymous(nil, []byte(secretValue), &peersPubKey, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Return Base64-encoded encrypted value
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// SetActionsSecret creates or updates a repository secret.
func (c *Client) SetActionsSecret(owner, name, secretName, encryptedValue, keyID string) error {
	url := fmt.Sprintf("repos/%s/%s/actions/secrets/%s", owner, name, secretName)

	payload := map[string]string{
		"encrypted_value": encryptedValue,
		"key_id":          keyID,
	}

	req, err := c.client.NewRequest(http.MethodPut, url, payload)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(c.ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to set secret: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
