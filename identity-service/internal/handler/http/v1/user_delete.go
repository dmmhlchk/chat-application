package v1

import (
	"encoding/json"
	"net/http"

	"identity-service/internal/application/usecase"

	"mime"
)

// 1. Determine presentation input
type userDeleteRequestPayload struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

// 2. Inject dependencies
type UserDeletion struct {
	userDeletion *usecase.UserDelete
}

func NewUserDeletion(userDeletion *usecase.UserDelete) *UserDeletion {
	return &UserDeletion{userDeletion: userDeletion}
}

// 3. helpers
func (h *UserDeletion) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *UserDeletion) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

// 4. Handle user deletion use case
func (h *UserDeletion) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req userDeleteRequestPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.UserDeleteInput{
		UserID:   req.UserID,
		Password: req.Password,
	}

	err = h.userDeletion.Execute(r.Context(), input)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, messageResponse{
		Message: "user has been removed",
	})
}
