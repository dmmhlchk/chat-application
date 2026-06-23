package domain

type Device struct {
	Hash     string
	Name     string
	Version  string
	Platform Platform
}

func NewDevice(hash, name, version string, platform Platform) (Device, error) {
	if hash == "" {
		return Device{}, ErrInvalidDeviceHash
	}
	if name == "" {
		return Device{}, ErrInvalidDeviceName
	}
	if version == "" {
		return Device{}, ErrInvalidDeviceVersion
	}
	if !platform.IsValid() {
		return Device{}, ErrInvalidPlatform
	}

	return Device{
		Hash:     hash,
		Name:     name,
		Version:  version,
		Platform: platform,
	}, nil
}
