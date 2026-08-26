package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
)

const inviteTTL = 7 * 24 * time.Hour

type projectResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func toProjectResponse(proj *ent.Project) projectResponse {
	return projectResponse{
		ID:        proj.ID,
		Name:      proj.Name,
		Slug:      proj.Slug,
		CreatedBy: proj.CreatedBy,
		CreatedAt: proj.CreatedAt,
	}
}

type agentSessionResponse struct {
	ID        string     `json:"id"`
	Repo      string     `json:"repo"`
	Cwd       string     `json:"cwd"`
	Harness   string     `json:"harness"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

func toAgentSessionResponse(sess *ent.AgentSession) agentSessionResponse {
	return agentSessionResponse{
		ID:        sess.ID,
		Repo:      sess.Repo,
		Cwd:       sess.Cwd,
		Harness:   sess.Harness,
		Status:    sess.Status,
		StartedAt: sess.StartedAt,
		EndedAt:   sess.EndedAt,
	}
}

type createProjectRequestBody struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func handleCreateProject(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body createProjectRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Name == "" {
			writeError(w, http.StatusBadRequest, "name: required")
			return
		}

		if body.Slug == "" {
			writeError(w, http.StatusBadRequest, "slug: required")
			return
		}

		proj, err := pool.CreateProject(r.Context(), body.Name, body.Slug, actor.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create project")
			return
		}

		writeJSON(w, http.StatusCreated, toProjectResponse(proj))
	}
}

type listProjectsResponse struct {
	Projects []projectResponse `json:"projects"`
}

func handleListUserProjects(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		projects, err := pool.ListUserProjects(r.Context(), actor.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list projects")
			return
		}

		responses := make([]projectResponse, 0, len(projects))
		for _, proj := range projects {
			responses = append(responses, toProjectResponse(proj))
		}

		writeJSON(w, http.StatusOK, listProjectsResponse{Projects: responses})
	}
}

type createInviteResponseBody struct {
	Token string `json:"token"`
}

func handleCreateInvite(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		slug := r.PathValue("slug")

		proj, err := pool.GetProjectBySlug(r.Context(), slug)
		if err != nil {
			if db.IsNotFound(err) {
				http.NotFound(w, r)
				return
			}

			writeError(w, http.StatusInternalServerError, "failed to look up project")
			return
		}

		role, isMember, err := pool.MemberRole(r.Context(), proj.ID, actor.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check membership")
			return
		}

		if !isMember || role != db.RoleOwner {
			http.NotFound(w, r)
			return
		}

		token, err := pool.CreateInvite(r.Context(), proj, actor.UserID, inviteTTL)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create invite")
			return
		}

		writeJSON(w, http.StatusCreated, createInviteResponseBody{Token: token})
	}
}

func handleAcceptInvite(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		token := r.PathValue("token")

		proj, err := pool.AcceptInvite(r.Context(), token, actor.UserID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		writeJSON(w, http.StatusOK, toProjectResponse(proj))
	}
}

func handleListProjectSessions(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		slug := r.PathValue("slug")

		proj, err := pool.GetProjectBySlug(r.Context(), slug)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		_, isMember, err := pool.MemberRole(r.Context(), proj.ID, actor.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check membership")
			return
		}

		if !isMember {
			http.NotFound(w, r)
			return
		}

		sessions, err := pool.ListProjectSessions(r.Context(), proj.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list sessions")
			return
		}

		responses := make([]agentSessionResponse, 0, len(sessions))
		for _, sess := range sessions {
			responses = append(responses, toAgentSessionResponse(sess))
		}

		writeJSON(w, http.StatusOK, listSessionsResponse{Sessions: responses})
	}
}

type listSessionsResponse struct {
	Sessions []agentSessionResponse `json:"sessions"`
}
