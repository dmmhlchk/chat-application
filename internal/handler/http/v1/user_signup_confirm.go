package v1

import (
	"encoding/json"
	"internal/application/usecase"
	"mime"
	"net/http"
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

// 3. helpers
func (h *SignUpConfirm) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *SignUpConfirm) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

// 4. Handle sign up confirm use cases
func (h *SignUpConfirm) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req signUpConfirmPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json payload")
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
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, messageResponse{
		Message: "user registered successfully",
	})
}
