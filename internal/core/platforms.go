// Package core defines known platforms
package core

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// Platforms are the known platforms for lockbox
var Platforms = PlatformTypes{
	MacOSPlatform:        "macos",
	LinuxWaylandPlatform: "linux-wayland",
	LinuxXPlatform:       "linux-x",
	WindowsLinuxPlatform: "wsl",
}

const unknownPlatform = ""

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

// NewPlatform gets a new system platform.
func NewPlatform(candidate string) (SystemPlatform, error) {
	env := candidate
	if env != "" {
		for _, p := range Platforms.List() {
			if p == env {
				return SystemPlatform(p), nil
			}
		}
		return unknownPlatform, errors.New("unknown platform mode")
	}
	b, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return unknownPlatform, err
	}
	raw := strings.ToLower(strings.TrimSpace(string(b)))
	parts := strings.Split(raw, " ")
	switch parts[0] {
	case "darwin":
		return Platforms.MacOSPlatform, nil
	case "linux":
		if strings.Contains(raw, "microsoft-standard-wsl") {
			return Platforms.WindowsLinuxPlatform, nil
		}
		if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
			if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
				return unknownPlatform, errors.New("unable to detect linux clipboard mode")
			}
			return Platforms.LinuxXPlatform, nil
		}
		return Platforms.LinuxWaylandPlatform, nil
	}
	return unknownPlatform, errors.New("unable to detect clipboard mode")
}
