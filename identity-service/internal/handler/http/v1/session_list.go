package v1

import (
	"encoding/json"
	"net/http"

	"identity-service/internal/application/usecase"
)

// 1. Inject dependencies
type SessionList struct {
	sessionList *usecase.SessionList
}

func NewSessionList(sessionList *usecase.SessionList) *SessionList {
	return &SessionList{sessionList: sessionList}
}

// 2. helpers
func (h *SessionList) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *SessionList) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

// 3. Handle retrieving of session list use case
func (h *SessionList) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
	currentRefreshToken := cookie.Value

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.respondWithError(w, http.StatusBadRequest, "missing user_id query parameter")
		return
	}

	input := usecase.SessionListInput{
		UserID:              userID,
		CurrentRefreshToken: currentRefreshToken,
	}

	output, err := h.sessionList.Execute(r.Context(), input)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, output)
}
