// Package platform handles platform-specific operations.
package platform

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/enckse/lockbox/internal/inputs"
)

const (
	unknownPlatform = ""
)

// NewPlatform gets a new system platform.
func NewPlatform() (inputs.SystemPlatform, error) {
	env := os.Getenv(inputs.PlatformEnv)
	if env != "" {
		switch env {
		case inputs.MacOSPlatform:
			return inputs.MacOSPlatform, nil
		case inputs.LinuxWaylandPlatform:
			return inputs.LinuxWaylandPlatform, nil
		case inputs.WindowsLinuxPlatform:
			return inputs.WindowsLinuxPlatform, nil
		case inputs.LinuxXPlatform:
			return inputs.LinuxXPlatform, nil
		default:
			return unknownPlatform, errors.New("unknown platform mode")
		}
	}
	b, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return unknownPlatform, err
	}
	raw := strings.ToLower(strings.TrimSpace(string(b)))
	parts := strings.Split(raw, " ")
	switch parts[0] {
	case "darwin":
		return inputs.MacOSPlatform, nil
	case "linux":
		if strings.Contains(raw, "microsoft-standard-wsl") {
			return inputs.WindowsLinuxPlatform, nil
		}
		if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
			if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
				return unknownPlatform, errors.New("unable to detect linux clipboard mode")
			}
			return inputs.LinuxXPlatform, nil
		}
		return inputs.LinuxWaylandPlatform, nil
	}
	return unknownPlatform, errors.New("unable to detect clipboard mode")
}
