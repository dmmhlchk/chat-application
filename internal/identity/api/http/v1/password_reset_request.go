package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/shared/api"
)

// 1. Determine presentation inputs
type passwordResetRequestPayload struct {
	Phone string `json:"phone"`
}

// 2. Inject dependencies
type PasswordResetRequest struct {
	reqPasswordReset *usecase.PasswordResetRequest
}

func NewPasswordResetRequest(reqPasswordReset *usecase.PasswordResetRequest) *PasswordResetRequest {
	return &PasswordResetRequest{reqPasswordReset: reqPasswordReset}
}

// Handle password reset request use case
func (h *PasswordResetRequest) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req passwordResetRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input := usecase.PasswordResetRequestInput{Phone: req.Phone}
	if err := h.reqPasswordReset.Execute(r.Context(), input); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, api.MessageResponse{
		Message: "verification code sent via SMS",
	})
}
