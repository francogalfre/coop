package relayclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francogalfre/coop/packages/cli/internal/config"
)

func testConfig(baseURL string) config.Config {
	return config.Config{RelayURL: baseURL, SessionID: "sess-a"}
}

func TestPostEventSendsBodyToIngestEndpoint(t *testing.T) {
	var gotPath, gotMethod, gotContentType string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}))
	defer server.Close()

	body := []byte(`{"v":1,"session_id":"sess-a","seq":0,"ts":"2026-08-25T00:00:00Z","type":"session.start"}`)

	if err := PostEvent(context.Background(), testConfig(server.URL), body); err != nil {
		t.Fatalf("PostEvent() error = %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("got method %q, want POST", gotMethod)
	}
	if gotPath != "/v1/events" {
		t.Errorf("got path %q, want /v1/events", gotPath)
	}
	if gotContentType != "application/json" {
		t.Errorf("got Content-Type %q, want application/json", gotContentType)
	}
	if string(gotBody) != string(body) {
		t.Errorf("got body %q, want %q", gotBody, body)
	}
}

func TestPostEventReturnsErrorOnNon202(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "v: must be 1"})
	}))
	defer server.Close()

	err := PostEvent(context.Background(), testConfig(server.URL), []byte(`{}`))
	if err == nil {
		t.Fatal("PostEvent() error = nil, want an error on 400")
	}
}

func TestGetSteerReturnsMessageOn200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sessions/sess-a/steer" {
			t.Errorf("got path %q, want /v1/sessions/sess-a/steer", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("got method %q, want GET", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"from": "Alice", "text": "try the other branch"})
	}))
	defer server.Close()

	from, text, ok, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err != nil {
		t.Fatalf("GetSteer() error = %v", err)
	}
	if !ok {
		t.Fatal("GetSteer() ok = false, want true")
	}
	if from != "Alice" || text != "try the other branch" {
		t.Fatalf("got from=%q text=%q, want from=Alice text=%q", from, text, "try the other branch")
	}
}

func TestGetSteerReturnsNotOkOn204(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	from, text, ok, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err != nil {
		t.Fatalf("GetSteer() error = %v", err)
	}
	if ok {
		t.Fatal("GetSteer() ok = true, want false on 204")
	}
	if from != "" || text != "" {
		t.Fatalf("got from=%q text=%q, want empty strings", from, text)
	}
}

func TestGetSteerReturnsErrorOnUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, _, _, err := GetSteer(context.Background(), testConfig(server.URL), "sess-a")
	if err == nil {
		t.Fatal("GetSteer() error = nil, want an error on 500")
	}
}
