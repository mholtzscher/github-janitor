package sync //nolint:testpackage // Tests internal implementation details

import (
	"errors"
	"reflect"
	"testing"

	gogithub "github.com/google/go-github/v82/github"

	"github.com/mholtzscher/github-janitor/internal/config"
	"github.com/mholtzscher/github-janitor/internal/github"
)

func boolPtr(v bool) *bool       { return &v }
func stringPtr(v string) *string { return &v }
func intPtr(v int) *int          { return &v }

type fakeGitHubClient struct {
	getRepoCalls      int
	updateRepoCalls   int
	getBranchCalls    int
	updateBranchCalls int
	lastRepoOwner     string
	lastRepoName      string
	lastRepoPatch     *gogithub.Repository
	lastBPPattern     string
	lastBPProtection  *github.BranchProtectionInfo
	getRepoResp       *github.RepositoryInfo
	getRepoErr        error
	updateRepoErr     error
	getBranchResp     *github.BranchProtectionInfo
	getBranchErr      error
	updateBranchErr   error
}

func (f *fakeGitHubClient) GetRepository(owner, name string) (*github.RepositoryInfo, error) {
	f.getRepoCalls++
	f.lastRepoOwner = owner
	f.lastRepoName = name
	return f.getRepoResp, f.getRepoErr
}

func (f *fakeGitHubClient) UpdateRepositorySettings(owner, name string, patch *gogithub.Repository) error {
	f.updateRepoCalls++
	f.lastRepoOwner = owner
	f.lastRepoName = name
	f.lastRepoPatch = patch
	return f.updateRepoErr
}

func (f *fakeGitHubClient) GetBranchProtection(owner, name, pattern string) (*github.BranchProtectionInfo, error) {
	f.getBranchCalls++
	f.lastRepoOwner = owner
	f.lastRepoName = name
	f.lastBPPattern = pattern
	return f.getBranchResp, f.getBranchErr
}

func (f *fakeGitHubClient) UpdateBranchProtection(owner, name string, protection *github.BranchProtectionInfo) error {
	f.updateBranchCalls++
	f.lastRepoOwner = owner
	f.lastRepoName = name
	f.lastBPProtection = protection
	return f.updateBranchErr
}

func changeByField(t *testing.T, changes []Change) map[string]Change {
	t.Helper()
	got := make(map[string]Change, len(changes))
	for _, c := range changes {
		if _, ok := got[c.Field]; ok {
			t.Fatalf("duplicate change for field %q", c.Field)
		}
		got[c.Field] = c
	}
	return got
}

func TestApplySetting(t *testing.T) { //nolint:gocognit // Table-driven tests with subtests
	t.Run("configured_nil_noop", func(t *testing.T) {
		result := &Result{}
		var patchVal *bool
		current := false

		changed := applySetting(result, "x", (*bool)(nil), current, &patchVal)
		if changed {
			t.Fatal("changed = true; want false")
		}
		if patchVal != nil {
			t.Fatalf("patchVal = %v; want nil", *patchVal)
		}
		if len(result.Changes) != 0 {
			t.Fatalf("len(Changes) = %d; want 0", len(result.Changes))
		}
	})

	t.Run("configured_sets_patch_and_tracks_change", func(t *testing.T) {
		result := &Result{}
		var patchVal *bool
		configured := true
		current := false

		changed := applySetting(result, "allow_merge_commit", &configured, current, &patchVal)
		if !changed {
			t.Fatal("changed = false; want true")
		}
		if patchVal == nil || *patchVal != true {
			t.Fatalf("patchVal = %v; want true", patchVal)
		}
		if len(result.Changes) != 1 {
			t.Fatalf("len(Changes) = %d; want 1", len(result.Changes))
		}
		c := result.Changes[0]
		if c.Field != "allow_merge_commit" {
			t.Fatalf("Field = %q; want %q", c.Field, "allow_merge_commit")
		}
		if c.Current != false || c.Desired != true {
			t.Fatalf("Current/Desired = %v/%v; want false/true", c.Current, c.Desired)
		}
	})

	t.Run("configured_sets_patch_but_no_change_when_equal", func(t *testing.T) {
		result := &Result{}
		var patchVal *bool
		configured := true
		current := true

		changed := applySetting(result, "allow_merge_commit", &configured, current, &patchVal)
		if changed {
			t.Fatal("changed = true; want false")
		}
		if patchVal == nil || *patchVal != true {
			t.Fatalf("patchVal = %v; want true", patchVal)
		}
		if len(result.Changes) != 0 {
			t.Fatalf("len(Changes) = %d; want 0", len(result.Changes))
		}
	})
}

func TestApplyDesiredSetting(t *testing.T) { //nolint:gocognit // Table-driven tests with subtests
	t.Run("configured_nil_noop", func(t *testing.T) {
		result := &Result{}
		desired := 1

		changed := applyDesiredSetting(result, "required_reviews", (*int)(nil), &desired)
		if changed {
			t.Fatal("changed = true; want false")
		}
		if desired != 1 {
			t.Fatalf("desired = %d; want 1", desired)
		}
		if len(result.Changes) != 0 {
			t.Fatalf("len(Changes) = %d; want 0", len(result.Changes))
		}
	})

	t.Run("updates_desired_and_tracks_change", func(t *testing.T) {
		result := &Result{}
		desired := 1
		configured := 2

		changed := applyDesiredSetting(result, "required_reviews", &configured, &desired)
		if !changed {
			t.Fatal("changed = false; want true")
		}
		if desired != 2 {
			t.Fatalf("desired = %d; want 2", desired)
		}
		if len(result.Changes) != 1 {
			t.Fatalf("len(Changes) = %d; want 1", len(result.Changes))
		}
		c := result.Changes[0]
		if c.Field != "required_reviews" {
			t.Fatalf("Field = %q; want %q", c.Field, "required_reviews")
		}
		if c.Current != 1 || c.Desired != 2 {
			t.Fatalf("Current/Desired = %v/%v; want 1/2", c.Current, c.Desired)
		}
	})

	t.Run("updates_desired_without_change_when_equal", func(t *testing.T) {
		result := &Result{}
		desired := 2
		configured := 2

		changed := applyDesiredSetting(result, "required_reviews", &configured, &desired)
		if changed {
			t.Fatal("changed = true; want false")
		}
		if desired != 2 {
			t.Fatalf("desired = %d; want 2", desired)
		}
		if len(result.Changes) != 0 {
			t.Fatalf("len(Changes) = %d; want 0", len(result.Changes))
		}
	})
}

func TestApplyDesiredStringSlice(t *testing.T) {
	t.Run("configured_empty_noop", func(t *testing.T) {
		result := &Result{}
		desired := []string{"a"}
		configured := []string{}

		changed := applyDesiredStringSlice(result, "status_check_contexts", configured, &desired)
		if changed {
			t.Fatal("changed = true; want false")
		}
		if !reflect.DeepEqual(desired, []string{"a"}) {
			t.Fatalf("desired = %v; want [a]", desired)
		}
		if len(result.Changes) != 0 {
			t.Fatalf("len(Changes) = %d; want 0", len(result.Changes))
		}
	})

	t.Run("copies_configured_and_tracks_change", func(t *testing.T) {
		result := &Result{}
		desired := []string{"old"}
		configured := []string{"ci/test"}

		changed := applyDesiredStringSlice(result, "status_check_contexts", configured, &desired)
		if !changed {
			t.Fatal("changed = false; want true")
		}
		if len(result.Changes) != 1 {
			t.Fatalf("len(Changes) = %d; want 1", len(result.Changes))
		}
		c := result.Changes[0]
		if c.Field != "status_check_contexts" {
			t.Fatalf("Field = %q; want %q", c.Field, "status_check_contexts")
		}
		if !reflect.DeepEqual(c.Current, []string{"old"}) || !reflect.DeepEqual(c.Desired, []string{"ci/test"}) {
			t.Fatalf("Current/Desired = %v/%v; want [old]/[ci/test]", c.Current, c.Desired)
		}

		if !reflect.DeepEqual(desired, []string{"ci/test"}) {
			t.Fatalf("desired = %v; want [ci/test]", desired)
		}
		configured[0] = "mutated"
		if !reflect.DeepEqual(desired, []string{"ci/test"}) {
			t.Fatalf("desired shares backing array with configured: %v", desired)
		}
	})
}

func TestSyncRepository_SkipsWhenRepoDoesNotExist(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}

	fake := &fakeGitHubClient{getRepoResp: &github.RepositoryInfo{Owner: "o", Name: "r", Exists: false}}
	s := &Syncer{client: fake, config: &config.Config{Repositories: []config.Repository{repo}}}

	result := s.syncRepository(repo, false)
	if result.Error != nil {
		t.Fatalf("Error = %v; want nil", result.Error)
	}
	if result.Exists {
		t.Fatal("Exists = true; want false")
	}
	if fake.updateRepoCalls != 0 {
		t.Fatalf("updateRepoCalls = %d; want 0", fake.updateRepoCalls)
	}
}

func TestSyncRepository_DryRunDoesNotUpdate(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}

	fake := &fakeGitHubClient{
		getRepoResp: &github.RepositoryInfo{
			Owner:            "o",
			Name:             "r",
			Exists:           true,
			AllowMergeCommit: false,
			Private:          false,
		},
	}
	cfg := &config.Config{
		Repositories: []config.Repository{repo},
		Settings: config.Settings{
			AllowMergeCommit: boolPtr(true),
			Visibility:       stringPtr("private"),
		},
	}

	s := &Syncer{client: fake, config: cfg}
	result := s.syncRepository(repo, true)
	if result.Error != nil {
		t.Fatalf("Error = %v; want nil", result.Error)
	}
	if !result.Exists {
		t.Fatal("Exists = false; want true")
	}
	if fake.updateRepoCalls != 0 {
		t.Fatalf("updateRepoCalls = %d; want 0", fake.updateRepoCalls)
	}

	got := changeByField(t, result.Changes)
	if c, ok := got["allow_merge_commit"]; !ok || c.Current != false || c.Desired != true {
		t.Fatalf("allow_merge_commit change = %v; want false -> true", c)
	}
	if c, ok := got["visibility"]; !ok || c.Current != "public" || c.Desired != "private" {
		t.Fatalf("visibility change = %v; want public -> private", c)
	}
}

func TestSyncRepository_AppliesUpdateWhenChanged(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}

	fake := &fakeGitHubClient{
		getRepoResp: &github.RepositoryInfo{
			Owner:            "o",
			Name:             "r",
			Exists:           true,
			AllowMergeCommit: false,
			AllowSquashMerge: true,
			Private:          false,
		},
	}
	cfg := &config.Config{
		Repositories: []config.Repository{repo},
		Settings: config.Settings{
			AllowMergeCommit: boolPtr(true),
			AllowSquashMerge: boolPtr(true),
			Visibility:       stringPtr("private"),
		},
	}

	s := &Syncer{client: fake, config: cfg}
	result := s.syncRepository(repo, false)
	if result.Error != nil {
		t.Fatalf("Error = %v; want nil", result.Error)
	}
	if fake.updateRepoCalls != 1 {
		t.Fatalf("updateRepoCalls = %d; want 1", fake.updateRepoCalls)
	}
	if fake.lastRepoPatch == nil {
		t.Fatal("lastRepoPatch is nil")
	}
	if fake.lastRepoPatch.AllowMergeCommit == nil || *fake.lastRepoPatch.AllowMergeCommit != true {
		t.Fatalf("AllowMergeCommit patch = %v; want true", fake.lastRepoPatch.AllowMergeCommit)
	}
	if fake.lastRepoPatch.AllowSquashMerge == nil || *fake.lastRepoPatch.AllowSquashMerge != true {
		t.Fatalf("AllowSquashMerge patch = %v; want true", fake.lastRepoPatch.AllowSquashMerge)
	}
	if fake.lastRepoPatch.Private == nil || *fake.lastRepoPatch.Private != true {
		t.Fatalf("Private patch = %v; want true", fake.lastRepoPatch.Private)
	}
}

func TestSyncRepository_PropagatesUpdateError(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}
	boom := errors.New("boom")

	fake := &fakeGitHubClient{
		getRepoResp:   &github.RepositoryInfo{Owner: "o", Name: "r", Exists: true, AllowMergeCommit: false},
		updateRepoErr: boom,
	}
	cfg := &config.Config{
		Repositories: []config.Repository{repo},
		Settings:     config.Settings{AllowMergeCommit: boolPtr(true)},
	}

	s := &Syncer{client: fake, config: cfg}
	result := s.syncRepository(repo, false)
	if result.Error == nil {
		t.Fatal("Error = nil; want error")
	}
	if !errors.Is(result.Error, boom) {
		t.Fatalf("Error = %v; want wrapped %v", result.Error, boom)
	}
}

func TestSyncBranchProtection_DisableRemovesProtection(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}

	fake := &fakeGitHubClient{getBranchResp: &github.BranchProtectionInfo{Enabled: true, Pattern: "main"}}
	cfg := &config.Config{
		Repositories: []config.Repository{repo},
		Settings: config.Settings{
			BranchProtection: &config.BranchProtection{Enabled: false, Pattern: "main"},
		},
	}

	s := &Syncer{client: fake, config: cfg}
	result := s.syncBranchProtection(repo, false)
	if result.Error != nil {
		t.Fatalf("Error = %v; want nil", result.Error)
	}
	if fake.updateBranchCalls != 1 {
		t.Fatalf("updateBranchCalls = %d; want 1", fake.updateBranchCalls)
	}
	if fake.lastBPProtection == nil || fake.lastBPProtection.Enabled != false {
		t.Fatalf("lastBPProtection.Enabled = %v; want false", fake.lastBPProtection)
	}

	got := changeByField(t, result.Changes)
	if c, ok := got["branch_protection"]; !ok || c.Current != "enabled" || c.Desired != "disabled" {
		t.Fatalf("branch_protection change = %v; want enabled -> disabled", c)
	}
}

func TestSyncBranchProtection_ErrorsWhenStatusChecksRequiredButNoneConfiguredOrExisting(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}

	fake := &fakeGitHubClient{getBranchResp: &github.BranchProtectionInfo{Enabled: false, Pattern: "main"}}
	cfg := &config.Config{
		Repositories: []config.Repository{repo},
		Settings: config.Settings{
			BranchProtection: &config.BranchProtection{
				Enabled:             true,
				Pattern:             "main",
				RequireStatusChecks: boolPtr(true),
			},
		},
	}

	s := &Syncer{client: fake, config: cfg}
	result := s.syncBranchProtection(repo, false)
	if result.Error == nil {
		t.Fatal("Error = nil; want error")
	}
	if fake.updateBranchCalls != 0 {
		t.Fatalf("updateBranchCalls = %d; want 0", fake.updateBranchCalls)
	}
}

func TestSyncBranchProtection_ConfiguredContextsOverrideAndClearChecks(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}

	fake := &fakeGitHubClient{
		getBranchResp: &github.BranchProtectionInfo{
			Enabled:             true,
			Pattern:             "main",
			StatusChecksEnabled: true,
			StatusCheckContexts: []string{"old"},
			StatusCheckChecks:   []*gogithub.RequiredStatusCheck{{}},
		},
	}
	cfg := &config.Config{
		Repositories: []config.Repository{repo},
		Settings: config.Settings{
			BranchProtection: &config.BranchProtection{
				Enabled:             true,
				Pattern:             "main",
				RequireStatusChecks: boolPtr(true),
				StatusCheckContexts: []string{"ci/test"},
			},
		},
	}

	s := &Syncer{client: fake, config: cfg}
	result := s.syncBranchProtection(repo, false)
	if result.Error != nil {
		t.Fatalf("Error = %v; want nil", result.Error)
	}
	if fake.updateBranchCalls != 1 {
		t.Fatalf("updateBranchCalls = %d; want 1", fake.updateBranchCalls)
	}
	if fake.lastBPProtection == nil {
		t.Fatal("lastBPProtection is nil")
	}
	if !reflect.DeepEqual(fake.lastBPProtection.StatusCheckContexts, []string{"ci/test"}) {
		t.Fatalf("StatusCheckContexts = %v; want [ci/test]", fake.lastBPProtection.StatusCheckContexts)
	}
	if fake.lastBPProtection.StatusCheckChecks != nil {
		t.Fatalf("StatusCheckChecks = %v; want nil", fake.lastBPProtection.StatusCheckChecks)
	}

	got := changeByField(t, result.Changes)
	if c, ok := got["status_check_contexts"]; !ok || !reflect.DeepEqual(c.Current, []string{"old"}) ||
		!reflect.DeepEqual(c.Desired, []string{"ci/test"}) {
		t.Fatalf("status_check_contexts change = %v; want [old] -> [ci/test]", c)
	}
}

func TestSyncBranchProtection_ConfiguredReviewsEnablePRReviews(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}

	fake := &fakeGitHubClient{
		getBranchResp: &github.BranchProtectionInfo{
			Enabled:                   true,
			Pattern:                   "main",
			PullRequestReviewsEnabled: false,
			RequiredReviews:           0,
		},
	}
	cfg := &config.Config{
		Repositories: []config.Repository{repo},
		Settings: config.Settings{
			BranchProtection: &config.BranchProtection{
				Enabled:             true,
				Pattern:             "main",
				RequiredReviews:     intPtr(2),
				DismissStaleReviews: boolPtr(true),
			},
		},
	}

	s := &Syncer{client: fake, config: cfg}
	result := s.syncBranchProtection(repo, false)
	if result.Error != nil {
		t.Fatalf("Error = %v; want nil", result.Error)
	}
	if fake.updateBranchCalls != 1 {
		t.Fatalf("updateBranchCalls = %d; want 1", fake.updateBranchCalls)
	}
	if fake.lastBPProtection == nil || fake.lastBPProtection.PullRequestReviewsEnabled != true {
		t.Fatalf("PullRequestReviewsEnabled = %v; want true", fake.lastBPProtection)
	}

	got := changeByField(t, result.Changes)
	if c, ok := got["required_reviews"]; !ok || c.Current != 0 || c.Desired != 2 {
		t.Fatalf("required_reviews change = %v; want 0 -> 2", c)
	}
	if c, ok := got["dismiss_stale_reviews"]; !ok || c.Current != false || c.Desired != true {
		t.Fatalf("dismiss_stale_reviews change = %v; want false -> true", c)
	}
}

func TestSyncBranchProtection_ConfiguredZeroReviewsEnablesPRReviews(t *testing.T) {
	repo := config.Repository{Owner: "o", Name: "r"}

	fake := &fakeGitHubClient{
		getBranchResp: &github.BranchProtectionInfo{
			Enabled:                   true,
			Pattern:                   "main",
			PullRequestReviewsEnabled: false,
			RequiredReviews:           0,
		},
	}
	cfg := &config.Config{
		Repositories: []config.Repository{repo},
		Settings: config.Settings{
			BranchProtection: &config.BranchProtection{
				Enabled:         true,
				Pattern:         "main",
				RequiredReviews: intPtr(0),
			},
		},
	}

	s := &Syncer{client: fake, config: cfg}
	result := s.syncBranchProtection(repo, false)
	if result.Error != nil {
		t.Fatalf("Error = %v; want nil", result.Error)
	}
	if fake.updateBranchCalls != 1 {
		t.Fatalf("updateBranchCalls = %d; want 1", fake.updateBranchCalls)
	}
	if fake.lastBPProtection == nil || fake.lastBPProtection.PullRequestReviewsEnabled != true {
		t.Fatalf("PullRequestReviewsEnabled = %v; want true", fake.lastBPProtection)
	}

	got := changeByField(t, result.Changes)
	if c, ok := got["pull_request_reviews_enabled"]; !ok || c.Current != false || c.Desired != true {
		t.Fatalf("pull_request_reviews_enabled change = %v; want false -> true", c)
	}
}
