package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestProjectNotePostThenGetRoundTrips(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/notes", strings.NewReader(`{"text":"the rate limit is per-IP not per-user"}`))
	postReq.SetPathValue("slug", "coop")
	postReq = withActorNamed(postReq, "user-owner", "Alice")
	postRec := httptest.NewRecorder()

	handlePostProjectNote(pool)(postRec, postReq)

	if postRec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", postRec.Code, postRec.Body.String())
	}

	var posted projectNotePostResponse
	if err := json.NewDecoder(postRec.Body).Decode(&posted); err != nil {
		t.Fatalf("failed to decode post response: %v", err)
	}
	if posted.Note.ID == "" || posted.Note.Source != "human" || posted.Note.AuthorDisplayName != "Alice" {
		t.Fatalf("got %+v, want a human note attributed to Alice with an id", posted.Note)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/projects/coop/notes", nil)
	getReq.SetPathValue("slug", "coop")
	getReq = withActorNamed(getReq, "user-owner", "Alice")
	getRec := httptest.NewRecorder()

	handleGetProjectNotes(pool)(getRec, getReq)

	var got projectNotesGetResponse
	if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}
	if len(got.Notes) != 1 || got.Notes[0].ID != posted.Note.ID {
		t.Fatalf("got %+v, want the note just posted", got.Notes)
	}
}

func TestProjectNotePostFromAnAgentIsTagged(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/notes", strings.NewReader(`{"text":"found it","source":"agent","session_id":"sess-a"}`))
	req.SetPathValue("slug", "coop")
	req = withActorNamed(req, "user-owner", "Alice")
	rec := httptest.NewRecorder()

	handlePostProjectNote(pool)(rec, req)

	var posted projectNotePostResponse
	if err := json.NewDecoder(rec.Body).Decode(&posted); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if posted.Note.Source != "agent" || posted.Note.SessionID != "sess-a" {
		t.Fatalf("got %+v, want source=agent session_id=sess-a", posted.Note)
	}
}

func TestProjectNotesReturnNewestFirst(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	for _, text := range []string{"first", "second", "third"} {
		req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/notes", strings.NewReader(`{"text":"`+text+`"}`))
		req.SetPathValue("slug", "coop")
		req = withActorNamed(req, "user-owner", "Alice")
		handlePostProjectNote(pool)(httptest.NewRecorder(), req)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/coop/notes", nil)
	req.SetPathValue("slug", "coop")
	req = withActorNamed(req, "user-owner", "Alice")
	rec := httptest.NewRecorder()

	handleGetProjectNotes(pool)(rec, req)

	var got projectNotesGetResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got.Notes) != 3 || got.Notes[0].Text != "third" {
		t.Fatalf("got %+v, want 3 notes newest-first starting with \"third\"", got.Notes)
	}
}

func TestProjectNotePostRejectsNonMember(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/notes", strings.NewReader(`{"text":"hi"}`))
	req.SetPathValue("slug", "coop")
	req = withActor(req, "user-stranger")
	rec := httptest.NewRecorder()

	handlePostProjectNote(pool)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}
