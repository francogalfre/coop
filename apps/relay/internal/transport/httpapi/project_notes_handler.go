package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
)

const projectNoteTextMax = 2_000

type projectNoteResponse struct {
	ID                string    `json:"id"`
	AuthorID          string    `json:"author_id"`
	AuthorDisplayName string    `json:"author_display_name"`
	AuthorAvatarURL   string    `json:"author_avatar_url,omitempty"`
	Source            string    `json:"source"`
	SessionID         string    `json:"session_id,omitempty"`
	Text              string    `json:"text"`
	CreatedAt         time.Time `json:"created_at"`
}

func projectNoteResponseFrom(n *ent.ProjectNote) projectNoteResponse {
	return projectNoteResponse{
		ID:                n.ID,
		AuthorID:          n.AuthorID,
		AuthorDisplayName: n.AuthorDisplayName,
		AuthorAvatarURL:   n.AuthorAvatarURL,
		Source:            n.Source,
		SessionID:         n.SessionID,
		Text:              n.Text,
		CreatedAt:         n.CreatedAt,
	}
}

type projectNotePostRequest struct {
	Text      string `json:"text"`
	SessionID string `json:"session_id"`
	Source    string `json:"source"`
}

type projectNotePostResponse struct {
	Note projectNoteResponse `json:"note"`
}

func handlePostProjectNote(pool *db.Pool) http.HandlerFunc {
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

		var body projectNotePostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Text == "" || len(body.Text) > projectNoteTextMax {
			writeError(w, http.StatusBadRequest, "text: required, max length "+strconv.Itoa(projectNoteTextMax))
			return
		}

		source := body.Source
		if source != "agent" {
			source = "human"
		}

		note, err := pool.CreateProjectNote(r.Context(), proj.ID, actor.UserID, actor.DisplayName, actor.AvatarURL, source, body.SessionID, body.Text)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save note")
			return
		}

		writeJSON(w, http.StatusAccepted, projectNotePostResponse{Note: projectNoteResponseFrom(note)})
	}
}

type projectNotesGetResponse struct {
	Notes []projectNoteResponse `json:"notes"`
}

func handleGetProjectNotes(pool *db.Pool) http.HandlerFunc {
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

		limit := 0
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 0 {
				writeError(w, http.StatusBadRequest, "limit: must be a non-negative integer")
				return
			}
			limit = parsed
		}

		notes, err := pool.ListProjectNotes(r.Context(), proj.ID, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list notes")
			return
		}

		resp := projectNotesGetResponse{Notes: make([]projectNoteResponse, len(notes))}
		for i, n := range notes {
			resp.Notes[i] = projectNoteResponseFrom(n)
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
