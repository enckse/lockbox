package platform

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

type (
	// System represents the platform lockbox is running on.
	System string
)

const (
	// MacOS based systems.
	MacOS System = "macos"
	// LinuxWayland running Wayland.
	LinuxWayland System = "linux-wayland"
	// LinuxX running X.
	LinuxX System = "linux-x"
	// WindowsLinux with WSL.
	WindowsLinux System = "wsl"
	// Unknown platform.
	Unknown = ""
)

// NewPlatform gets a new system platform.
func NewPlatform() (System, error) {
	env := os.Getenv("LOCKBOX_PLATFORM")
	if env != "" {
		switch env {
		case string(MacOS):
			return MacOS, nil
		case string(LinuxWayland):
			return LinuxWayland, nil
		case string(WindowsLinux):
			return WindowsLinux, nil
		case string(LinuxX):
			return LinuxX, nil
		default:
			return Unknown, errors.New("unknown platform mode")
		}
	}
	b, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return Unknown, err
	}
	raw := strings.TrimSpace(string(b))
	parts := strings.Split(raw, " ")
	switch parts[0] {
	case "Darwin":
		return MacOS, nil
	case "Linux":
		if strings.Contains(raw, "microsoft-standard-WSL2") {
			return WindowsLinux, nil
		}
		if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
			if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
				return Unknown, errors.New("unable to detect linux clipboard mode")
			}
			return LinuxX, nil
		}
		return LinuxWayland, nil
	}
	return Unknown, errors.New("unable to detect clipboard mode")
}
