package auth_test

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/config"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/dbtest"
)

func bearerFor(t *testing.T, pool *db.Pool, userID string) string {
	t.Helper()

	rawToken, err := pool.CreateCliCredential(t.Context(), userID, "Display "+userID)
	if err != nil {
		t.Fatalf("CreateCliCredential: %v", err)
	}

	return "Bearer " + hex.EncodeToString(rawToken)
}

func sessionMemberFixture(t *testing.T) (pool *db.Pool, sessionID string) {
	t.Helper()

	pool = dbtest.OpenScratch(t)

	proj, err := pool.CreateProject(t.Context(), "Coop", "coop", "user-owner")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if err := pool.AddMember(t.Context(), proj, "user-member", db.RoleMember); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	sess, err := pool.CreateAgentSession(t.Context(), "sess-a", proj, "user-owner", "/repo", "/repo", "claude-code", time.Now())
	if err != nil {
		t.Fatalf("CreateAgentSession: %v", err)
	}

	return pool, sess.ID
}

func requestForSession(sessionID, authHeader string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/"+sessionID+"/stream", nil)
	req.SetPathValue("id", sessionID)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	return req
}

func TestRequireSessionMemberAllowsOwnerAndMember(t *testing.T) {
	pool, sessionID := sessionMemberFixture(t)
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionMember(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, userID := range []string{"user-owner", "user-member"} {
		rec := httptest.NewRecorder()
		handler(rec, requestForSession(sessionID, bearerFor(t, pool, userID)))

		if rec.Code != http.StatusOK {
			t.Fatalf("actor %s: got status %d, want 200: %s", userID, rec.Code, rec.Body.String())
		}
	}
}

func TestRequireSessionMemberRejectsNonMember(t *testing.T) {
	pool, sessionID := sessionMemberFixture(t)
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionMember(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	rec := httptest.NewRecorder()
	handler(rec, requestForSession(sessionID, bearerFor(t, pool, "user-stranger")))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireSessionMemberRejectsUnauthenticated(t *testing.T) {
	pool, sessionID := sessionMemberFixture(t)
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionMember(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	rec := httptest.NewRecorder()
	handler(rec, requestForSession(sessionID, ""))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireSessionMemberRejectsUnknownSession(t *testing.T) {
	pool, _ := sessionMemberFixture(t)
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionMember(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	rec := httptest.NewRecorder()
	handler(rec, requestForSession("sess-ghost", bearerFor(t, pool, "user-owner")))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireSessionOwnerAllowsOnlyOwner(t *testing.T) {
	pool, sessionID := sessionMemberFixture(t)
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionOwner(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	handler(rec, requestForSession(sessionID, bearerFor(t, pool, "user-owner")))

	if rec.Code != http.StatusOK {
		t.Fatalf("owner: got status %d, want 200: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireSessionOwnerRejectsNonOwnerMember(t *testing.T) {
	pool, sessionID := sessionMemberFixture(t)
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionOwner(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	rec := httptest.NewRecorder()
	handler(rec, requestForSession(sessionID, bearerFor(t, pool, "user-member")))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("member-but-not-owner: got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func TestRequireSessionOwnerRejectsNonMember(t *testing.T) {
	pool, sessionID := sessionMemberFixture(t)
	cfg := config.Config{WebInternalURL: "http://unused.invalid"}

	handler := auth.RequireSessionOwner(pool, cfg)(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	rec := httptest.NewRecorder()
	handler(rec, requestForSession(sessionID, bearerFor(t, pool, "user-stranger")))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404: %s", rec.Code, rec.Body.String())
	}
}
