package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"

	"mime"
)

// 1. Determine presentation input
type terminateSessionRequestPayload struct {
	UserID string `json:"user_id"`
}

// 2. Inject dependencies
type TerminateSession struct {
	terminateSession *usecase.TerminateSession
}

func NewTerminateSession(terminateSession *usecase.TerminateSession) *TerminateSession {
	return &TerminateSession{terminateSession: terminateSession}
}

// 3. helpers
func (h *TerminateSession) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *TerminateSession) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

// 4. Handle session termination use case
func (h *TerminateSession) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			h.respondWithError(w, http.StatusUnauthorized, "missing refresh token cookie")
			return
		}
		h.respondWithError(w, http.StatusBadRequest, "error reading cookie")
		return
	}
	refreshToken := cookie.Value

	mediatype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediatype != "application/json" {
		h.respondWithError(
			w,
			http.StatusUnsupportedMediaType,
			"unsupported media type: request body must be application/json",
		)
		return
	}

	var req terminateSessionRequestPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.TerminateSessionInput{
		UserID:       req.UserID,
		RefreshToken: refreshToken,
	}

	err = h.terminateSession.Execute(r.Context(), input)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, messageResponse{
		Message: "session has been terminated",
	})
}
