package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
)

const projectContextTextMax = 16_384

type projectContextResponse struct {
	Text      string     `json:"text"`
	Version   int        `json:"version"`
	UpdatedBy string     `json:"updated_by,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func handleGetProjectContext(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		proj, ok := projectForMember(w, r, pool, actor)
		if !ok {
			return
		}

		writeJSON(w, http.StatusOK, projectContextResponseFrom(proj))
	}
}

type projectContextPutRequest struct {
	Text string `json:"text"`
}

func handlePutProjectContext(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		proj, ok := projectForMember(w, r, pool, actor)
		if !ok {
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body projectContextPutRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if len(body.Text) > projectContextTextMax {
			writeError(w, http.StatusBadRequest, "text: too long")
			return
		}

		updated, err := pool.SetProjectContext(r.Context(), proj.ID, body.Text, actor.DisplayName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save project context")
			return
		}

		writeJSON(w, http.StatusOK, projectContextResponseFrom(updated))
	}
}

func projectContextResponseFrom(proj *ent.Project) projectContextResponse {
	return projectContextResponse{
		Text:      proj.ContextText,
		Version:   proj.ContextVersion,
		UpdatedBy: proj.ContextUpdatedBy,
		UpdatedAt: proj.ContextUpdatedAt,
	}
}
