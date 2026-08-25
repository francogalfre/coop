package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func doSteerPost(t *testing.T, mailbox *stream.Mailbox, sessionID, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/"+sessionID+"/steer", strings.NewReader(body))
	req.SetPathValue("id", sessionID)
	rec := httptest.NewRecorder()

	handleSteerPost(mailbox)(rec, req)

	return rec
}

func doSteerGet(t *testing.T, mailbox *stream.Mailbox, sessionID string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/steer", nil)
	req.SetPathValue("id", sessionID)
	rec := httptest.NewRecorder()

	handleSteerGet(mailbox)(rec, req)

	return rec
}

func TestSteerPostThenGetOnceThenEmpty(t *testing.T) {
	mailbox := stream.NewMailbox()

	postRec := doSteerPost(t, mailbox, "sess-a", `{"from":"Alice","text":"try the other branch"}`)
	if postRec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", postRec.Code, postRec.Body.String())
	}

	getRec := doSteerGet(t, mailbox, "sess-a")
	if getRec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", getRec.Code, getRec.Body.String())
	}

	var payload steerMessageBody
	if err := json.NewDecoder(getRec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.From != "Alice" || payload.Text != "try the other branch" {
		t.Fatalf("unexpected payload: %+v", payload)
	}

	secondGetRec := doSteerGet(t, mailbox, "sess-a")
	if secondGetRec.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want 204 on second get: %s", secondGetRec.Code, secondGetRec.Body.String())
	}
}

func TestSteerGetUnknownSessionNoContent(t *testing.T) {
	mailbox := stream.NewMailbox()

	rec := doSteerGet(t, mailbox, "sess-ghost")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want 204", rec.Code)
	}
}

func TestHandleSteerPostValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing from", `{"text":"hello"}`},
		{"missing text", `{"from":"Alice"}`},
		{"empty from", `{"from":"","text":"hello"}`},
		{"empty text", `{"from":"Alice","text":""}`},
		{"malformed JSON", `{not json`},
	}

	mailbox := stream.NewMailbox()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doSteerPost(t, mailbox, "sess-a", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got status %d, want 400: %s", rec.Code, rec.Body.String())
			}
		})
	}
}
