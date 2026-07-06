package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/shared/api"

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

// 3. Handle session termination use case
func (h *TerminateSession) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		if err == http.ErrNoCookie {
			api.RespondWithError(w, http.StatusUnauthorized, "missing refresh token cookie")
			return
		}
		api.RespondWithError(w, http.StatusBadRequest, "error reading cookie")
		return
	}
	refreshToken := cookie.Value

	mediatype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediatype != "application/json" {
		api.RespondWithError(
			w,
			http.StatusUnsupportedMediaType,
			"unsupported media type: request body must be application/json",
		)
		return
	}

	var req terminateSessionRequestPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.TerminateSessionInput{
		UserID:       req.UserID,
		RefreshToken: refreshToken,
	}

	err = h.terminateSession.Execute(r.Context(), input)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, api.MessageResponse{
		Message: "session has been terminated",
	})
}
