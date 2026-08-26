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
	cfg              config.Config
	redactor         *redact.Redactor
	seq              atomic.Int64
	steer            bool
	postCh           chan []byte
	takeoverNotified atomic.Bool
}

// steer distinguishes attach (this server owns delivery) from run (the pty poller does) since GetSteer is take-once.
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

	events, err := translateEvent(harnessName, s.cfg, s.nextSeq, s.cfg.SessionID, event, payload, s.redactor)
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

// This is live telemetry, not a ledger, so a full queue drops the oldest event rather than blocking the hook response.
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

// Claude Code only accepts additionalContext on a safe subset of events (harnesses.md); opencode/pi accept a "steer" field on any event.
func (s *Server) respondWithSteer(w http.ResponseWriter, r *http.Request, harnessName, event string) {
	if harnessName == claudeCodeName && !claudeCodeSteerableEvents[event] {
		writeJSON(w, map[string]any{})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), steerTimeout)
	defer cancel()

	steer, err := relayclient.GetSteer(ctx, s.cfg, s.cfg.SessionID)
	if err != nil {
		log.Printf("coop: get steer: %v", err)
		writeJSON(w, map[string]any{})
		return
	}

	// Live-verified (harnesses.md, 2026-08-26): only Claude Code's PreToolUse deny shape actually blocks the call.
	if steer.Takeover.Active && harnessName == claudeCodeName && event == "PreToolUse" {
		writeJSON(w, map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":            event,
				"permissionDecision":       "deny",
				"permissionDecisionReason": steer.Takeover.By + " has taken over this session via coop",
			},
		})
		return
	}

	text := s.steerText(steer)
	if text == "" {
		writeJSON(w, map[string]any{})
		return
	}

	if harnessName == claudeCodeName {
		writeJSON(w, map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":     event,
				"additionalContext": text,
			},
		})
		return
	}

	writeJSON(w, map[string]any{"steer": text})
}

// steerText edge-triggers the takeover notice (once per claim, not once per poll) and merges it with any pending mailbox message.
func (s *Server) steerText(steer relayclient.SteerResult) string {
	notice := ""
	if steer.Takeover.Active {
		if !s.takeoverNotified.Swap(true) {
			notice = steer.Takeover.By + " has taken over this session via coop"
		}
	} else {
		s.takeoverNotified.Store(false)
	}

	message := ""
	if steer.HasMessage {
		message = fmt.Sprintf("[%s via coop] %s", steer.From, steer.Text)
	}

	switch {
	case notice != "" && message != "":
		return message + "\n" + notice
	case notice != "":
		return notice
	default:
		return message
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
