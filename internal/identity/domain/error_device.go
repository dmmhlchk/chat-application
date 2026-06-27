package domain

import "errors"

var (
	ErrInvalidDeviceHash    = errors.New("invalid device hash")
	ErrInvalidDeviceName    = errors.New("invalid device name")
	ErrInvalidDeviceVersion = errors.New("invalid device version")
	ErrInvalidPlatform      = errors.New("invalid platform")
)
