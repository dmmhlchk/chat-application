package v1

import (
	"encoding/json"
	"net/http"

	"chat-app/internal/messenger/application/usecase"
	"chat-app/internal/shared/api"

	"mime"
)

// 1. Determine presentation input and output
type directChatCreationRequest struct {
	UserID1 string
	UserID2 string
}

type directChatCreationResponse struct {
	ChatID string `json:"chat_id"`
}

// 2. Inject dependencies
type DirectChatCreation struct {
	directChatCreation *usecase.DirectCreation
}

func NewDirectChatCreation(uc *usecase.DirectCreation) *DirectChatCreation {
	return &DirectChatCreation{directChatCreation: uc}
}

// 3. Handle direct chat creation use case
func (h *DirectChatCreation) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req directChatCreationRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.DirectCreationInput{
		UserID1: req.UserID1,
		UserID2: req.UserID2,
	}

	output, err := h.directChatCreation.Execute(r.Context(), input)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, directChatCreationResponse{
		ChatID: output.ChatID,
	})
}
