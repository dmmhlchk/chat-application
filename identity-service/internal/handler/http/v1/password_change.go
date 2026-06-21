package v1

import (
	"encoding/json"
	"net/http"

	"identity-service/internal/application/usecase"
)

// 1. Determine presentation inputs
// We don't put user_id here because it should come securely from the request context or params, not the body.
type changePasswordRequestPayload struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// 2. Inject dependencies
type ChangePassword struct {
	changePassword *usecase.ChangePassword
}

func NewChangePassword(changePassword *usecase.ChangePassword) *ChangePassword {
	return &ChangePassword{changePassword: changePassword}
}

// 3. helpers
func (h *ChangePassword) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *ChangePassword) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

// 4. Handle password change use case
func (h *ChangePassword) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.respondWithError(w, http.StatusBadRequest, "missing user_id query parameter")
		return
	}

	var req changePasswordRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input := usecase.ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.changePassword.Execute(r.Context(), input); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, messageResponse{
		Message: "password changed successfully",
	})
}
