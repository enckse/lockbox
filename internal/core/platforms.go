// Package core defines known platforms
package core

// Platforms are the known platforms for lockbox
var Platforms = PlatformTypes{
	MacOSPlatform:        "macos",
	LinuxWaylandPlatform: "linux-wayland",
	LinuxXPlatform:       "linux-x",
	WindowsLinuxPlatform: "wsl",
}

type (
	// SystemPlatform represents the platform lockbox is running on.
	SystemPlatform string

	// PlatformTypes defines systems lockbox is known to run on or can run on
	PlatformTypes struct {
		MacOSPlatform        SystemPlatform
		LinuxWaylandPlatform SystemPlatform
		LinuxXPlatform       SystemPlatform
		WindowsLinuxPlatform SystemPlatform
	}
)

// List will list the platform types on the struct
func (p PlatformTypes) List() []string {
	return listFields[SystemPlatform](p)
}
