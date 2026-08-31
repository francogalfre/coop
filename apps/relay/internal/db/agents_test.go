package db_test

import (
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func TestCreateAgentIsIdempotent(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	first, err := pool.CreateAgent(t.Context(), proj.ID, "reviewer", "Reviewer", "user-owner")
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	second, err := pool.CreateAgent(t.Context(), proj.ID, "reviewer", "Second Call", "user-someone-else")
	if err != nil {
		t.Fatalf("CreateAgent (existing): %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("got id %q, want existing agent id %q", second.ID, first.ID)
	}
	if second.DisplayName != first.DisplayName {
		t.Fatalf("got display name %q, want the existing agent's %q (create must not overwrite)", second.DisplayName, first.DisplayName)
	}
}

func TestGetAgentByNameNotFound(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.GetAgentByName(t.Context(), proj.ID, "ghost"); !db.IsNotFound(err) {
		t.Fatalf("got err %v, want a not-found error", err)
	}
}

func TestListProjectAgentsOrdersByName(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if _, err := pool.CreateAgent(t.Context(), proj.ID, "zeta", "Zeta", "user-owner"); err != nil {
		t.Fatalf("CreateAgent(zeta): %v", err)
	}
	if _, err := pool.CreateAgent(t.Context(), proj.ID, "alpha", "Alpha", "user-owner"); err != nil {
		t.Fatalf("CreateAgent(alpha): %v", err)
	}

	agents, err := pool.ListProjectAgents(t.Context(), proj.ID)
	if err != nil {
		t.Fatalf("ListProjectAgents: %v", err)
	}

	if len(agents) != 2 || agents[0].Name != "alpha" || agents[1].Name != "zeta" {
		t.Fatalf("got agents %+v, want [alpha zeta]", agents)
	}
}

func TestCurrentSessionForAgentReturnsLiveSession(t *testing.T) {
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

	got, err := pool.CurrentSessionForAgent(t.Context(), a.ID)
	if err != nil {
		t.Fatalf("CurrentSessionForAgent: %v", err)
	}
	if got == nil || got.ID != sess.ID {
		t.Fatalf("got session %+v, want %q", got, sess.ID)
	}
}

func TestCurrentSessionForAgentReturnsNilWithoutLiveSession(t *testing.T) {
	pool := dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	a, err := pool.CreateAgent(t.Context(), proj.ID, "reviewer", "Reviewer", "user-owner")
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	got, err := pool.CurrentSessionForAgent(t.Context(), a.ID)
	if err != nil {
		t.Fatalf("CurrentSessionForAgent: %v", err)
	}
	if got != nil {
		t.Fatalf("got session %+v, want nil", got)
	}
}
