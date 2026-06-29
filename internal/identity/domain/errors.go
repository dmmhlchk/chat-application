package domain

import "errors"

var (
	// user errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUsername   = errors.New("invalid username")
	ErrInvalidPhone      = errors.New("invalid phone number")
)

var (
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionAlreadyRevoked = errors.New("session already revoked")
	ErrSessionInvalid        = errors.New("session is invalid")
	ErrInvalidUserID         = errors.New("invalid user id")
	ErrAlreadyCleanSessions  = errors.New("user doesn't have any active session")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token hash")
	ErrInvalidTTL            = errors.New("invalid ttl")
)

var (
	ErrOTPInvalid = errors.New("invalid otp code")
	ErrOTPExpired = errors.New("otp expired")
)

var (
	ErrInvalidDeviceHash    = errors.New("invalid device hash")
	ErrInvalidDeviceName    = errors.New("invalid device name")
	ErrInvalidDeviceVersion = errors.New("invalid device version")
	ErrInvalidPlatform      = errors.New("invalid platform")
)
