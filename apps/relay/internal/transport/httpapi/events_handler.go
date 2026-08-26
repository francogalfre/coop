package httpapi

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/db/ent"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const (
	eventsDefaultLimit = 50
	eventsMaxLimit     = 200
	eventsMaxBeforeSeq = math.MaxInt32
)

type eventsResponse struct {
	Events    []json.RawMessage `json:"events"`
	OldestSeq int               `json:"oldest_seq"`
	HasMore   bool              `json:"has_more"`
}

func handleEvents(pool *db.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		before, ok := parseEventsBefore(r.URL.Query().Get("before"))
		if !ok {
			writeError(w, http.StatusBadRequest, "before: must be a non-negative integer")
			return
		}

		limit, ok := parseEventsLimit(r.URL.Query().Get("limit"))
		if !ok {
			writeError(w, http.StatusBadRequest, "limit: must be a positive integer")
			return
		}

		rows, err := pool.EventsBefore(r.Context(), sessionID, before, limit+1)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load events")
			return
		}

		hasMore := len(rows) > limit
		if hasMore {
			rows = rows[len(rows)-limit:]
		}

		writeJSON(w, http.StatusOK, eventsResponse{
			Events:    stampedEventData(rows),
			OldestSeq: oldestSeq(rows),
			HasMore:   hasMore,
		})
	}
}

func stampedEventData(rows []*ent.Event) []json.RawMessage {
	events := make([]json.RawMessage, 0, len(rows))
	for _, row := range rows {
		data, err := stream.StampSeq(row.Data, row.Seq)
		if err != nil {
			continue
		}
		events = append(events, data)
	}
	return events
}

func oldestSeq(rows []*ent.Event) int {
	if len(rows) == 0 {
		return 0
	}
	return rows[0].Seq
}

func parseEventsBefore(raw string) (int, bool) {
	if raw == "" {
		return eventsMaxBeforeSeq, true
	}

	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, false
	}

	return v, true
}

func parseEventsLimit(raw string) (int, bool) {
	if raw == "" {
		return eventsDefaultLimit, true
	}

	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, false
	}

	if v > eventsMaxLimit {
		v = eventsMaxLimit
	}

	return v, true
}
