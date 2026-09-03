package stream

import (
	"testing"
	"time"

	"github.com/francogalfre/coop/apps/relay/internal/auth"
)

func TestQuestionRegistryWaitReturnsOpenBeforeAnAnswer(t *testing.T) {
	r := NewQuestionRegistry()
	r.Open("q-1")

	_, answered, existed := r.Wait(t.Context(), "q-1", 10*time.Millisecond)
	if !existed {
		t.Fatal("got existed=false, want the question to exist")
	}
	if answered {
		t.Fatal("got answered=true, want no answer before the wait times out")
	}
}

func TestQuestionRegistryWaitWakesOnAnswer(t *testing.T) {
	r := NewQuestionRegistry()
	r.Open("q-1")

	done := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		if !r.Answer("q-1", QuestionAnswer{Text: "reuse it", Actor: auth.Actor{DisplayName: "Bob"}}) {
			t.Error("expected Answer to succeed on an open question")
		}
		close(done)
	}()

	answer, answered, existed := r.Wait(t.Context(), "q-1", 2*time.Second)
	<-done

	if !existed || !answered {
		t.Fatalf("got existed=%v answered=%v, want both true", existed, answered)
	}
	if answer.Text != "reuse it" || answer.Actor.DisplayName != "Bob" {
		t.Fatalf("got %+v, want Bob's answer", answer)
	}
}

func TestQuestionRegistryAnswerIsOneShot(t *testing.T) {
	r := NewQuestionRegistry()
	r.Open("q-1")

	if !r.Answer("q-1", QuestionAnswer{Text: "first"}) {
		t.Fatal("expected the first answer to succeed")
	}
	if r.Answer("q-1", QuestionAnswer{Text: "second"}) {
		t.Fatal("expected a second answer to be rejected")
	}
}

func TestQuestionRegistryAnswerOnUnknownQuestionFails(t *testing.T) {
	r := NewQuestionRegistry()

	if r.Answer("nope", QuestionAnswer{Text: "x"}) {
		t.Fatal("expected Answer to fail for an unknown question id")
	}
}

func TestQuestionRegistryWaitOnUnknownQuestionReportsNotExisted(t *testing.T) {
	r := NewQuestionRegistry()

	_, _, existed := r.Wait(t.Context(), "nope", 10*time.Millisecond)
	if existed {
		t.Fatal("got existed=true for a question that was never opened")
	}
}
