package relayclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProjectContextReturnsTextAndVersionOn200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/acme/context" {
			t.Errorf("got path %q, want /v1/projects/acme/context", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("got method %q, want GET", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"text":       "read the runbook first",
			"version":    3,
			"updated_by": "Mara",
			"updated_at": "2026-09-03T10:00:00Z",
		})
	}))
	defer server.Close()

	result, err := GetProjectContext(context.Background(), testConfig(server.URL), "acme")
	if err != nil {
		t.Fatalf("GetProjectContext() error = %v", err)
	}
	if result.Text != "read the runbook first" || result.Version != 3 {
		t.Fatalf("got %+v, want text=%q version=3", result, "read the runbook first")
	}
}

func TestGetProjectContextReturnsEmptyOnNeverSetProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"text": "", "version": 0})
	}))
	defer server.Close()

	result, err := GetProjectContext(context.Background(), testConfig(server.URL), "acme")
	if err != nil {
		t.Fatalf("GetProjectContext() error = %v", err)
	}
	if result.Text != "" || result.Version != 0 {
		t.Fatalf("got %+v, want empty text and version 0", result)
	}
}

func TestGetProjectContextReturnsErrorOnUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := GetProjectContext(context.Background(), testConfig(server.URL), "acme")
	if err == nil {
		t.Fatal("GetProjectContext() error = nil, want an error on 404")
	}
}

func TestGetProjectContextSendsCLICredential(t *testing.T) {
	var gotAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"text": "", "version": 0})
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	cfg.CLICredential = "deadbeef"

	if _, err := GetProjectContext(context.Background(), cfg, "acme"); err != nil {
		t.Fatalf("GetProjectContext() error = %v", err)
	}

	if gotAuth != "Bearer deadbeef" {
		t.Fatalf("got Authorization %q, want %q", gotAuth, "Bearer deadbeef")
	}
}
