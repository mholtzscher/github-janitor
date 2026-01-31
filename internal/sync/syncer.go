package sync

import (
	"fmt"
	"reflect"

	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/github"
)

// Syncer orchestrates the synchronization of repository settings
type Syncer struct {
	client *github.Client
	config *config.Config
}

// Change represents a single setting change
type Change struct {
	Field   string
	Current interface{}
	Desired interface{}
}

// Result represents the result of syncing a single repository
type Result struct {
	Repository string
	Exists     bool
	Changes    []Change
	Error      error
}

// NewSyncer creates a new syncer instance
func NewSyncer(client *github.Client, cfg *config.Config) *Syncer {
	return &Syncer{
		client: client,
		config: cfg,
	}
}

// SyncAll syncs all configured repositories
func (s *Syncer) SyncAll(dryRun bool) ([]Result, error) {
	results := make([]Result, 0, len(s.config.Repositories))

	for _, repo := range s.config.Repositories {
		result := s.syncRepository(repo, dryRun)
		results = append(results, result)
	}

	return results, nil
}

// syncRepository syncs a single repository
func (s *Syncer) syncRepository(repo config.Repository, dryRun bool) Result {
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

	patch := &github.RepositorySettingsPatch{}
	changed := false

	if s.config.Settings.AllowMergeCommit != nil {
		desired := *s.config.Settings.AllowMergeCommit
		patch.AllowMergeCommit = &desired
		if current.AllowMergeCommit != desired {
			result.Changes = append(result.Changes, Change{Field: "allow_merge_commit", Current: current.AllowMergeCommit, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.AllowSquashMerge != nil {
		desired := *s.config.Settings.AllowSquashMerge
		patch.AllowSquashMerge = &desired
		if current.AllowSquashMerge != desired {
			result.Changes = append(result.Changes, Change{Field: "allow_squash_merge", Current: current.AllowSquashMerge, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.AllowRebaseMerge != nil {
		desired := *s.config.Settings.AllowRebaseMerge
		patch.AllowRebaseMerge = &desired
		if current.AllowRebaseMerge != desired {
			result.Changes = append(result.Changes, Change{Field: "allow_rebase_merge", Current: current.AllowRebaseMerge, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.DeleteBranchOnMerge != nil {
		desired := *s.config.Settings.DeleteBranchOnMerge
		patch.DeleteBranchOnMerge = &desired
		if current.DeleteBranchOnMerge != desired {
			result.Changes = append(result.Changes, Change{Field: "delete_branch_on_merge", Current: current.DeleteBranchOnMerge, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.SquashMergeCommitTitle != nil {
		desired := *s.config.Settings.SquashMergeCommitTitle
		patch.SquashMergeCommitTitle = &desired
		if current.SquashMergeCommitTitle != desired {
			result.Changes = append(result.Changes, Change{Field: "squash_merge_commit_title", Current: current.SquashMergeCommitTitle, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.SquashMergeCommitMessage != nil {
		desired := *s.config.Settings.SquashMergeCommitMessage
		patch.SquashMergeCommitMessage = &desired
		if current.SquashMergeCommitMessage != desired {
			result.Changes = append(result.Changes, Change{Field: "squash_merge_commit_message", Current: current.SquashMergeCommitMessage, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.MergeCommitTitle != nil {
		desired := *s.config.Settings.MergeCommitTitle
		patch.MergeCommitTitle = &desired
		if current.MergeCommitTitle != desired {
			result.Changes = append(result.Changes, Change{Field: "merge_commit_title", Current: current.MergeCommitTitle, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.MergeCommitMessage != nil {
		desired := *s.config.Settings.MergeCommitMessage
		patch.MergeCommitMessage = &desired
		if current.MergeCommitMessage != desired {
			result.Changes = append(result.Changes, Change{Field: "merge_commit_message", Current: current.MergeCommitMessage, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.HasIssues != nil {
		desired := *s.config.Settings.HasIssues
		patch.HasIssues = &desired
		if current.HasIssues != desired {
			result.Changes = append(result.Changes, Change{Field: "has_issues", Current: current.HasIssues, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.HasProjects != nil {
		desired := *s.config.Settings.HasProjects
		patch.HasProjects = &desired
		if current.HasProjects != desired {
			result.Changes = append(result.Changes, Change{Field: "has_projects", Current: current.HasProjects, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.HasWiki != nil {
		desired := *s.config.Settings.HasWiki
		patch.HasWiki = &desired
		if current.HasWiki != desired {
			result.Changes = append(result.Changes, Change{Field: "has_wiki", Current: current.HasWiki, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.HasDiscussions != nil {
		desired := *s.config.Settings.HasDiscussions
		patch.HasDiscussions = &desired
		if current.HasDiscussions != desired {
			result.Changes = append(result.Changes, Change{Field: "has_discussions", Current: current.HasDiscussions, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.Archived != nil {
		desired := *s.config.Settings.Archived
		patch.Archived = &desired
		if current.Archived != desired {
			result.Changes = append(result.Changes, Change{Field: "archived", Current: current.Archived, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.AllowUpdateBranch != nil {
		desired := *s.config.Settings.AllowUpdateBranch
		patch.AllowUpdateBranch = &desired
		if current.AllowUpdateBranch != desired {
			result.Changes = append(result.Changes, Change{Field: "allow_update_branch", Current: current.AllowUpdateBranch, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.WebCommitSignoffRequired != nil {
		desired := *s.config.Settings.WebCommitSignoffRequired
		patch.WebCommitSignoffRequired = &desired
		if current.WebCommitSignoffRequired != desired {
			result.Changes = append(result.Changes, Change{Field: "web_commit_signoff_required", Current: current.WebCommitSignoffRequired, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.AllowForking != nil {
		desired := *s.config.Settings.AllowForking
		patch.AllowForking = &desired
		if current.AllowForking != desired {
			result.Changes = append(result.Changes, Change{Field: "allow_forking", Current: current.AllowForking, Desired: desired})
			changed = true
		}
	}
	if s.config.Settings.Visibility != nil {
		desiredPrivate := *s.config.Settings.Visibility == "private"
		patch.Private = &desiredPrivate
		if current.Private != desiredPrivate {
			result.Changes = append(result.Changes, Change{
				Field:   "visibility",
				Current: map[bool]string{true: "private", false: "public"}[current.Private],
				Desired: map[bool]string{true: "private", false: "public"}[desiredPrivate],
			})
			changed = true
		}
	}

	if !dryRun && changed {
		if err := s.client.UpdateRepositorySettings(repo.Owner, repo.Name, patch); err != nil {
			result.Error = fmt.Errorf("failed to update settings: %w", err)
			return result
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

	return result
}

// syncBranchProtection syncs branch protection settings
func (s *Syncer) syncBranchProtection(repo config.Repository, dryRun bool) Result {
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

	if current.Enabled != desired.Enabled {
		result.Changes = append(result.Changes, Change{Field: "branch_protection", Current: map[bool]string{true: "enabled", false: "disabled"}[current.Enabled], Desired: map[bool]string{true: "enabled", false: "disabled"}[desired.Enabled]})
	}

	// If branch protection is being disabled, the only action is removal.
	if !bp.Enabled {
		if !dryRun && len(result.Changes) > 0 {
			if err := s.client.UpdateBranchProtection(repo.Owner, repo.Name, &desired); err != nil {
				result.Error = fmt.Errorf("failed to update branch protection: %w", err)
			}
		}
		return result
	}

	// Pull request review requirements
	if bp.RequiredReviews != nil || bp.DismissStaleReviews != nil || bp.RequireCodeOwnerReviews != nil {
		desired.PullRequestReviewsEnabled = true
	}
	if bp.RequiredReviews != nil {
		if desired.RequiredReviews != *bp.RequiredReviews {
			result.Changes = append(result.Changes, Change{Field: "required_reviews", Current: desired.RequiredReviews, Desired: *bp.RequiredReviews})
		}
		desired.RequiredReviews = *bp.RequiredReviews
	}
	if bp.DismissStaleReviews != nil {
		if desired.DismissStaleReviews != *bp.DismissStaleReviews {
			result.Changes = append(result.Changes, Change{Field: "dismiss_stale_reviews", Current: desired.DismissStaleReviews, Desired: *bp.DismissStaleReviews})
		}
		desired.DismissStaleReviews = *bp.DismissStaleReviews
	}
	if bp.RequireCodeOwnerReviews != nil {
		if desired.RequireCodeOwnerReviews != *bp.RequireCodeOwnerReviews {
			result.Changes = append(result.Changes, Change{Field: "require_code_owner_reviews", Current: desired.RequireCodeOwnerReviews, Desired: *bp.RequireCodeOwnerReviews})
		}
		desired.RequireCodeOwnerReviews = *bp.RequireCodeOwnerReviews
	}

	// Status checks
	if bp.RequireStatusChecks != nil {
		if desired.StatusChecksEnabled != *bp.RequireStatusChecks {
			result.Changes = append(result.Changes, Change{Field: "require_status_checks", Current: desired.StatusChecksEnabled, Desired: *bp.RequireStatusChecks})
		}
		desired.StatusChecksEnabled = *bp.RequireStatusChecks
	}
	if bp.RequireBranchesUpToDate != nil {
		if desired.RequireBranchesUpToDate != *bp.RequireBranchesUpToDate {
			result.Changes = append(result.Changes, Change{Field: "require_branches_up_to_date", Current: desired.RequireBranchesUpToDate, Desired: *bp.RequireBranchesUpToDate})
		}
		desired.RequireBranchesUpToDate = *bp.RequireBranchesUpToDate
	}
	if len(bp.StatusCheckContexts) > 0 {
		if !reflect.DeepEqual(desired.StatusCheckContexts, bp.StatusCheckContexts) {
			result.Changes = append(result.Changes, Change{Field: "status_check_contexts", Current: desired.StatusCheckContexts, Desired: bp.StatusCheckContexts})
		}
		desired.StatusCheckContexts = bp.StatusCheckContexts
		desired.StatusCheckChecks = nil
	}

	if desired.StatusChecksEnabled {
		if len(desired.StatusCheckContexts) == 0 && len(desired.StatusCheckChecks) == 0 {
			result.Error = fmt.Errorf("branch protection %s: require_status_checks is true but no status_check_contexts are configured and none exist on the branch", repo.FullName())
			return result
		}
	}

	if bp.IncludeAdmins != nil {
		if desired.IncludeAdmins != *bp.IncludeAdmins {
			result.Changes = append(result.Changes, Change{Field: "include_admins", Current: desired.IncludeAdmins, Desired: *bp.IncludeAdmins})
		}
		desired.IncludeAdmins = *bp.IncludeAdmins
	}
	if bp.RequireLinearHistory != nil {
		if desired.RequireLinearHistory != *bp.RequireLinearHistory {
			result.Changes = append(result.Changes, Change{Field: "require_linear_history", Current: desired.RequireLinearHistory, Desired: *bp.RequireLinearHistory})
		}
		desired.RequireLinearHistory = *bp.RequireLinearHistory
	}
	if bp.RequireSignedCommits != nil {
		if desired.RequireSignedCommits != *bp.RequireSignedCommits {
			result.Changes = append(result.Changes, Change{Field: "require_signed_commits", Current: desired.RequireSignedCommits, Desired: *bp.RequireSignedCommits})
		}
		desired.RequireSignedCommits = *bp.RequireSignedCommits
	}
	if bp.RequireConversationResolution != nil {
		if desired.RequireConversationResolution != *bp.RequireConversationResolution {
			result.Changes = append(result.Changes, Change{Field: "require_conversation_resolution", Current: desired.RequireConversationResolution, Desired: *bp.RequireConversationResolution})
		}
		desired.RequireConversationResolution = *bp.RequireConversationResolution
	}
	if bp.AllowForcePushes != nil {
		if desired.AllowForcePushes != *bp.AllowForcePushes {
			result.Changes = append(result.Changes, Change{Field: "allow_force_pushes", Current: desired.AllowForcePushes, Desired: *bp.AllowForcePushes})
		}
		desired.AllowForcePushes = *bp.AllowForcePushes
	}
	if bp.AllowDeletions != nil {
		if desired.AllowDeletions != *bp.AllowDeletions {
			result.Changes = append(result.Changes, Change{Field: "allow_deletions", Current: desired.AllowDeletions, Desired: *bp.AllowDeletions})
		}
		desired.AllowDeletions = *bp.AllowDeletions
	}

	// Apply changes if not dry-run
	if !dryRun && len(result.Changes) > 0 {
		if err := s.client.UpdateBranchProtection(repo.Owner, repo.Name, &desired); err != nil {
			result.Error = fmt.Errorf("failed to update branch protection: %w", err)
			return result
		}
	}

	return result
}
