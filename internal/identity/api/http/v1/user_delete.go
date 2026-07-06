package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/shared/api"

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

// 3. Handle user deletion use case
func (h *UserDeletion) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req userDeleteRequestPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.UserDeleteInput{
		UserID:   req.UserID,
		Password: req.Password,
	}

	err = h.userDeletion.Execute(r.Context(), input)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, api.MessageResponse{
		Message: "user has been removed",
	})
}
