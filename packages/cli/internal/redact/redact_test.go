package redact

import (
	"strings"
	"testing"
)

func TestTextRedactsKnownSecretShapes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMasked  []string
		wantVisible []string
	}{
		{
			name:       "anthropic api key",
			input:      "export ANTHROPIC_API_KEY=sk-ant-api03-abcdefghijklmnopqrstuvwxyz0123456789",
			wantMasked: []string{"sk-ant-api03-abcdefghijklmnopqrstuvwxyz0123456789"},
		},
		{
			name:       "openai-style api key",
			input:      "key is sk-abcdefghijklmnopqrstuvwx",
			wantMasked: []string{"sk-abcdefghijklmnopqrstuvwx"},
		},
		{
			name:       "github personal access token",
			input:      "curl -H \"Authorization: token ghp_1234567890abcdefghijklmnopqrstuvwxyz\"",
			wantMasked: []string{"ghp_1234567890abcdefghijklmnopqrstuvwxyz"},
		},
		{
			name:       "aws access key id",
			input:      "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			wantMasked: []string{"AKIAIOSFODNN7EXAMPLE"},
		},
		{
			name:       "slack token",
			input:      "webhook token xoxb-1234567890-abcdefghij",
			wantMasked: []string{"xoxb-1234567890-abcdefghij"},
		},
		{
			name:       "google api key",
			input:      "AIzaSyD-abcdefghijklmnopqrstuvwxyz012345",
			wantMasked: []string{"AIzaSyD-abcdefghijklmnopqrstuvwxyz012345"},
		},
		{
			name:       "gitlab personal access token",
			input:      "glpat-abcdefghijklmnopqrstu",
			wantMasked: []string{"glpat-abcdefghijklmnopqrstu"},
		},
		{
			name:       "bearer token",
			input:      "curl -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.payload'",
			wantMasked: []string{"Bearer"},
		},
		{
			name:       "authorization header basic auth",
			input:      `headers = {"Authorization": "Basic dXNlcjpwYXNzd29yZA=="}`,
			wantMasked: []string{"Basic dXNlcjpwYXNzd29yZA=="},
		},
		{
			name:       "cookie header",
			input:      `Set-Cookie: session=abc123def456; Path=/; HttpOnly`,
			wantMasked: []string{"session=abc123def456"},
		},
		{
			name:       "postgres connection string",
			input:      "DATABASE_URL=postgres://appuser:s3cr3tpassword@db.internal:5432/prod",
			wantMasked: []string{"appuser:s3cr3tpassword@db.internal:5432/prod"},
		},
		{
			name:       "mongodb connection string",
			input:      "mongodb+srv://admin:hunter2pass@cluster0.mongodb.net/mydb",
			wantMasked: []string{"admin:hunter2pass@cluster0.mongodb.net/mydb"},
		},
		{
			name:  "private key pem block",
			input: "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA1abcXYZ\nmoredata==\n-----END RSA PRIVATE KEY-----",
			wantMasked: []string{
				"-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA1abcXYZ\nmoredata==\n-----END RSA PRIVATE KEY-----",
			},
		},
		{
			name:       "assigned password",
			input:      `{"password": "hunter2secret"}`,
			wantMasked: []string{"hunter2secret"},
		},
		{
			name:       "assigned generic token",
			input:      "token: abcd1234efgh5678",
			wantMasked: []string{"abcd1234efgh5678"},
		},
	}

	r := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, count := r.Text(tt.input)

			if count == 0 {
				t.Fatalf("Text(%q) redacted nothing, want at least one match", tt.input)
			}

			for _, secret := range tt.wantMasked {
				if strings.Contains(out, secret) {
					t.Errorf("Text(%q) = %q, still contains secret %q", tt.input, out, secret)
				}
			}

			if !strings.Contains(out, mask) {
				t.Errorf("Text(%q) = %q, want it to contain %q", tt.input, out, mask)
			}
		})
	}
}

func TestTextDoesNotOverRedactNormalText(t *testing.T) {
	tests := []string{
		"the quick brown fox jumps over the lazy dog",
		"please read the README and run npm install",
		"the api_key parameter is optional for this endpoint",
		"status: ok",
		"https://example.com/docs/getting-started",
		"git clone https://github.com/example/repo.git",
		"the password must be at least 8 characters",
		"session started at 10:30am",
		"token is required to authenticate",
		"the tool call returned exit code 0",
	}

	r := New()

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			out, count := r.Text(input)
			if count != 0 {
				t.Errorf("Text(%q) = %q, redacted %d matches, want 0", input, out, count)
			}
		})
	}
}

func TestTextCountsEachMatch(t *testing.T) {
	input := "key1=sk-abcdefghijklmnopqrstuv key2=sk-zyxwvutsrqponmlkjihg"

	r := New()

	out, count := r.Text(input)
	if count != 2 {
		t.Fatalf("got %d redactions, want 2: %q", count, out)
	}
}

func TestValueRedactsNestedObjectsAndArrays(t *testing.T) {
	r := New()

	input := map[string]any{
		"prompt": "use this key: sk-abcdefghijklmnopqrstuv",
		"nested": map[string]any{
			"env": []any{
				"NORMAL_VAR=hello",
				"AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			},
		},
		"count": float64(3),
		"ok":    true,
		"empty": nil,
	}

	out, count := r.Value(input)

	if count != 2 {
		t.Fatalf("got %d redactions, want 2", count)
	}

	outMap, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("Value returned %T, want map[string]any", out)
	}

	if strings.Contains(outMap["prompt"].(string), "sk-abcdefghijklmnopqrstuv") {
		t.Errorf("prompt still contains the api key: %v", outMap["prompt"])
	}

	nested := outMap["nested"].(map[string]any)
	env := nested["env"].([]any)

	if env[0] != "NORMAL_VAR=hello" {
		t.Errorf("normal env var was altered: %v", env[0])
	}

	if strings.Contains(env[1].(string), "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("aws key still present: %v", env[1])
	}

	if outMap["count"] != float64(3) || outMap["ok"] != true || outMap["empty"] != nil {
		t.Errorf("non-string values were altered: count=%v ok=%v empty=%v", outMap["count"], outMap["ok"], outMap["empty"])
	}
}

func TestValueLeavesNonSecretValuesUntouched(t *testing.T) {
	r := New()

	input := map[string]any{
		"tool_name": "Bash",
		"args":      []any{"echo", "hello world"},
	}

	out, count := r.Value(input)
	if count != 0 {
		t.Fatalf("got %d redactions, want 0", count)
	}

	outMap := out.(map[string]any)
	if outMap["tool_name"] != "Bash" {
		t.Errorf("tool_name changed: %v", outMap["tool_name"])
	}
}
