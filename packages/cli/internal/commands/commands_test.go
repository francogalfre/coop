package commands

import (
	"reflect"
	"testing"
)

func TestAllowedMatchesProtocolAllowlist(t *testing.T) {
	want := []string{"model", "compact", "clear", "context", "cost", "status"}

	if !reflect.DeepEqual(Allowed, want) {
		t.Fatalf("got %v, want %v (packages/protocol/src/shared/commands.ts HARNESS_COMMANDS drifted)", Allowed, want)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"/model sonnet", true},
		{"/compact", true},
		{"/clear", true},
		{"/context", true},
		{"/cost", true},
		{"/status", true},
		{"/rm -rf /", false},
		{"model sonnet", false},
		{"", false},
		{"/", false},
		{"/model; rm -rf /", false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := Validate(tt.text); got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}
