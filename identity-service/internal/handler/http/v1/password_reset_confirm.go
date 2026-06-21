package v1

import (
	"encoding/json"
	"net/http"

	"identity-service/internal/application/usecase"
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

// 3. helpers
func (h *PasswordResetConfirm) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *PasswordResetConfirm) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

// 4. Handle password reset confirm use cases
func (h *PasswordResetConfirm) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req passwordResetConfirmPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input := usecase.PasswordResetConfirmInput{
		Phone:       req.Phone,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	}

	if err := h.confPasswordReset.Execute(r.Context(), input); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	h.respondWithJSON(w, http.StatusOK, messageResponse{
		Message: "password updated successfully and all sessions revoked",
	})
}
