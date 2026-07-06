package v1

import (
	"net/http"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/shared/api"
)

// 1. Inject dependencies
type SessionList struct {
	sessionList *usecase.SessionList
}

func NewSessionList(sessionList *usecase.SessionList) *SessionList {
	return &SessionList{sessionList: sessionList}
}

// 2. Handle retrieving of session list use case
func (h *SessionList) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
	currentRefreshToken := cookie.Value

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		api.RespondWithError(w, http.StatusBadRequest, "missing user_id query parameter")
		return
	}

	input := usecase.SessionListInput{
		UserID:              userID,
		CurrentRefreshToken: currentRefreshToken,
	}

	output, err := h.sessionList.Execute(r.Context(), input)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, output)
}
