package api

import (
	"net/http"

	"github.com/stephenmontague/ttm-tracker/server/internal/config"
)

// SecurityHeaders adds standard security headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// MaxBodySize limits request body size to prevent memory exhaustion.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RequireSession is Chi middleware that validates the session cookie.
func (h *Handler) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(config.GetSessionCookieName())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		_, err = h.authRepo.ValidateSession(r.Context(), cookie.Value, config.GetSessionMaxAge())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid or expired session")
			return
		}

		next.ServeHTTP(w, r)
	})
}
