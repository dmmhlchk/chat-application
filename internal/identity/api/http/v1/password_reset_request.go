package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"
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

// 3. helpers
func (h *PasswordResetRequest) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *PasswordResetRequest) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

func (h *PasswordResetRequest) Handle(w http.ResponseWriter, r *http.Request) {
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
