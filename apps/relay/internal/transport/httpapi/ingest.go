package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const coopProjectHeader = "X-Coop-Project"

type ingestEvent struct {
	V         int    `json:"v"`
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
	TS        string `json:"ts"`
	Type      string `json:"type"`

	Cwd     string `json:"cwd"`
	Repo    string `json:"repo"`
	Harness string `json:"harness"`

	Owner struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`

	Path string `json:"path"`
	Mode string `json:"mode"`
}

func parseTimestamp(ts string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t, nil
	}

	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("ts: invalid RFC3339 timestamp %q", ts)
}

func handleIngest(pool *db.Pool, registry *presence.Registry, store *stream.Store) http.HandlerFunc {
	cache := newSessionPersistCache()

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		var event ingestEvent

		if err := json.Unmarshal(body, &event); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if event.V != 1 {
			writeError(w, http.StatusBadRequest, "v: must be 1")
			return
		}

		if event.SessionID == "" || len(event.SessionID) > presence.SessionIDMax {
			writeError(w, http.StatusBadRequest, "session_id: required, max length "+fmt.Sprint(presence.SessionIDMax))
			return
		}

		if event.Seq < 0 {
			writeError(w, http.StatusBadRequest, "seq: must be >= 0")
			return
		}

		at, err := parseTimestamp(event.TS)

		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		switch event.Type {
		case "session.start":
			if event.Cwd == "" {
				writeError(w, http.StatusBadRequest, "cwd: required")
				return
			}

			if event.Owner.DisplayName == "" || len(event.Owner.DisplayName) > presence.DisplayNameMax {
				writeError(w, http.StatusBadRequest, "owner.display_name: required, max length "+fmt.Sprint(presence.DisplayNameMax))
				return
			}

			projectSlug := r.Header.Get(coopProjectHeader)
			if projectSlug == "" {
				writeError(w, http.StatusBadRequest, coopProjectHeader+": required, every session must belong to a project")
				return
			}

			registry.SessionStarted(event.SessionID, event.Cwd, event.Owner.DisplayName, event.Harness, at)

			if !persistSessionStart(w, r, pool, event, projectSlug, at) {
				return
			}

			cache.markPersisted(event.SessionID)
		case "session.end":
			registry.SessionEnded(event.SessionID, at)

			if cache.check(r.Context(), pool, event.SessionID) {
				if err := pool.EndAgentSession(r.Context(), event.SessionID, at); err != nil {
					writeError(w, http.StatusInternalServerError, "failed to end agent session")
					return
				}
			}
		case "file.touched":
			if event.Path == "" || len(event.Path) > presence.PathMax {
				writeError(w, http.StatusBadRequest, "path: required, max length "+fmt.Sprint(presence.PathMax))
				return
			}

			if event.Mode != "read" && event.Mode != "write" {
				writeError(w, http.StatusBadRequest, `mode: must be "read" or "write"`)
				return
			}

			if err := registry.FileTouched(event.SessionID, event.Path, event.Mode, at); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		if cache.check(r.Context(), pool, event.SessionID) {
			dbEvent, err := pool.AppendEvent(r.Context(), event.SessionID, body)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to record event")
				return
			}

			if _, err := store.AppendWithSeq(event.SessionID, dbEvent.Seq, json.RawMessage(body)); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to record event")
				return
			}
		} else if _, err := store.Append(event.SessionID, json.RawMessage(body)); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record event")
			return
		}

		writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
	}
}

func persistSessionStart(w http.ResponseWriter, r *http.Request, pool *db.Pool, event ingestEvent, projectSlug string, at time.Time) bool {
	actor, ok := auth.FromContext(r.Context())
	if !ok {
		http.NotFound(w, r)
		return false
	}

	proj, err := pool.GetProjectBySlug(r.Context(), projectSlug)
	if err != nil {
		if db.IsNotFound(err) {
			http.NotFound(w, r)
			return false
		}

		writeError(w, http.StatusInternalServerError, "failed to look up project")
		return false
	}

	_, isMember, err := pool.MemberRole(r.Context(), proj.ID, actor.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return false
	}

	if !isMember {
		http.NotFound(w, r)
		return false
	}

	repo := event.Repo
	if repo == "" {
		repo = event.Cwd
	}

	if _, err := pool.CreateAgentSession(r.Context(), event.SessionID, proj, actor.UserID, repo, event.Cwd, event.Harness, at); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create agent session")
		return false
	}

	return true
}
