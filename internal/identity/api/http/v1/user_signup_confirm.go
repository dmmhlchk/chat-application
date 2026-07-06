package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/shared/api"

	"mime"
)

// 1. Determine presentation inputs
type signUpConfirmPayload struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// 2. Injects dependencies
type SignUpConfirm struct {
	confSignUp *usecase.SignUpConfirm
}

func NewSignUpConfirmHandler(confSignUp *usecase.SignUpConfirm) *SignUpConfirm {
	return &SignUpConfirm{confSignUp: confSignUp}
}

// 3. Handle sign up confirm use cases
func (h *SignUpConfirm) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req signUpConfirmPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.SignUpConfirmInput{
		Username: req.Username,
		Phone:    req.Phone,
		Code:     req.Code,
		Password: req.Password,
	}

	err = h.confSignUp.Execute(r.Context(), input)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, api.MessageResponse{
		Message: "user registered successfully",
	})
}
