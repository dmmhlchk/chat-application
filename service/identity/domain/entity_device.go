package domain

type Device struct {
	Hash     string
	Name     string
	Version  string
	Platform Platform
}

func NewDevice(hash, name, version string, platform Platform) Device {
	return Device{
		Hash:     hash,
		Name:     name,
		Version:  version,
		Platform: platform,
	}
}
