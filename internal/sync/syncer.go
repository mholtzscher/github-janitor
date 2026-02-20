package reposync

import (
	"context"
	"fmt"
	"reflect"

	gogithub "github.com/google/go-github/v82/github"

	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/github"
	"github.com/mholtzscher/github-janitor/internal/secrets"
)

// Syncer orchestrates the synchronization of repository settings.
type Syncer struct {
	client githubAPI
	config *config.Config
}

type githubAPI interface {
	GetRepository(owner, name string) (*github.RepositoryInfo, error)
	UpdateRepositorySettings(owner, name string, patch *gogithub.Repository) error
	GetSecuritySettings(owner, name string) (*github.SecuritySettingsInfo, error)
	SetDependabotAlerts(owner, name string, enabled bool) error
	SetDependabotSecurityUpdates(owner, name string, enabled bool) error
	GetBranchProtection(owner, name, pattern string) (*github.BranchProtectionInfo, error)
	UpdateBranchProtection(owner, name string, protection *github.BranchProtectionInfo) error
	GetActionsSecretPublicKey(owner, name string) (*github.ActionsSecretPublicKey, error)
	EncryptSecret(publicKeyB64, secretValue string) (string, error)
	SetActionsSecret(owner, name, secretName, encryptedValue, keyID string) error
}

// Change represents a single setting change.
type Change struct {
	Field   string
	Current any
	Desired any
}

// applySetting updates the API patch (when configured) and tracks changes.
// Returns true if a change was detected.
func applySetting[T comparable](result *Result, field string, configured *T, current T, patchField **T) bool {
	if configured == nil {
		return false
	}

	*patchField = configured

	if current != *configured {
		result.Changes = append(result.Changes, Change{
			Field:   field,
			Current: current,
			Desired: *configured,
		})
		return true
	}

	return false
}

// applyDesiredSetting applies an optional config value to a desired target and tracks changes.
// Returns true if a change was detected.
func applyDesiredSetting[T comparable](result *Result, field string, configured *T, desired *T) bool {
	if configured == nil {
		return false
	}

	if *desired != *configured {
		result.Changes = append(result.Changes, Change{
			Field:   field,
			Current: *desired,
			Desired: *configured,
		})
		*desired = *configured
		return true
	}

	return false
}

func applyDesiredStringSlice(result *Result, field string, configured []string, desired *[]string) bool {
	if len(configured) == 0 {
		return false
	}

	changed := !reflect.DeepEqual(*desired, configured)
	if changed {
		result.Changes = append(result.Changes, Change{
			Field:   field,
			Current: *desired,
			Desired: configured,
		})
	}

	*desired = append([]string(nil), configured...)
	return changed
}

// Result represents the result of syncing a single repository.
type Result struct {
	Repository string
	Exists     bool
	Changes    []Change
	Error      error
}

// NewSyncer creates a new syncer instance.
func NewSyncer(client *github.Client, cfg *config.Config) *Syncer {
	return &Syncer{
		client: client,
		config: cfg,
	}
}

// SyncAll syncs all configured repositories.
func (s *Syncer) SyncAll(ctx context.Context, dryRun bool) ([]Result, error) {
	results := make([]Result, 0, len(s.config.Repositories))

	// Resolve secrets once before iterating repos (only in apply mode)
	var secretValues map[string]string

	if !dryRun && len(s.config.Settings.ActionsSecrets) > 0 {
		resolver := secrets.NewResolver(s.config.Settings.ActionsSecrets)

		values, err := resolver.ResolveAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve secrets: %w", err)
		}

		secretValues = values
	}

	for _, repo := range s.config.Repositories {
		result := s.syncRepository(ctx, repo, dryRun, secretValues)
		results = append(results, result)
	}

	return results, nil
}

// syncRepository syncs a single repository.
func (s *Syncer) syncRepository( //nolint:cyclop,funlen,gocognit,gocyclo // Sync logic maps many settings
	ctx context.Context,
	repo config.Repository,
	dryRun bool,
	secretValues map[string]string,
) Result {
	result := Result{
		Repository: repo.FullName(),
		Changes:    make([]Change, 0),
	}

	// Get current repository info
	current, err := s.client.GetRepository(repo.Owner, repo.Name)
	if err != nil {
		result.Error = err
		return result
	}

	if !current.Exists {
		result.Exists = false
		return result
	}

	result.Exists = true

	patch := &gogithub.Repository{}
	changed := applySetting(
		&result,
		"allow_merge_commit",
		s.config.Settings.AllowMergeCommit,
		current.AllowMergeCommit,
		&patch.AllowMergeCommit,
	)

	// Track boolean settings
	changed = applySetting(
		&result,
		"allow_squash_merge",
		s.config.Settings.AllowSquashMerge,
		current.AllowSquashMerge,
		&patch.AllowSquashMerge,
	) ||
		changed
	changed = applySetting(
		&result,
		"allow_rebase_merge",
		s.config.Settings.AllowRebaseMerge,
		current.AllowRebaseMerge,
		&patch.AllowRebaseMerge,
	) ||
		changed
	changed = applySetting(
		&result,
		"delete_branch_on_merge",
		s.config.Settings.DeleteBranchOnMerge,
		current.DeleteBranchOnMerge,
		&patch.DeleteBranchOnMerge,
	) ||
		changed
	changed = applySetting(&result, "has_issues", s.config.Settings.HasIssues, current.HasIssues, &patch.HasIssues) ||
		changed
	changed = applySetting(
		&result,
		"has_projects",
		s.config.Settings.HasProjects,
		current.HasProjects,
		&patch.HasProjects,
	) ||
		changed
	changed = applySetting(&result, "has_wiki", s.config.Settings.HasWiki, current.HasWiki, &patch.HasWiki) || changed
	changed = applySetting(
		&result,
		"has_discussions",
		s.config.Settings.HasDiscussions,
		current.HasDiscussions,
		&patch.HasDiscussions,
	) ||
		changed
	changed = applySetting(&result, "archived", s.config.Settings.Archived, current.Archived, &patch.Archived) ||
		changed
	changed = applySetting(
		&result,
		"allow_update_branch",
		s.config.Settings.AllowUpdateBranch,
		current.AllowUpdateBranch,
		&patch.AllowUpdateBranch,
	) ||
		changed
	changed = applySetting(
		&result,
		"web_commit_signoff_required",
		s.config.Settings.WebCommitSignoffRequired,
		current.WebCommitSignoffRequired,
		&patch.WebCommitSignoffRequired,
	) ||
		changed
	changed = applySetting(
		&result,
		"allow_forking",
		s.config.Settings.AllowForking,
		current.AllowForking,
		&patch.AllowForking,
	) ||
		changed

	// Track string settings
	changed = applySetting(
		&result,
		"squash_merge_commit_title",
		s.config.Settings.SquashMergeCommitTitle,
		current.SquashMergeCommitTitle,
		&patch.SquashMergeCommitTitle,
	) ||
		changed
	changed = applySetting(
		&result,
		"squash_merge_commit_message",
		s.config.Settings.SquashMergeCommitMessage,
		current.SquashMergeCommitMessage,
		&patch.SquashMergeCommitMessage,
	) ||
		changed
	changed = applySetting(
		&result,
		"merge_commit_title",
		s.config.Settings.MergeCommitTitle,
		current.MergeCommitTitle,
		&patch.MergeCommitTitle,
	) ||
		changed
	changed = applySetting(
		&result,
		"merge_commit_message",
		s.config.Settings.MergeCommitMessage,
		current.MergeCommitMessage,
		&patch.MergeCommitMessage,
	) ||
		changed

	// Track repository metadata
	changed = applySetting(
		&result,
		"description",
		s.config.Settings.Description,
		current.Description,
		&patch.Description,
	) ||
		changed
	changed = applySetting(&result, "homepage", s.config.Settings.Homepage, current.Homepage, &patch.Homepage) ||
		changed

	// Track topics (special case: slice)
	if len(s.config.Settings.Topics) > 0 {
		changed = applyDesiredStringSlice(&result, "topics", s.config.Settings.Topics, &patch.Topics) || changed
	}

	// Track default branch
	changed = applySetting(
		&result,
		"default_branch",
		s.config.Settings.DefaultBranch,
		current.DefaultBranch,
		&patch.DefaultBranch,
	) ||
		changed

	// Track auto-merge setting
	changed = applySetting(
		&result,
		"allow_auto_merge",
		s.config.Settings.AllowAutoMerge,
		current.AllowAutoMerge,
		&patch.AllowAutoMerge,
	) ||
		changed

	// Track visibility (special case: maps string to bool)
	if s.config.Settings.Visibility != nil {
		desiredPrivate := *s.config.Settings.Visibility == "private"
		patch.Private = &desiredPrivate
		if current.Private != desiredPrivate {
			visibilityMap := map[bool]string{true: "private", false: "public"}
			result.Changes = append(result.Changes, Change{
				Field:   "visibility",
				Current: visibilityMap[current.Private],
				Desired: visibilityMap[desiredPrivate],
			})
			changed = true
		}
	}

	if !dryRun && changed {
		if updateErr := s.client.UpdateRepositorySettings(repo.Owner, repo.Name, patch); updateErr != nil {
			result.Error = fmt.Errorf("failed to update settings: %w", updateErr)
			return result
		}
	}

	if s.config.Settings.Security != nil {
		securityResult := s.syncSecuritySettings(repo, dryRun)
		result.Changes = append(result.Changes, securityResult.Changes...)
		if securityResult.Error != nil {
			result.Error = securityResult.Error
		}
	}

	// Handle GitHub Pages separately (requires different API)
	if s.config.Settings.GitHubPages != nil && s.config.Settings.GitHubPages.Enabled != nil {
		desiredPagesEnabled := *s.config.Settings.GitHubPages.Enabled
		if current.GitHubPagesEnabled != desiredPagesEnabled {
			// GitHub Pages requires a separate API call to enable/disable
			// For now, we track the change but don't apply it automatically
			result.Changes = append(result.Changes, Change{
				Field:   "github_pages",
				Current: current.GitHubPagesEnabled,
				Desired: desiredPagesEnabled,
			})
			result.Error = fmt.Errorf(
				"GitHub Pages must be configured manually (API limitation): visit https://github.com/%s/%s/settings/pages",
				repo.Owner,
				repo.Name,
			)
		}
	}

	// Sync branch protection if configured
	if s.config.Settings.BranchProtection != nil {
		bpResult := s.syncBranchProtection(repo, dryRun)
		result.Changes = append(result.Changes, bpResult.Changes...)
		if bpResult.Error != nil {
			result.Error = bpResult.Error
		}
	}

	// Sync actions secrets if configured
	if len(s.config.Settings.ActionsSecrets) > 0 {
		secretsResult := s.syncActionsSecrets(ctx, repo, dryRun, secretValues)
		result.Changes = append(result.Changes, secretsResult.Changes...)
		if secretsResult.Error != nil {
			result.Error = secretsResult.Error
		}
	}

	return result
}

//nolint:gocognit // Explicit ordering logic is intentional.
func (s *Syncer) syncSecuritySettings(repo config.Repository, dryRun bool) Result {
	result := Result{
		Repository: fmt.Sprintf("%s (security)", repo.FullName()),
		Changes:    make([]Change, 0),
	}

	securityConfig := s.config.Settings.Security
	if securityConfig == nil {
		return result
	}

	current, err := s.client.GetSecuritySettings(repo.Owner, repo.Name)
	if err != nil {
		result.Error = err
		return result
	}
	if current == nil {
		result.Error = fmt.Errorf("failed to get security settings for %s: empty response", repo.FullName())
		return result
	}

	alertsChanged := false
	alertsDesired := current.DependabotAlerts
	if securityConfig.DependabotAlerts != nil {
		alertsDesired = *securityConfig.DependabotAlerts
		if current.DependabotAlerts != alertsDesired {
			alertsChanged = true
			result.Changes = append(result.Changes, Change{
				Field:   "security.dependabot_alerts",
				Current: current.DependabotAlerts,
				Desired: alertsDesired,
			})
		}
	}

	updatesChanged := false
	updatesDesired := current.DependabotSecurityUpdates
	if securityConfig.DependabotSecurityUpdates != nil {
		updatesDesired = *securityConfig.DependabotSecurityUpdates
		if current.DependabotSecurityUpdates != updatesDesired {
			updatesChanged = true
			result.Changes = append(result.Changes, Change{
				Field:   "security.dependabot_security_updates",
				Current: current.DependabotSecurityUpdates,
				Desired: updatesDesired,
			})
		}
	}

	if dryRun {
		return result
	}

	if updatesChanged && !updatesDesired {
		if setErr := s.client.SetDependabotSecurityUpdates(repo.Owner, repo.Name, false); setErr != nil {
			result.Error = setErr
			return result
		}
	}

	if alertsChanged && !alertsDesired {
		if setErr := s.client.SetDependabotAlerts(repo.Owner, repo.Name, false); setErr != nil {
			result.Error = setErr
			return result
		}
	}

	if alertsChanged && alertsDesired {
		if setErr := s.client.SetDependabotAlerts(repo.Owner, repo.Name, true); setErr != nil {
			result.Error = setErr
			return result
		}
	}

	if updatesChanged && updatesDesired {
		if setErr := s.client.SetDependabotSecurityUpdates(repo.Owner, repo.Name, true); setErr != nil {
			result.Error = setErr
			return result
		}
	}

	return result
}

// syncBranchProtection syncs branch protection settings.
func (s *Syncer) syncBranchProtection( //nolint:funlen,gocognit // Protection settings are numerous
	repo config.Repository,
	dryRun bool,
) Result {
	bp := s.config.Settings.BranchProtection

	result := Result{
		Repository: fmt.Sprintf("%s (branch: %s)", repo.FullName(), bp.Pattern),
		Changes:    make([]Change, 0),
	}

	pattern := bp.Pattern

	// Get current protection
	current, err := s.client.GetBranchProtection(repo.Owner, repo.Name, pattern)
	if err != nil {
		result.Error = err
		return result
	}

	desired := *current
	desired.Pattern = pattern
	desired.Enabled = bp.Enabled

	changed := false

	if current.Enabled != desired.Enabled {
		result.Changes = append(
			result.Changes,
			Change{
				Field:   "branch_protection",
				Current: map[bool]string{true: "enabled", false: "disabled"}[current.Enabled],
				Desired: map[bool]string{true: "enabled", false: "disabled"}[desired.Enabled],
			},
		)
		changed = true
	}

	// If branch protection is being disabled, the only action is removal.
	if !bp.Enabled {
		if !dryRun && changed {
			if updateErr := s.client.UpdateBranchProtection(repo.Owner, repo.Name, &desired); updateErr != nil {
				result.Error = fmt.Errorf("failed to update branch protection: %w", updateErr)
			}
		}
		return result
	}

	// Pull request review requirements
	if bp.RequiredReviews != nil || bp.DismissStaleReviews != nil || bp.RequireCodeOwnerReviews != nil {
		desired.PullRequestReviewsEnabled = true
	}
	if current.PullRequestReviewsEnabled != desired.PullRequestReviewsEnabled {
		result.Changes = append(
			result.Changes,
			Change{
				Field:   "pull_request_reviews_enabled",
				Current: current.PullRequestReviewsEnabled,
				Desired: desired.PullRequestReviewsEnabled,
			},
		)
		changed = true
	}
	changed = applyDesiredSetting(&result, "required_reviews", bp.RequiredReviews, &desired.RequiredReviews) || changed
	changed = applyDesiredSetting(
		&result,
		"dismiss_stale_reviews",
		bp.DismissStaleReviews,
		&desired.DismissStaleReviews,
	) ||
		changed
	changed = applyDesiredSetting(
		&result,
		"require_code_owner_reviews",
		bp.RequireCodeOwnerReviews,
		&desired.RequireCodeOwnerReviews,
	) ||
		changed

	// Status checks
	changed = applyDesiredSetting(
		&result,
		"require_status_checks",
		bp.RequireStatusChecks,
		&desired.StatusChecksEnabled,
	) ||
		changed
	changed = applyDesiredSetting(
		&result,
		"require_branches_up_to_date",
		bp.RequireBranchesUpToDate,
		&desired.RequireBranchesUpToDate,
	) ||
		changed
	if len(bp.StatusCheckContexts) > 0 {
		changed = applyDesiredStringSlice(
			&result,
			"status_check_contexts",
			bp.StatusCheckContexts,
			&desired.StatusCheckContexts,
		) ||
			changed
		desired.StatusCheckChecks = nil
	}

	if desired.StatusChecksEnabled {
		if len(desired.StatusCheckContexts) == 0 && len(desired.StatusCheckChecks) == 0 {
			result.Error = fmt.Errorf(
				"branch protection %s: require_status_checks is true but no status_check_contexts are configured and none exist on the branch",
				repo.FullName(),
			)
			return result
		}
	}

	changed = applyDesiredSetting(&result, "include_admins", bp.IncludeAdmins, &desired.IncludeAdmins) || changed
	changed = applyDesiredSetting(
		&result,
		"require_linear_history",
		bp.RequireLinearHistory,
		&desired.RequireLinearHistory,
	) ||
		changed
	changed = applyDesiredSetting(
		&result,
		"require_signed_commits",
		bp.RequireSignedCommits,
		&desired.RequireSignedCommits,
	) ||
		changed
	changed = applyDesiredSetting(
		&result,
		"require_conversation_resolution",
		bp.RequireConversationResolution,
		&desired.RequireConversationResolution,
	) ||
		changed
	changed = applyDesiredSetting(&result, "allow_force_pushes", bp.AllowForcePushes, &desired.AllowForcePushes) ||
		changed
	changed = applyDesiredSetting(&result, "allow_deletions", bp.AllowDeletions, &desired.AllowDeletions) || changed

	// Apply changes if not dry-run
	if !dryRun && changed {
		if updateErr := s.client.UpdateBranchProtection(repo.Owner, repo.Name, &desired); updateErr != nil {
			result.Error = fmt.Errorf("failed to update branch protection: %w", updateErr)
			return result
		}
	}

	return result
}

// syncActionsSecrets syncs GitHub Actions secrets for a repository.
func (s *Syncer) syncActionsSecrets(
	_ context.Context,
	repo config.Repository,
	dryRun bool,
	secretValues map[string]string,
) Result {
	result := Result{
		Repository: fmt.Sprintf("%s (secrets)", repo.FullName()),
		Changes:    make([]Change, 0),
	}

	// In dry-run mode, just track that secrets would be set (don't resolve values)
	if dryRun {
		resolver := secrets.NewResolver(s.config.Settings.ActionsSecrets)

		for _, secret := range s.config.Settings.ActionsSecrets {
			source, _ := resolver.GetSource(secret.Name)
			desc := "unknown"
			if source != nil {
				desc = source.Describe()
			}

			result.Changes = append(result.Changes, Change{
				Field:   fmt.Sprintf("actions_secret.%s", secret.Name),
				Current: "<write-only>",
				Desired: fmt.Sprintf("set (%s)", desc),
			})
		}

		return result
	}

	// Apply mode: secretValues should already be resolved
	if len(secretValues) == 0 {
		return result
	}

	// Fetch public key once for all secrets
	publicKey, err := s.client.GetActionsSecretPublicKey(repo.Owner, repo.Name)
	if err != nil {
		result.Error = fmt.Errorf("failed to get public key: %w", err)
		return result
	}

	// Set each secret
	for name, value := range secretValues {
		encryptedValue, encErr := s.client.EncryptSecret(publicKey.Key, value)
		if encErr != nil {
			result.Error = fmt.Errorf("failed to encrypt secret %s: %w", name, encErr)
			return result
		}

		if setErr := s.client.SetActionsSecret(
			repo.Owner,
			repo.Name,
			name,
			encryptedValue,
			publicKey.KeyID,
		); setErr != nil {
			result.Error = fmt.Errorf("failed to set secret %s: %w", name, setErr)
			return result
		}

		result.Changes = append(result.Changes, Change{
			Field:   fmt.Sprintf("actions_secret.%s", name),
			Current: "<write-only>",
			Desired: "set",
		})
	}

	return result
}
