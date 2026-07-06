package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/shared/api"

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

// 3. Handle sign up request use cases
func (h *SignUpRequest) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	mediatype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediatype != "application/json" {
		api.RespondWithError(
			w,
			http.StatusUnsupportedMediaType,
			"unsupported media type: request body must be application/json",
		)
		return
	}

	var req signUpRequestPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.SignUpRequestInput{
		Phone: req.Phone,
	}

	err = h.reqSignUp.Execute(r.Context(), input)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, api.MessageResponse{
		Message: "verification code sent successfully",
	})
}
