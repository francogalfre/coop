package capabilities

import "testing"

func TestForAttachClaudeCode(t *testing.T) {
	got := ForAttach("claude-code")
	want := Capabilities{Steer: true, Block: true, Commands: false, PTY: false}

	if got != want {
		t.Fatalf("ForAttach(claude-code) = %+v, want %+v", got, want)
	}
}

func TestForAttachOpencodeAndPi(t *testing.T) {
	for _, harnessName := range []string{"opencode", "pi"} {
		t.Run(harnessName, func(t *testing.T) {
			got := ForAttach(harnessName)
			want := Capabilities{Steer: true, Block: false, Commands: false, PTY: false}

			if got != want {
				t.Fatalf("ForAttach(%s) = %+v, want %+v", harnessName, got, want)
			}
		})
	}
}

func TestForRunIsUniformAcrossHarnesses(t *testing.T) {
	got := ForRun()
	want := Capabilities{Steer: true, Block: false, Commands: true, PTY: true}

	if got != want {
		t.Fatalf("ForRun() = %+v, want %+v", got, want)
	}
}
