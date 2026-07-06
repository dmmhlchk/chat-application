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

// 3. helpers
func (h *DirectChatCreation) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *DirectChatCreation) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, api.ErrorResponse{Error: message})
}

// 4. Handle direct chat creation use case
func (h *DirectChatCreation) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req directChatCreationRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.DirectCreationInput{
		UserID1: req.UserID1,
		UserID2: req.UserID2,
	}

	output, err := h.directChatCreation.Execute(r.Context(), input)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, directChatCreationResponse{
		ChatID: output.ChatID,
	})
}
