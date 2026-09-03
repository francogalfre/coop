package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestProjectContextRoundTripsThroughPutThenGet(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	putReq := httptest.NewRequest(http.MethodPut, "/v1/projects/coop/context", strings.NewReader(`{"text":"This repo powers billing."}`))
	putReq.SetPathValue("slug", "coop")
	putReq = withActorNamed(putReq, "user-owner", "Alice")
	putRec := httptest.NewRecorder()

	handlePutProjectContext(pool)(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", putRec.Code, putRec.Body.String())
	}

	var put projectContextResponse
	if err := json.NewDecoder(putRec.Body).Decode(&put); err != nil {
		t.Fatalf("failed to decode put response: %v", err)
	}
	if put.Text != "This repo powers billing." || put.Version != 1 || put.UpdatedBy != "Alice" {
		t.Fatalf("got %+v, want version 1 attributed to Alice", put)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/projects/coop/context", nil)
	getReq.SetPathValue("slug", "coop")
	getReq = withActorNamed(getReq, "user-owner", "Alice")
	getRec := httptest.NewRecorder()

	handleGetProjectContext(pool)(getRec, getReq)

	var got projectContextResponse
	if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}
	if got.Text != put.Text || got.Version != put.Version || got.UpdatedBy != put.UpdatedBy {
		t.Fatalf("got %+v from GET, want it to match the PUT response %+v", got, put)
	}
}

func TestProjectContextVersionIncrementsOnEachEdit(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	for i, text := range []string{"first", "second", "third"} {
		req := httptest.NewRequest(http.MethodPut, "/v1/projects/coop/context", strings.NewReader(`{"text":"`+text+`"}`))
		req.SetPathValue("slug", "coop")
		req = withActorNamed(req, "user-owner", "Alice")
		rec := httptest.NewRecorder()

		handlePutProjectContext(pool)(rec, req)

		var resp projectContextResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response %d: %v", i, err)
		}
		if resp.Version != i+1 {
			t.Fatalf("got version %d on edit %d, want %d", resp.Version, i, i+1)
		}
	}
}

func TestProjectContextRejectsNonMember(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/coop/context", nil)
	req.SetPathValue("slug", "coop")
	req = withActor(req, "user-stranger")
	rec := httptest.NewRecorder()

	handleGetProjectContext(pool)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestProjectContextAnyMemberCanEdit(t *testing.T) {
	pool := dbtest.OpenScratch(t)
	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := pool.AddMember(t.Context(), proj, "user-mara", "member"); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/v1/projects/coop/context", strings.NewReader(`{"text":"updated by a member"}`))
	req.SetPathValue("slug", "coop")
	req = withActorNamed(req, "user-mara", "Mara")
	rec := httptest.NewRecorder()

	handlePutProjectContext(pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200 (any member can edit): %s", rec.Code, rec.Body.String())
	}
}
