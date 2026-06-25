package v1

import (
	"encoding/json"
	"net/http"

	"chat-application/internal/application/usecase"

	"mime"
)

// 1. Determine presentation inputs
type signUpRequestPayload struct {
	Phone string `json:"phone"`
}

// 2. Injects dependencies
type SignUpRequest struct {
	reqSignUp *usecase.SignUpRequest
}

func NewSignUpRequestHandler(reqSignUp *usecase.SignUpRequest) *SignUpRequest {
	return &SignUpRequest{reqSignUp: reqSignUp}
}

// 3. helpers
func (h *SignUpRequest) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *SignUpRequest) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

// 4. Handle sign up request use cases
func (h *SignUpRequest) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	mediatype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediatype != "application/json" {
		h.respondWithError(
			w,
			http.StatusUnsupportedMediaType,
			"unsupported media type: request body must be application/json",
		)
		return
	}

	var req signUpRequestPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.SignUpRequestInput{
		Phone: req.Phone,
	}

	err = h.reqSignUp.Execute(r.Context(), input)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, messageResponse{
		Message: "verification code sent successfully",
	})
}
