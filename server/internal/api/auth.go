package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/stephenmontague/ttm-tracker/server/internal/config"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PostLogin validates credentials, creates a session, and sets an HTTP-only cookie.
func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.authRepo.GetAdminUserByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	maxAge := config.GetSessionMaxAge()
	token, err := h.authRepo.CreateSession(r.Context(), user.ID, maxAge)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.GetSessionCookieName(),
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   config.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})

	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// PostLogout deletes the session and clears the cookie.
func (h *Handler) PostLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(config.GetSessionCookieName())
	if err == nil {
		_ = h.authRepo.DeleteSession(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.GetSessionCookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   config.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})

	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GetAuthStatus returns 200 if session is valid (placed behind RequireSession middleware).
func (h *Handler) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{"authenticated": true})
}
