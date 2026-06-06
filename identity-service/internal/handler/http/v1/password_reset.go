package v1

import (
	"encoding/json"
	"identity-service/internal/usecase"
	"net/http"
)

// 1. Determine presentation inputs
type passwordResetRequestPayload struct {
	Phone string `json:"phone"`
}

type passwordResetConfirmPayload struct {
	Phone       string `json:"phone"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// 2. Inject dependencies
type PasswordReset struct {
	reqPasswordReset  *usecase.PasswordResetRequest
	confPasswordReset *usecase.PasswordResetConfirm
}

func NewPasswordReset(
	reqPasswordReset *usecase.PasswordResetRequest,
	confPasswordReset *usecase.PasswordResetConfirm,
) *PasswordReset {
	return &PasswordReset{
		reqPasswordReset:  reqPasswordReset,
		confPasswordReset: confPasswordReset,
	}
}

// 3. helpers
func (h *PasswordReset) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *PasswordReset) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

func (h *PasswordReset) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req passwordResetRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input := usecase.PasswordResetRequestInput{Phone: req.Phone}
	if err := h.reqPasswordReset.Execute(r.Context(), input); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, messageResponse{
		Message: "verification code sent via SMS",
	})
}

func (h *PasswordReset) HandleConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req passwordResetConfirmPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	input := usecase.ResetConfirmInput{
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
