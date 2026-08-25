package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/francogalfre/coop/packages/cli/internal/config"
	"github.com/francogalfre/coop/packages/cli/internal/redact"
	"github.com/francogalfre/coop/packages/cli/internal/relayclient"
)

const (
	postQueueDepth = 64
	steerTimeout   = 2 * time.Second
	claudeCodeName = "claude-code"
)

var claudeCodeSteerableEvents = map[string]bool{
	"PreToolUse":       true,
	"PostToolUse":      true,
	"UserPromptSubmit": true,
}

type Server struct {
	cfg      config.Config
	redactor *redact.Redactor
	seq      atomic.Int64
	steer    bool
	postCh   chan []byte
}

// NewServer's steer flag distinguishes `coop attach` (this server owns
// steering delivery) from `coop run` (the pty poller owns it instead) --
// GetSteer is take-once, so only one consumer may call it per session.
func NewServer(cfg config.Config, steer bool) *Server {
	s := &Server{
		cfg:      cfg,
		redactor: redact.New(),
		steer:    steer,
		postCh:   make(chan []byte, postQueueDepth),
	}

	go s.runPostLoop()

	return s
}

func (s *Server) Close() {
	close(s.postCh)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /hook/{harness}/{event}", s.handleHook)

	return mux
}

func (s *Server) handleHook(w http.ResponseWriter, r *http.Request) {
	harnessName := r.PathValue("harness")
	event := r.PathValue("event")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, map[string]any{})
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSON(w, map[string]any{})
		return
	}

	events, err := translateEvent(harnessName, s.nextSeq, s.cfg.SessionID, event, payload, s.redactor)
	if err != nil {
		log.Printf("coop: build event for %s/%s: %v", harnessName, event, err)
		writeJSON(w, map[string]any{})
		return
	}

	for _, e := range events {
		s.enqueuePost(e)
	}

	if !s.steer {
		writeJSON(w, map[string]any{})
		return
	}

	s.respondWithSteer(w, r, harnessName, event)
}

func (s *Server) nextSeq() int {
	return int(s.seq.Add(1)) - 1
}

func (s *Server) runPostLoop() {
	for body := range s.postCh {
		if err := relayclient.PostEvent(context.Background(), s.cfg, body); err != nil {
			log.Printf("coop: post event: %v", err)
		}
	}
}

// enqueuePost never blocks the hook response on the relay: this is live
// telemetry, not a ledger, so a full queue drops the oldest event.
func (s *Server) enqueuePost(body []byte) {
	select {
	case s.postCh <- body:
		return
	default:
	}

	select {
	case <-s.postCh:
		log.Printf("coop: post queue full, dropped oldest event")
	default:
	}

	select {
	case s.postCh <- body:
	default:
		log.Printf("coop: post queue full, dropped event")
	}
}

// respondWithSteer shapes the reply differently per harness: Claude Code
// only accepts additionalContext on a safe subset of events (see
// harnesses.md), while opencode/pi shims opportunistically check every
// response for a "steer" field regardless of which event triggered it.
func (s *Server) respondWithSteer(w http.ResponseWriter, r *http.Request, harnessName, event string) {
	if harnessName == claudeCodeName && !claudeCodeSteerableEvents[event] {
		writeJSON(w, map[string]any{})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), steerTimeout)
	defer cancel()

	from, text, ok, err := relayclient.GetSteer(ctx, s.cfg, s.cfg.SessionID)
	if err != nil {
		log.Printf("coop: get steer: %v", err)
		writeJSON(w, map[string]any{})
		return
	}

	if !ok {
		writeJSON(w, map[string]any{})
		return
	}

	attributed := fmt.Sprintf("[%s via coop] %s", from, text)

	if harnessName == claudeCodeName {
		writeJSON(w, map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":     event,
				"additionalContext": attributed,
			},
		})
		return
	}

	writeJSON(w, map[string]any{"steer": attributed})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
