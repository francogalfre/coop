package httpapi

import (
	"net"
	"net/http"

	"github.com/francogalfre/coop/apps/relay/internal/ratelimit"
)

const (
	ingestRatePerSecond = 5.0
	ingestBurst         = 20.0
	steerRatePerSecond  = 1.0
	steerBurst          = 5.0
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

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
