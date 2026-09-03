package relayclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSteerReturnsMessageOn200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sessions/sess-a/steer" {
			t.Errorf("got path %q, want /v1/sessions/sess-a/steer", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("got method %q, want GET", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"has_message": true,
			"from":        "Alice",
			"text":        "try the other branch",
			"takeover":    map[string]any{"active": false},
		})
	}))
	defer server.Close()

	result, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err != nil {
		t.Fatalf("GetSteer() error = %v", err)
	}
	if !result.HasMessage {
		t.Fatal("GetSteer() HasMessage = false, want true")
	}
	if result.From != "Alice" || result.Text != "try the other branch" {
		t.Fatalf("got from=%q text=%q, want from=Alice text=%q", result.From, result.Text, "try the other branch")
	}
}

func TestGetSteerReturnsNoMessageWhenMailboxEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"has_message": false, "takeover": map[string]any{"active": false}})
	}))
	defer server.Close()

	result, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err != nil {
		t.Fatalf("GetSteer() error = %v", err)
	}
	if result.HasMessage {
		t.Fatal("GetSteer() HasMessage = true, want false on an empty mailbox")
	}
}

func TestGetSteerDecodesIDAndKindWhenPresent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"has_message": true,
			"from":        "Alice",
			"text":        "/model sonnet",
			"id":          "steer-1",
			"kind":        "command",
			"takeover":    map[string]any{"active": false},
		})
	}))
	defer server.Close()

	result, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err != nil {
		t.Fatalf("GetSteer() error = %v", err)
	}
	if result.ID != "steer-1" || result.Kind != "command" {
		t.Fatalf("got id=%q kind=%q, want id=steer-1 kind=command", result.ID, result.Kind)
	}
}

func TestGetSteerDegradesQuietlyWhenIDAndKindAbsent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"has_message": true,
			"from":        "Alice",
			"text":        "try the other branch",
			"takeover":    map[string]any{"active": false},
		})
	}))
	defer server.Close()

	result, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err != nil {
		t.Fatalf("GetSteer() error = %v", err)
	}
	if result.ID != "" || result.Kind != "" {
		t.Fatalf("got id=%q kind=%q, want both empty (relay hasn't shipped them yet)", result.ID, result.Kind)
	}
}

func TestGetSteerReturnsTakeoverState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"has_message": false,
			"takeover":    map[string]any{"active": true, "by": "Alice"},
		})
	}))
	defer server.Close()

	result, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err != nil {
		t.Fatalf("GetSteer() error = %v", err)
	}
	if !result.Takeover.Active || result.Takeover.By != "Alice" {
		t.Fatalf("got takeover %+v, want active held by Alice", result.Takeover)
	}
}

func TestGetSteerReturnsErrorOnUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err == nil {
		t.Fatal("GetSteer() error = nil, want an error on 500")
	}
}
