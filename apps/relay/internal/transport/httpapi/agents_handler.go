package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

type agentResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	DisplayName      string  `json:"display_name"`
	Status           string  `json:"status"`
	CurrentSessionID *string `json:"current_session_id"`
}

type listAgentsResponse struct {
	Agents []agentResponse `json:"agents"`
}

func toAgentResponse(w http.ResponseWriter, r *http.Request, pool *db.Pool, a *ent.Agent) (agentResponse, bool) {
	resp := agentResponse{
		ID:          a.ID,
		Name:        a.Name,
		DisplayName: a.DisplayName,
		Status:      "offline",
	}

	sess, err := pool.CurrentSessionForAgent(r.Context(), a.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to look up agent status")
		return agentResponse{}, false
	}

	// "idle" needs runner-registration/heartbeat, which doesn't exist yet.
	if sess != nil {
		resp.Status = "online"
		id := sess.ID
		resp.CurrentSessionID = &id
	}

	return resp, true
}

func handleListProjectAgents(pool *db.Pool) http.HandlerFunc {
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

		agents, err := pool.ListProjectAgents(r.Context(), proj.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list agents")
			return
		}

		responses := make([]agentResponse, 0, len(agents))
		for _, a := range agents {
			resp, ok := toAgentResponse(w, r, pool, a)
			if !ok {
				return
			}
			responses = append(responses, resp)
		}

		writeJSON(w, http.StatusOK, listAgentsResponse{Agents: responses})
	}
}

func handleMessageAgent(pool *db.Pool, mailbox *stream.Mailbox, store *stream.Store, steerRequests *stream.SteerRequestRegistry) http.HandlerFunc {
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

		a, err := pool.GetAgentByName(r.Context(), proj.ID, r.PathValue("name"))
		if err != nil {
			if db.IsNotFound(err) {
				http.NotFound(w, r)
				return
			}

			writeError(w, http.StatusInternalServerError, "failed to look up agent")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body steerPostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Text == "" || len(body.Text) > steerTextMax {
			writeError(w, http.StatusBadRequest, "text: required, max length "+fmt.Sprint(steerTextMax))
			return
		}

		sess, err := pool.CurrentSessionForAgent(r.Context(), a.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to look up agent session")
			return
		}

		if sess == nil {
			writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("%s isn't running right now", a.Name))
			return
		}

		if sess.Mode == db.SessionModeRestricted && actor.UserID != sess.OwnerID {
			handleSteerRequestPending(w, r, pool, store, steerRequests, sess.ID, actor, body.Text)
			return
		}

		deliverSteerNow(w, r, pool, mailbox, store, sess.ID, actor, body.Text, body.ClientID)
	}
}

func projectForMember(w http.ResponseWriter, r *http.Request, pool *db.Pool, actor auth.Actor) (*ent.Project, bool) {
	proj, err := pool.GetProjectBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		http.NotFound(w, r)
		return nil, false
	}

	_, isMember, err := pool.MemberRole(r.Context(), proj.ID, actor.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check membership")
		return nil, false
	}

	if !isMember {
		http.NotFound(w, r)
		return nil, false
	}

	return proj, true
}
