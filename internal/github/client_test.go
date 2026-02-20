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

func testClientForServer(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	ghClient := gogithub.NewClient(server.Client())
	baseURL, err := ghClient.BaseURL.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("parse base URL: %v", err)
	}
	ghClient.BaseURL = baseURL

	return &Client{client: ghClient, ctx: context.Background()}
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

	client := testClientForServer(t, server)

	err := client.SetActionsSecret(owner, repo, secretName, encryptedValue, keyID)
	if err != nil {
		t.Fatalf("SetActionsSecret returned error: %v", err)
	}
}

func TestGetSecuritySettings_ReadsEndpoints(t *testing.T) {
	t.Parallel()

	const (
		owner = "mholtzscher"
		repo  = "aerospace-utils"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/"+owner+"/"+repo+"/vulnerability-alerts":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/repos/"+owner+"/"+repo+"/automated-security-fixes":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"enabled":true,"paused":false}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := testClientForServer(t, server)

	settings, err := client.GetSecuritySettings(owner, repo)
	if err != nil {
		t.Fatalf("GetSecuritySettings returned error: %v", err)
	}

	if !settings.DependabotAlerts {
		t.Fatal("DependabotAlerts = false; want true")
	}
	if !settings.DependabotSecurityUpdates {
		t.Fatal("DependabotSecurityUpdates = false; want true")
	}
	if settings.DependabotSecurityUpdatesPaused {
		t.Fatal("DependabotSecurityUpdatesPaused = true; want false")
	}
}

func TestSetDependabotAlerts_UsesVulnerabilityAlertEndpoint(t *testing.T) {
	t.Parallel()
	testToggleEndpoint(t, "vulnerability-alerts", func(client *Client, owner, repo string, enabled bool) error {
		return client.SetDependabotAlerts(owner, repo, enabled)
	})
}

func TestSetDependabotSecurityUpdates_UsesAutomatedFixesEndpoint(t *testing.T) {
	t.Parallel()
	testToggleEndpoint(t, "automated-security-fixes", func(client *Client, owner, repo string, enabled bool) error {
		return client.SetDependabotSecurityUpdates(owner, repo, enabled)
	})
}

func testToggleEndpoint(t *testing.T, endpoint string, setter func(*Client, string, string, bool) error) {
	t.Helper()

	const (
		owner = "mholtzscher"
		repo  = "aerospace-utils"
	)

	var got []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		got = append(got, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := testClientForServer(t, server)

	if err := setter(client, owner, repo, true); err != nil {
		t.Fatalf("setter(true) returned error: %v", err)
	}
	if err := setter(client, owner, repo, false); err != nil {
		t.Fatalf("setter(false) returned error: %v", err)
	}

	want := []string{
		http.MethodPut + " /repos/" + owner + "/" + repo + "/" + endpoint,
		http.MethodDelete + " /repos/" + owner + "/" + repo + "/" + endpoint,
	}
	if !strings.EqualFold(strings.Join(got, "|"), strings.Join(want, "|")) {
		t.Fatalf("requests = %v; want %v", got, want)
	}
}
