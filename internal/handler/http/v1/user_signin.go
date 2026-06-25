package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"chat-application/internal/application/usecase"
	"chat-application/internal/domain"

	"mime"
)

// 1. Determine presentation input and output
type signInRequestPayload struct {
	Phone             string `json:"phone"`
	Password          string `json:"password"`
	NotificationToken string `json:"notify_token"`
	DeviceHash        string `json:"device_hash"`
	DeviceName        string `json:"device_name"`
	DeviceVersion     string `json:"device_version"`
	DevicePlatform    string `json:"device_platform"`
}

type signInResponsePayload struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// 2. Inject dependencies
type SignIn struct {
	signin *usecase.SignIn
}

func NewSignInHandler(signin *usecase.SignIn) *SignIn {
	return &SignIn{signin: signin}
}

// 3. helpers
func (h *SignIn) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *SignIn) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, errorResponse{Error: message})
}

func (h *SignIn) extractIP(r *http.Request) string {
	// Check if the app is behind a trusted proxy gateway first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0]) // First IP in the chain is the original client
	}

	// Fallback to direct connection address if no proxy header exists
	// r.RemoteAddr usually looks like "127.0.0.1:54321", so we split the port away
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		parts := strings.Split(ip, ":")
		return parts[0]
	}
	return ip
}

// 4. Handle sign in use case
func (h *SignIn) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req signInRequestPayload
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	ipAddress := h.extractIP(r)

	device := domain.Device{
		Hash:     req.DeviceHash,
		Name:     req.DeviceName,
		Version:  req.DeviceVersion,
		Platform: domain.Platform(req.DevicePlatform),
	}

	input := usecase.SignInInput{
		Phone:             req.Phone,
		Password:          req.Password,
		NotificationToken: req.NotificationToken,
		Device:            device,
		IPAddress:         ipAddress,
	}

	output, err := h.signin.Execute(r.Context(), input)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondWithJSON(w, http.StatusOK, signInResponsePayload{
		UserID:       output.UserID,
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	})
}
