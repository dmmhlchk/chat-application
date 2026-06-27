package domain

import "errors"

var (
	ErrOTPInvalid = errors.New("invalid otp code")
	ErrOTPExpired = errors.New("otp expired")
)
