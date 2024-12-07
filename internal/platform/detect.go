package platform

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/seanenck/lockbox/internal/config"
	"github.com/seanenck/lockbox/internal/core"
)

const unknownPlatform = ""

// NewPlatform gets a new system platform.
func NewPlatform() (core.SystemPlatform, error) {
	env := config.EnvPlatform.Get()
	if env != "" {
		for _, p := range core.Platforms.List() {
			if p == env {
				return core.SystemPlatform(p), nil
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
		return core.Platforms.MacOSPlatform, nil
	case "linux":
		if strings.Contains(raw, "microsoft-standard-wsl") {
			return core.Platforms.WindowsLinuxPlatform, nil
		}
		if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
			if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
				return unknownPlatform, errors.New("unable to detect linux clipboard mode")
			}
			return core.Platforms.LinuxXPlatform, nil
		}
		return core.Platforms.LinuxWaylandPlatform, nil
	}
	return unknownPlatform, errors.New("unable to detect clipboard mode")
}
