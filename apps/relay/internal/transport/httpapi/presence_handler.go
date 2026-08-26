package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/presence"
)

const defaultWindowSeconds = 900

type presencePathResult struct {
	Path    string            `json:"path"`
	Signals []presence.Signal `json:"signals"`
}

type presenceResponse struct {
	Repo          string               `json:"repo"`
	WindowSeconds int                  `json:"window_seconds"`
	Paths         []presencePathResult `json:"paths"`
}

func handlePresence(pool *db.Pool, registry *presence.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		if repo == "" {
			writeError(w, http.StatusBadRequest, "repo: required")
			return
		}

		rawPaths := r.URL.Query().Get("paths")
		if rawPaths == "" {
			writeError(w, http.StatusBadRequest, "paths: required")
			return
		}

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		allowed, err := memberSessionIDs(r.Context(), pool, actor.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list projects")
			return
		}

		paths := strings.Split(rawPaths, ",")

		windowSeconds := defaultWindowSeconds
		if raw := r.URL.Query().Get("window_seconds"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed <= 0 {
				writeError(w, http.StatusBadRequest, "window_seconds: must be a positive integer")
				return
			}

			windowSeconds = parsed
		}

		window := time.Duration(windowSeconds) * time.Second
		signalsByPath := registry.Query(repo, paths, time.Now(), window)

		results := make([]presencePathResult, 0, len(paths))
		for _, path := range paths {
			signals := make([]presence.Signal, 0, len(signalsByPath[path]))
			for _, sig := range signalsByPath[path] {
				if allowed[sig.SessionID] {
					signals = append(signals, sig)
				}
			}

			results = append(results, presencePathResult{
				Path:    path,
				Signals: signals,
			})
		}

		writeJSON(w, http.StatusOK, presenceResponse{
			Repo:          repo,
			WindowSeconds: windowSeconds,
			Paths:         results,
		})
	}
}
