package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/shared/api"
)

// 1. Determine presentation inputs
type passwordResetConfirmPayload struct {
	Phone       string `json:"phone"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// 2. Inject dependencies
type PasswordResetConfirm struct {
	confPasswordReset *usecase.PasswordResetConfirm
}

func NewPasswordResetConfirm(confPasswordReset *usecase.PasswordResetConfirm) *PasswordResetConfirm {
	return &PasswordResetConfirm{confPasswordReset: confPasswordReset}
}

// 3. Handler password reset confirm use case
func (h *PasswordResetConfirm) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req passwordResetConfirmPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input := usecase.PasswordResetConfirmInput{
		Phone:       req.Phone,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	}

	if err := h.confPasswordReset.Execute(r.Context(), input); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	api.RespondWithJSON(w, http.StatusOK, api.MessageResponse{
		Message: "password updated successfully and all sessions revoked",
	})
}
