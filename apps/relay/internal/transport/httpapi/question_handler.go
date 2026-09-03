package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
	"github.com/francogalfre/coop/apps/relay/internal/db"
	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

const (
	questionTextMax    = 2048
	questionOptionsMax = 6
	questionOptionMax  = 200
	questionWaitMax    = 30 * time.Second
	questionWaitMin    = 1 * time.Second
)

type questionPostRequest struct {
	Text    string   `json:"text"`
	Options []string `json:"options"`
}

type questionPostResponse struct {
	QuestionID string `json:"question_id"`
}

// handleQuestionPost is the write side of ask_team, broadcasting agent.asked_team.
func handleQuestionPost(pool *db.Pool, store *stream.Store, questions *stream.QuestionRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")

		if _, ok := auth.FromContext(r.Context()); !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body questionPostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Text == "" || len(body.Text) > questionTextMax {
			writeError(w, http.StatusBadRequest, "text: required, max length "+strconv.Itoa(questionTextMax))
			return
		}

		if len(body.Options) > questionOptionsMax {
			writeError(w, http.StatusBadRequest, "options: at most "+strconv.Itoa(questionOptionsMax))
			return
		}

		for _, opt := range body.Options {
			if opt == "" || len(opt) > questionOptionMax {
				writeError(w, http.StatusBadRequest, "options: each must be 1-"+strconv.Itoa(questionOptionMax)+" chars")
				return
			}
		}

		questionID, err := randomID()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build question")
			return
		}

		envelope, err := questionEnvelope(sessionID, questionID, body.Text, body.Options)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build question event")
			return
		}

		if _, err := publishEvent(r.Context(), pool, store, sessionID, envelope); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record question")
			return
		}

		questions.Open(questionID)

		writeJSON(w, http.StatusAccepted, questionPostResponse{QuestionID: questionID})
	}
}

type questionGetResponse struct {
	Status string          `json:"status"`
	Answer *questionAnswer `json:"answer,omitempty"`
}

type questionAnswer struct {
	Text  string          `json:"text"`
	Actor json.RawMessage `json:"actor"`
}

func handleQuestionGet(questions *stream.QuestionRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionID := r.PathValue("qid")

		wait := questionWaitMin
		if raw := r.URL.Query().Get("wait_seconds"); raw != "" {
			seconds, err := strconv.Atoi(raw)
			if err != nil || seconds < 0 {
				writeError(w, http.StatusBadRequest, "wait_seconds: must be a non-negative integer")
				return
			}
			wait = time.Duration(seconds) * time.Second
			if wait > questionWaitMax {
				wait = questionWaitMax
			}
		}

		answer, answered, existed := questions.Wait(r.Context(), questionID, wait)
		if !existed {
			http.NotFound(w, r)
			return
		}

		if !answered {
			writeJSON(w, http.StatusOK, questionGetResponse{Status: "open"})
			return
		}

		actorJSON, err := json.Marshal(map[string]string{"display_name": answer.Actor.DisplayName})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to encode answer")
			return
		}

		writeJSON(w, http.StatusOK, questionGetResponse{
			Status: "answered",
			Answer: &questionAnswer{Text: answer.Text, Actor: actorJSON},
		})
	}
}

type questionAnswerPostRequest struct {
	Text string `json:"text"`
}

type questionAnswerPostResponse struct {
	Status string `json:"status"`
}

func handleQuestionAnswerPost(pool *db.Pool, store *stream.Store, questions *stream.QuestionRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")
		questionID := r.PathValue("qid")

		actor, ok := auth.FromContext(r.Context())
		if !ok {
			http.NotFound(w, r)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		var body questionAnswerPostRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "malformed JSON body")
			return
		}

		if body.Text == "" || len(body.Text) > questionTextMax {
			writeError(w, http.StatusBadRequest, "text: required, max length "+strconv.Itoa(questionTextMax))
			return
		}

		envelope, err := answeredEnvelope(sessionID, questionID, actor, body.Text)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to build answer event")
			return
		}

		if _, err := publishEvent(r.Context(), pool, store, sessionID, envelope); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record answer")
			return
		}

		if !questions.Answer(questionID, stream.QuestionAnswer{Text: body.Text, Actor: actor}) {
			writeError(w, http.StatusConflict, "question already answered or unknown")
			return
		}

		writeJSON(w, http.StatusOK, questionAnswerPostResponse{Status: "answered"})
	}
}

func questionEnvelope(sessionID, questionID, text string, options []string) (json.RawMessage, error) {
	fields := map[string]any{
		"v":           1,
		"session_id":  sessionID,
		"ts":          time.Now().UTC().Format(time.RFC3339),
		"type":        "agent.asked_team",
		"question_id": questionID,
		"text":        map[string]any{"text": text, "redactions": 0, "truncated": false},
	}

	if len(options) > 0 {
		fields["options"] = options
	}

	return json.Marshal(fields)
}

func answeredEnvelope(sessionID, questionID string, actor auth.Actor, text string) (json.RawMessage, error) {
	fields := map[string]any{
		"v":           1,
		"session_id":  sessionID,
		"ts":          time.Now().UTC().Format(time.RFC3339),
		"type":        "human.answered",
		"question_id": questionID,
		"actor":       actorJSON(actor),
		"text":        map[string]any{"text": text, "redactions": 0, "truncated": false},
	}

	return json.Marshal(fields)
}
