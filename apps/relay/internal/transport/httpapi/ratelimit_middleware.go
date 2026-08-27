package httpapi

import (
	"net"
	"net/http"
	"strings"

	"github.com/francogalfre/coop/apps/relay/internal/ratelimit"
)

const (
	ingestRatePerSecond   = 5.0
	ingestBurst           = 20.0
	steerRatePerSecond    = 1.0
	steerBurst            = 5.0
	exchangeRatePerSecond = 3.0 / 60.0
	exchangeBurst         = 3.0
)

func withIPRateLimit(limiter *ratelimit.Limiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow(clientIP(r)) {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next(w, r)
	}
}

// X-Forwarded-For/X-Real-IP are only trustworthy when the proxy in front of the relay strips client-supplied values, which is a deployment concern, not this code's to enforce.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		first, _, _ := strings.Cut(fwd, ",")
		return strings.TrimSpace(first)
	}

	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
