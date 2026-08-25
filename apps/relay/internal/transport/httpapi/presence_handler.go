package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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

func handlePresence(registry *presence.Registry) http.HandlerFunc {
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
			results = append(results, presencePathResult{
				Path:    path,
				Signals: signalsByPath[path],
			})
		}

		writeJSON(w, http.StatusOK, presenceResponse{
			Repo:          repo,
			WindowSeconds: windowSeconds,
			Paths:         results,
		})
	}
}
