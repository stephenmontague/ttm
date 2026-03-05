package api

import (
	"net/http"

	"github.com/stephenmontague/ttm-tracker/server/internal/config"
)

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
