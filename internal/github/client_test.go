package github //nolint:testpackage // Tests internal implementation details

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gogithub "github.com/google/go-github/v82/github"
)

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
	if req.RequiredStatusChecks.Contexts == nil || len(*req.RequiredStatusChecks.Contexts) != 1 ||
		(*req.RequiredStatusChecks.Contexts)[0] != "ci/test" {
		t.Fatalf("Contexts = %v; want [ci/test]", req.RequiredStatusChecks.Contexts)
	}
	if req.RequiredStatusChecks.Checks == nil || len(*req.RequiredStatusChecks.Checks) != 0 {
		t.Fatalf("Checks = %v; want empty slice", req.RequiredStatusChecks.Checks)
	}

	if req.RequiredPullRequestReviews == nil {
		t.Fatal("RequiredPullRequestReviews is nil; want non-nil")
	}
	if req.RequiredPullRequestReviews.RequiredApprovingReviewCount != 2 {
		t.Fatalf(
			"RequiredApprovingReviewCount = %d; want 2",
			req.RequiredPullRequestReviews.RequiredApprovingReviewCount,
		)
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

func TestSetActionsSecret_SendsExpectedJSONPayload(t *testing.T) {
	t.Parallel()

	const (
		owner          = "mholtzscher"
		repo           = "aerospace-utils"
		secretName     = "TEST"
		encryptedValue = "encrypted-123"
		keyID          = "key-123"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.Method != http.MethodPut {
			t.Fatalf("method = %q; want %q", r.Method, http.MethodPut)
		}

		wantPath := "/repos/" + owner + "/" + repo + "/actions/secrets/" + secretName
		if r.URL.Path != wantPath {
			t.Fatalf("path = %q; want %q", r.URL.Path, wantPath)
		}

		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/json") {
			t.Fatalf("content-type = %q; want application/json", got)
		}

		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}

		if payload["encrypted_value"] != encryptedValue {
			t.Fatalf("encrypted_value = %q; want %q", payload["encrypted_value"], encryptedValue)
		}

		if payload["key_id"] != keyID {
			t.Fatalf("key_id = %q; want %q", payload["key_id"], keyID)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	ghClient := gogithub.NewClient(server.Client())
	baseURL, err := ghClient.BaseURL.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("parse base URL: %v", err)
	}
	ghClient.BaseURL = baseURL

	client := &Client{client: ghClient, ctx: context.Background()}

	err = client.SetActionsSecret(owner, repo, secretName, encryptedValue, keyID)
	if err != nil {
		t.Fatalf("SetActionsSecret returned error: %v", err)
	}
}
