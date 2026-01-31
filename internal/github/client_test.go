package github

import "testing"

func TestBuildRepositoryEditRequest_Partial(t *testing.T) {
	allowMerge := true
	patch := &RepositorySettingsPatch{AllowMergeCommit: &allowMerge}

	req := buildRepositoryEditRequest(patch)
	if req.AllowMergeCommit == nil || *req.AllowMergeCommit != true {
		t.Fatalf("AllowMergeCommit = %v; want true", req.AllowMergeCommit)
	}
	if req.AllowSquashMerge != nil {
		t.Fatalf("AllowSquashMerge = %v; want nil", req.AllowSquashMerge)
	}
	if req.AllowRebaseMerge != nil {
		t.Fatalf("AllowRebaseMerge = %v; want nil", req.AllowRebaseMerge)
	}
	if req.Private != nil {
		t.Fatalf("Private = %v; want nil", req.Private)
	}
	if req.DeleteBranchOnMerge != nil {
		t.Fatalf("DeleteBranchOnMerge = %v; want nil", req.DeleteBranchOnMerge)
	}
}

func TestBuildProtectionRequest_StatusChecks(t *testing.T) {
	p := &BranchProtectionInfo{
		PullRequestReviewsEnabled: true,
		RequiredReviews:           2,
		DismissStaleReviews:       true,
		RequireCodeOwnerReviews:   false,

		StatusChecksEnabled:     true,
		RequireBranchesUpToDate: false,
		StatusCheckContexts:     []string{"ci/test"},

		IncludeAdmins:                 false,
		RequireLinearHistory:          true,
		AllowForcePushes:              false,
		AllowDeletions:                false,
		RequireConversationResolution: true,
	}

	req := buildProtectionRequest(p)
	if req.RequiredStatusChecks == nil {
		t.Fatal("RequiredStatusChecks is nil; want non-nil")
	}
	if req.RequiredStatusChecks.Strict != false {
		t.Fatalf("Strict = %v; want false", req.RequiredStatusChecks.Strict)
	}
	if len(req.RequiredStatusChecks.Contexts) != 1 || req.RequiredStatusChecks.Contexts[0] != "ci/test" {
		t.Fatalf("Contexts = %v; want [ci/test]", req.RequiredStatusChecks.Contexts)
	}

	if req.RequiredPullRequestReviews == nil {
		t.Fatal("RequiredPullRequestReviews is nil; want non-nil")
	}
	if req.RequiredPullRequestReviews.RequiredApprovingReviewCount != 2 {
		t.Fatalf("RequiredApprovingReviewCount = %d; want 2", req.RequiredPullRequestReviews.RequiredApprovingReviewCount)
	}
	if req.RequiredPullRequestReviews.DismissStaleReviews != true {
		t.Fatalf("DismissStaleReviews = %v; want true", req.RequiredPullRequestReviews.DismissStaleReviews)
	}

	if req.RequireLinearHistory == nil || *req.RequireLinearHistory != true {
		t.Fatalf("RequireLinearHistory = %v; want true", req.RequireLinearHistory)
	}
	if req.RequiredConversationResolution == nil || *req.RequiredConversationResolution != true {
		t.Fatalf("RequiredConversationResolution = %v; want true", req.RequiredConversationResolution)
	}
}
