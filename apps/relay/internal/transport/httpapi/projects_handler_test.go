package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func withActor(req *http.Request, userID string) *http.Request {
	return req.WithContext(auth.WithActor(req.Context(), auth.Actor{UserID: userID}))
}

func TestHandleCreateProjectPersists(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/projects", strings.NewReader(`{"name":"Coop","slug":"coop"}`))
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleCreateProject(pool)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, want 201: %s", rec.Code, rec.Body.String())
	}

	var proj projectResponse
	if err := json.NewDecoder(rec.Body).Decode(&proj); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if proj.Slug != "coop" || proj.Name != "Coop" || proj.CreatedBy != "user-owner" {
		t.Fatalf("unexpected project response: %+v", proj)
	}
}

func TestHandleCreateProjectRequiresIdentity(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/projects", strings.NewReader(`{"name":"Coop","slug":"coop"}`))
	rec := httptest.NewRecorder()

	handleCreateProject(pool)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestHandleCreateInviteRejectsNonOwner(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/invites", nil)
	req.SetPathValue("slug", "coop")
	req = withActor(req, "user-stranger")
	rec := httptest.NewRecorder()

	handleCreateInvite(pool)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleCreateInviteAllowsOwner(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/invites", nil)
	req.SetPathValue("slug", "coop")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleCreateInvite(pool)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got status %d, want 201: %s", rec.Code, rec.Body.String())
	}

	var body createInviteResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.Token == "" {
		t.Fatal("expected non-empty invite token")
	}
}

func TestHandleAcceptInviteAddsMembership(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	token, err := pool.CreateInvite(t.Context(), proj, "user-owner", inviteTTL)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/invites/"+token+"/accept", nil)
	req.SetPathValue("token", token)
	req = withActor(req, "user-newmember")
	rec := httptest.NewRecorder()

	handleAcceptInvite(pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	_, isMember, err := pool.MemberRole(t.Context(), proj.ID, "user-newmember")
	if err != nil {
		t.Fatalf("MemberRole: %v", err)
	}
	if !isMember {
		t.Fatal("expected user-newmember to be a project member after accepting invite")
	}
}

func TestHandleListProjectSessionsRejectsNonMember(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/coop/sessions", nil)
	req.SetPathValue("slug", "coop")
	req = withActor(req, "user-stranger")
	rec := httptest.NewRecorder()

	handleListProjectSessions(pool)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleListProjectSessionsAllowsMember(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/coop/sessions", nil)
	req.SetPathValue("slug", "coop")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleListProjectSessions(pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var body listSessionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(body.Sessions) != 0 {
		t.Fatalf("expected no sessions yet, got %d", len(body.Sessions))
	}

	_ = proj
}

func TestHandleListUserProjectsReturnsMemberships(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/projects", nil)
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleListUserProjects(pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var body listProjectsResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(body.Projects) != 1 || body.Projects[0].Slug != "coop" {
		t.Fatalf("unexpected projects response: %+v", body.Projects)
	}
}
