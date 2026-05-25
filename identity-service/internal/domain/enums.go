package domain

type Platform string

const (
	PlatformWeb     Platform = "web"
	PlatformWindows Platform = "windows"
	PlatformMacOS   Platform = "macos"
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
)

func (p Platform) IsValid() bool {
	switch p {
	case PlatformWeb, PlatformWindows, PlatformMacOS, PlatformAndroid, PlatformIOS:
		return true
	default:
		return false
	}
}
