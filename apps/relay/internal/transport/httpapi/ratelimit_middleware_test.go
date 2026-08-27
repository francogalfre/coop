package httpapi

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPPrefersForwardedHeaders(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xRealIP    string
		want       string
	}{
		{"falls back to RemoteAddr", "1.2.3.4:5678", "", "", "1.2.3.4"},
		{"uses X-Forwarded-For first entry", "1.2.3.4:5678", "5.6.7.8, 9.10.11.12", "", "5.6.7.8"},
		{"uses X-Real-IP when no X-Forwarded-For", "1.2.3.4:5678", "", "5.6.7.8", "5.6.7.8"},
		{"prefers X-Forwarded-For over X-Real-IP", "1.2.3.4:5678", "5.6.7.8", "9.10.11.12", "5.6.7.8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			if got := clientIP(req); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
