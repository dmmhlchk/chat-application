package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/shared/api"
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

// 4. Handle password change use case
func (h *ChangePassword) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		api.RespondWithError(w, http.StatusBadRequest, "missing user_id query parameter")
		return
	}

	var req changePasswordRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input := usecase.ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.changePassword.Execute(r.Context(), input); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, api.MessageResponse{
		Message: "password changed successfully",
	})
}
