package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/francogalfre/coop/apps/relay/internal/stream"
)

func TestQuestionRoundTripsThroughAnswer(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()
	questions := stream.NewQuestionRegistry()

	postReq := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/questions", strings.NewReader(`{"text":"new table or reuse?","options":["new","reuse"]}`))
	postReq.SetPathValue("id", "sess-a")
	postReq = withActorNamed(postReq, "user-alice", "Alice")
	postRec := httptest.NewRecorder()

	handleQuestionPost(pool, store, questions)(postRec, postReq)

	if postRec.Code != http.StatusAccepted {
		t.Fatalf("got status %d, want 202: %s", postRec.Code, postRec.Body.String())
	}

	var posted questionPostResponse
	if err := json.NewDecoder(postRec.Body).Decode(&posted); err != nil {
		t.Fatalf("failed to decode post response: %v", err)
	}
	if posted.QuestionID == "" {
		t.Fatal("got empty question_id")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/sessions/sess-a/questions/"+posted.QuestionID+"?wait_seconds=1", nil)
	getReq.SetPathValue("id", "sess-a")
	getReq.SetPathValue("qid", posted.QuestionID)
	getRec := httptest.NewRecorder()

	handleQuestionGet(questions)(getRec, getReq)

	var open questionGetResponse
	if err := json.NewDecoder(getRec.Body).Decode(&open); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}
	if open.Status != "open" {
		t.Fatalf("got status %q, want open before anyone answers", open.Status)
	}

	answerReq := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/questions/"+posted.QuestionID+"/answer", strings.NewReader(`{"text":"reuse the existing table"}`))
	answerReq.SetPathValue("id", "sess-a")
	answerReq.SetPathValue("qid", posted.QuestionID)
	answerReq = withActorNamed(answerReq, "user-bob", "Bob")
	answerRec := httptest.NewRecorder()

	handleQuestionAnswerPost(pool, store, questions)(answerRec, answerReq)

	if answerRec.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200: %s", answerRec.Code, answerRec.Body.String())
	}

	getReq2 := httptest.NewRequest(http.MethodGet, "/v1/sessions/sess-a/questions/"+posted.QuestionID+"?wait_seconds=1", nil)
	getReq2.SetPathValue("id", "sess-a")
	getReq2.SetPathValue("qid", posted.QuestionID)
	getRec2 := httptest.NewRecorder()

	handleQuestionGet(questions)(getRec2, getReq2)

	var answered questionGetResponse
	if err := json.NewDecoder(getRec2.Body).Decode(&answered); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}
	if answered.Status != "answered" || answered.Answer == nil || answered.Answer.Text != "reuse the existing table" {
		t.Fatalf("got %+v, want an answered question with Bob's text", answered)
	}

	events := store.Since("sess-a", 0)
	if len(events) != 2 {
		t.Fatalf("got %d events, want agent.asked_team + human.answered", len(events))
	}
}

func TestQuestionGetWaitsUpToTheDeadlineForAnAnswer(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()
	questions := stream.NewQuestionRegistry()

	postReq := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/questions", strings.NewReader(`{"text":"pick one"}`))
	postReq.SetPathValue("id", "sess-a")
	postReq = withActorNamed(postReq, "user-alice", "Alice")
	postRec := httptest.NewRecorder()

	handleQuestionPost(pool, store, questions)(postRec, postReq)

	var posted questionPostResponse
	_ = json.NewDecoder(postRec.Body).Decode(&posted)

	answered := make(chan struct{})
	go func() {
		answerReq := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/questions/"+posted.QuestionID+"/answer", strings.NewReader(`{"text":"the second one"}`))
		answerReq.SetPathValue("id", "sess-a")
		answerReq.SetPathValue("qid", posted.QuestionID)
		answerReq = withActorNamed(answerReq, "user-bob", "Bob")
		handleQuestionAnswerPost(pool, store, questions)(httptest.NewRecorder(), answerReq)
		close(answered)
	}()

	getReq := httptest.NewRequest(http.MethodGet, "/v1/sessions/sess-a/questions/"+posted.QuestionID+"?wait_seconds=5", nil)
	getReq.SetPathValue("id", "sess-a")
	getReq.SetPathValue("qid", posted.QuestionID)
	getRec := httptest.NewRecorder()

	handleQuestionGet(questions)(getRec, getReq)
	<-answered

	var resp questionGetResponse
	if err := json.NewDecoder(getRec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "answered" {
		t.Fatalf("got status %q, want the long poll to observe the answer", resp.Status)
	}
}

func TestQuestionGetOnUnknownIDIs404(t *testing.T) {
	questions := stream.NewQuestionRegistry()

	req := httptest.NewRequest(http.MethodGet, "/v1/sessions/sess-a/questions/nope", nil)
	req.SetPathValue("id", "sess-a")
	req.SetPathValue("qid", "nope")
	rec := httptest.NewRecorder()

	handleQuestionGet(questions)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", rec.Code)
	}
}

func TestQuestionCannotBeAnsweredTwice(t *testing.T) {
	pool := messageSessionFixture(t, "sess-a")
	store := stream.New()
	questions := stream.NewQuestionRegistry()

	postReq := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/questions", strings.NewReader(`{"text":"pick one"}`))
	postReq.SetPathValue("id", "sess-a")
	postReq = withActorNamed(postReq, "user-alice", "Alice")
	postRec := httptest.NewRecorder()
	handleQuestionPost(pool, store, questions)(postRec, postReq)

	var posted questionPostResponse
	_ = json.NewDecoder(postRec.Body).Decode(&posted)

	first := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/questions/"+posted.QuestionID+"/answer", strings.NewReader(`{"text":"a"}`))
	first.SetPathValue("id", "sess-a")
	first.SetPathValue("qid", posted.QuestionID)
	first = withActorNamed(first, "user-bob", "Bob")
	handleQuestionAnswerPost(pool, store, questions)(httptest.NewRecorder(), first)

	second := httptest.NewRequest(http.MethodPost, "/v1/sessions/sess-a/questions/"+posted.QuestionID+"/answer", strings.NewReader(`{"text":"b"}`))
	second.SetPathValue("id", "sess-a")
	second.SetPathValue("qid", posted.QuestionID)
	second = withActorNamed(second, "user-carol", "Carol")
	rec := httptest.NewRecorder()
	handleQuestionAnswerPost(pool, store, questions)(rec, second)

	if rec.Code != http.StatusConflict {
		t.Fatalf("got status %d, want 409 for a second answer", rec.Code)
	}
}
