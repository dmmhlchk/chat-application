package domain

import "errors"

var (
	// user errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPhone      = errors.New("invalid phone number")

	// session errors
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionAlreadyRevoked = errors.New("session already revoked")
	ErrSessionInvalid        = errors.New("session is invalid")
	ErrInvalidUserID         = errors.New("invalid user id")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token hash")
	ErrInvalidTTL            = errors.New("invalid ttl")

	// device errors
	ErrInvalidDeviceHash    = errors.New("invalid device hash")
	ErrInvalidDeviceName    = errors.New("invalid device name")
	ErrInvalidDeviceVersion = errors.New("invalid device version")
	ErrInvalidPlatform      = errors.New("invalid platform")
)
