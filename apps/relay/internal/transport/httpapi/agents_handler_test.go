package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func TestHandleListProjectAgentsRejectsNonMember(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/coop/agents", nil)
	req.SetPathValue("slug", "coop")
	req = withActor(req, "user-stranger")
	rec := httptest.NewRecorder()

	handleListProjectAgents(pool)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleListProjectAgentsReportsOnlineAndOffline(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.CreateAgent(t.Context(), proj.ID, "idle-agent", "Idle Agent", "user-owner"); err != nil {
		t.Fatalf("CreateAgent(idle-agent): %v", err)
	}

	running, err := pool.CreateAgent(t.Context(), proj.ID, "running-agent", "Running Agent", "user-owner")
	if err != nil {
		t.Fatalf("CreateAgent(running-agent): %v", err)
	}

	sess, err := pool.CreateAgentSession(t.Context(), "sess-a", proj, "user-owner", "/repo", "/repo", "claude-code", time.Now())
	if err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}
	if err := pool.LinkSessionToAgent(t.Context(), sess.ID, running.ID); err != nil {
		t.Fatalf("LinkSessionToAgent: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/coop/agents", nil)
	req.SetPathValue("slug", "coop")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleListProjectAgents(pool)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var body listAgentsResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(body.Agents) != 2 {
		t.Fatalf("got %d agents, want 2: %+v", len(body.Agents), body.Agents)
	}

	byName := map[string]agentResponse{}
	for _, a := range body.Agents {
		byName[a.Name] = a
	}

	idle := byName["idle-agent"]
	if idle.Status != "offline" || idle.CurrentSessionID != nil {
		t.Fatalf("got idle-agent %+v, want status offline with no current session", idle)
	}

	live := byName["running-agent"]
	if live.Status != "online" || live.CurrentSessionID == nil || *live.CurrentSessionID != "sess-a" {
		t.Fatalf("got running-agent %+v, want status online with current session sess-a", live)
	}
}

func TestHandleMessageAgentUnknownAgentNotFound(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	if _, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner"); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/agents/ghost/message", strings.NewReader(`{"text":"hi"}`))
	req.SetPathValue("slug", "coop")
	req.SetPathValue("name", "ghost")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleMessageAgent(pool, stream.NewMailbox(), stream.New(), stream.NewSteerRequestRegistry(pool))(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleMessageAgentNonMemberNotFound(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if _, err := pool.CreateAgent(t.Context(), proj.ID, "reviewer", "Reviewer", "user-owner"); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/agents/reviewer/message", strings.NewReader(`{"text":"hi"}`))
	req.SetPathValue("slug", "coop")
	req.SetPathValue("name", "reviewer")
	req = withActor(req, "user-stranger")
	rec := httptest.NewRecorder()

	handleMessageAgent(pool, stream.NewMailbox(), stream.New(), stream.NewSteerRequestRegistry(pool))(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleMessageAgentNoLiveSessionReturnsUnavailable(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if _, err := pool.CreateAgent(t.Context(), proj.ID, "reviewer", "Reviewer", "user-owner"); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/agents/reviewer/message", strings.NewReader(`{"text":"hi"}`))
	req.SetPathValue("slug", "coop")
	req.SetPathValue("name", "reviewer")
	req = withActor(req, "user-owner")
	rec := httptest.NewRecorder()

	handleMessageAgent(pool, stream.NewMailbox(), stream.New(), stream.NewSteerRequestRegistry(pool))(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("got status %d, want 503: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleMessageAgentLiveSessionAutoModeAccepts(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	a, err := pool.CreateAgent(t.Context(), proj.ID, "reviewer", "Reviewer", "user-owner")
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	sess, err := pool.CreateAgentSession(t.Context(), "sess-a", proj, "user-owner", "/repo", "/repo", "claude-code", time.Now())
	if err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}
	if err := pool.LinkSessionToAgent(t.Context(), sess.ID, a.ID); err != nil {
		t.Fatalf("LinkSessionToAgent: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/agents/reviewer/message", strings.NewReader(`{"text":"try the other branch"}`))
	req.SetPathValue("slug", "coop")
	req.SetPathValue("name", "reviewer")
	req = withActorNamed(req, "user-owner", "Owner")
	rec := httptest.NewRecorder()

	handleMessageAgent(pool, stream.NewMailbox(), stream.New(), stream.NewSteerRequestRegistry(pool))(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	var payload steerPostResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Status != "accepted" {
		t.Fatalf("got status %q, want accepted", payload.Status)
	}
}

func TestHandleMessageAgentRestrictedModeNonOwnerPending(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	a, err := pool.CreateAgent(t.Context(), proj.ID, "reviewer", "Reviewer", "user-owner")
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	sess, err := pool.CreateAgentSession(t.Context(), "sess-a", proj, "user-owner", "/repo", "/repo", "claude-code", time.Now())
	if err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}
	if err := pool.LinkSessionToAgent(t.Context(), sess.ID, a.ID); err != nil {
		t.Fatalf("LinkSessionToAgent: %v", err)
	}
	if _, err := pool.SetSessionMode(t.Context(), sess.ID, db.SessionModeRestricted); err != nil {
		t.Fatalf("SetSessionMode: %v", err)
	}

	if err := pool.AddMember(t.Context(), proj, "user-alice", db.RoleMember); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/projects/coop/agents/reviewer/message", strings.NewReader(`{"text":"try the other branch"}`))
	req.SetPathValue("slug", "coop")
	req.SetPathValue("name", "reviewer")
	req = withActorNamed(req, "user-alice", "Alice")
	rec := httptest.NewRecorder()

	handleMessageAgent(pool, stream.NewMailbox(), stream.New(), stream.NewSteerRequestRegistry(pool))(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", rec.Code, rec.Body.String())
	}

	var payload steerPostResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Status != "pending" || payload.RequestID == "" {
		t.Fatalf("got payload %+v, want pending status with a request id", payload)
	}
}
