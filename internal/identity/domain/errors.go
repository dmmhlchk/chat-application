package domain

import "errors"

type Kind uint8

const (
	KindNotFound     Kind = iota + 1 // 404 - Not Found
	KindConflict                     // 409 - Already exists, duplicate state
	KindValidation                   // 404 - Invalid input
	KindInvalidState                 // 422 - Business rule violation
)

type DomainError struct {
	Kind    Kind
	Message string
}

func newErr(kind Kind, message string) *DomainError {
	return &DomainError{
		Kind:    kind,
		Message: message,
	}
}

// Satisfying the "error" interface
func (de *DomainError) Error() string {
	return de.Message
}

func IsKind(err error, kind Kind) bool {
	var de *DomainError
	return errors.As(err, &de) && de.Kind == kind
}

// ___ User _________________________________________________________________
var (
	ErrUserNotFound        = newErr(KindNotFound, "user not found")
	ErrUserAlreadyExists   = newErr(KindConflict, "user already exists")
	ErrUserInvalidUsername = newErr(KindValidation, "invalid username")
	ErrUserInvalidPhone    = newErr(KindValidation, "invalid phone number")
	ErrUserInvalidID       = newErr(KindValidation, "invalid user id")
)

// ___ Session _________________________________________________________________
var (
	ErrSessionNotFound            = newErr(KindNotFound, "session not found")
	ErrSessionAlreadyRevoked      = newErr(KindConflict, "session already revoked")
	ErrSessionAlreadyClean        = newErr(KindConflict, "no active sessions to clean")
	ErrSessionInvalid             = newErr(KindInvalidState, "session is invalid")
	ErrSessionInvalidRefreshToken = newErr(KindValidation, "invalid refresh token")
)

// ___ Device _________________________________________________________________
var (
	ErrDeviceInvalidHash     = newErr(KindValidation, "invalid device hash")
	ErrDeviceInvalidName     = newErr(KindValidation, "invalid device name")
	ErrDeviceInvalidVersion  = newErr(KindValidation, "invalid device version")
	ErrDeviceInvalidPlatform = newErr(KindValidation, "invalid device platform")
)

// ___ OTP _________________________________________________________________
var (
	ErrOTPInvalid = newErr(KindValidation, "invalid otp code")
	ErrOTPExpired = newErr(KindInvalidState, "otp expired")
)
