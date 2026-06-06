package v1

import (
	"encoding/json"
	"identity-service/internal/usecase"
	"mime"
	"net/http"
)

// 1. Determine presentation inputs
type signUpRequestPayload struct {
	Phone string `json:"phone"`
}

type signUpConfirmPayload struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// 2. Injects dependencies
type SignUp struct {
	reqSignUp  *usecase.SignUpRequest
	confSignUp *usecase.SignUpConfirm
}

func NewSignUpHandler(reqSignUp *usecase.SignUpRequest, confSignUp *usecase.SignUpConfirm) *SignUp {
	return &SignUp{
		reqSignUp:  reqSignUp,
		confSignUp: confSignUp,
	}
}

// 3. helpers
func (h *SignUp) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *SignUp) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

// 4. Handle sign up request and confirm use cases
func (h *SignUp) HandleRequest(w http.ResponseWriter, r *http.Request) {
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

func (h *SignUp) HandleConfirm(w http.ResponseWriter, r *http.Request) {
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
