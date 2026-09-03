package relayclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeliverSteerSendsTextAndContextVersion(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}))
	defer server.Close()

	version := 3
	err := DeliverSteer(context.Background(), testConfig(server.URL), "sess-a", "read the runbook first", &version)
	if err != nil {
		t.Fatalf("DeliverSteer() error = %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("got method %q, want POST", gotMethod)
	}
	if gotPath != "/v1/sessions/sess-a/steer" {
		t.Errorf("got path %q, want /v1/sessions/sess-a/steer", gotPath)
	}
	if gotBody["text"] != "read the runbook first" {
		t.Errorf("got text %v, want %q", gotBody["text"], "read the runbook first")
	}
	if gotBody["project_context_version"] != float64(3) {
		t.Errorf("got project_context_version %v, want 3", gotBody["project_context_version"])
	}
}

func TestDeliverSteerOmitsContextVersionWhenNil(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}))
	defer server.Close()

	if err := DeliverSteer(context.Background(), testConfig(server.URL), "sess-a", "hi", nil); err != nil {
		t.Fatalf("DeliverSteer() error = %v", err)
	}

	if _, present := gotBody["project_context_version"]; present {
		t.Fatalf("got project_context_version present in body, want omitted")
	}
}

func TestDeliverSteerReturnsErrorOnUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	err := DeliverSteer(context.Background(), testConfig(server.URL), "sess-a", "hi", nil)
	if err == nil {
		t.Fatal("DeliverSteer() error = nil, want an error on 500")
	}
}
