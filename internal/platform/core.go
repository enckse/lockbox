// Package platform defines known platforms
package platform

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/seanenck/lockbox/internal/util"
)

// Systems are the known platforms for lockbox
var Systems = SystemTypes{
	MacOSSystem:        "macos",
	LinuxWaylandSystem: "linux-wayland",
	LinuxXSystem:       "linux-x",
	WindowsLinuxSystem: "wsl",
}

const unknownSystem = ""

type (
	// System represents the platform lockbox is running on.
	System string

	// SystemTypes defines systems lockbox is known to run on or can run on
	SystemTypes struct {
		MacOSSystem        System
		LinuxWaylandSystem System
		LinuxXSystem       System
		WindowsLinuxSystem System
	}
)

// List will list the platform types on the struct
func (p SystemTypes) List() []string {
	return util.ListFields(p)
}

// NewSystem gets a new system platform.
func NewSystem(candidate string) (System, error) {
	env := candidate
	if env != "" {
		for _, p := range Systems.List() {
			if p == env {
				return System(p), nil
			}
		}
		return unknownSystem, errors.New("unknown platform mode")
	}
	b, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return unknownSystem, err
	}
	raw := strings.ToLower(strings.TrimSpace(string(b)))
	parts := strings.Split(raw, " ")
	switch parts[0] {
	case "darwin":
		return Systems.MacOSSystem, nil
	case "linux":
		if strings.Contains(raw, "microsoft-standard-wsl") {
			return Systems.WindowsLinuxSystem, nil
		}
		if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
			if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
				return unknownSystem, errors.New("unable to detect linux clipboard mode")
			}
			return Systems.LinuxXSystem, nil
		}
		return Systems.LinuxWaylandSystem, nil
	}
	return unknownSystem, errors.New("unable to detect clipboard mode")
}
