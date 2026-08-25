package presence

import "time"

type Signal struct {
	Path      string    `json:"path"`
	SessionID string    `json:"session_id"`
	Owner     string    `json:"owner,omitempty"`
	Mode      string    `json:"mode"`
	At        time.Time `json:"at"`
	Active    bool      `json:"active"`
}

type SessionSummary struct {
	SessionID string    `json:"session_id"`
	Owner     string    `json:"owner"`
	StartedAt time.Time `json:"started_at"`
	Active    bool      `json:"active"`
}

func (r *Registry) Query(repo string, paths []string, now time.Time, window time.Duration) map[string][]Signal {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]Signal, len(paths))
	byPath := r.touches[repo]

	for _, path := range paths {
		signals := []Signal{}

		for sessionID, t := range byPath[path] {
			if now.Sub(t.At) > window {
				continue
			}

			session := r.sessions[sessionID]

			signal := Signal{
				Path:      path,
				SessionID: sessionID,
				Mode:      t.Mode,
				At:        t.At,
			}
			if session != nil {
				signal.Owner = session.Owner
				signal.Active = !session.Ended
			}

			signals = append(signals, signal)
		}

		result[path] = signals
	}

	return result
}

func (r *Registry) ActiveSessions(repo string) []SessionSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summaries := []SessionSummary{}

	for sessionID, session := range r.sessions {
		if session.Repo != repo || session.Ended {
			continue
		}

		summaries = append(summaries, SessionSummary{
			SessionID: sessionID,
			Owner:     session.Owner,
			StartedAt: session.StartedAt,
			Active:    true,
		})
	}

	return summaries
}
