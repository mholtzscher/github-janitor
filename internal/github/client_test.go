package github

import "testing"

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
	if req.RequiredStatusChecks.Contexts == nil || len(*req.RequiredStatusChecks.Contexts) != 1 || (*req.RequiredStatusChecks.Contexts)[0] != "ci/test" {
		t.Fatalf("Contexts = %v; want [ci/test]", req.RequiredStatusChecks.Contexts)
	}
	if req.RequiredStatusChecks.Checks == nil || len(*req.RequiredStatusChecks.Checks) != 0 {
		t.Fatalf("Checks = %v; want empty slice", req.RequiredStatusChecks.Checks)
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
