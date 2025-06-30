// Package clip handles platform-specific operations around clipboards.
package clip

import (
	"errors"
	"fmt"
	"os/exec"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/platform"
)

type (
	// Board represent system clipboard operations.
	Board struct {
		PIDFile string
		copying []string
		pasting []string
		MaxTime int64
	}
)

func newBoard(copying, pasting []string) (Board, error) {
	maximum, err := config.EnvClipTimeout.Get()
	if err != nil {
		return Board{}, err
	}
	pid := config.EnvClipProcessFile.Get()
	return Board{copying: copying, pasting: pasting, MaxTime: maximum, PIDFile: pid}, nil
}

// New creates a new clipboard
func New() (Board, error) {
	if !config.EnvFeatureClip.Get() {
		return Board{}, config.NewFeatureError("clip")
	}
	overridePaste := config.EnvClipPaste.Get()
	overrideCopy := config.EnvClipCopy.Get()
	setPaste := len(overridePaste) > 0
	setCopy := len(overrideCopy) > 0
	if setPaste && setCopy {
		return newBoard(overrideCopy, overridePaste)
	}
	sys, err := platform.NewSystem(config.EnvPlatform.Get())
	if err != nil {
		return Board{}, err
	}

	var copying []string
	var pasting []string
	switch sys {
	case platform.Systems.MacOSSystem:
		copying = []string{"pbcopy"}
		pasting = []string{"pbpaste"}
	case platform.Systems.LinuxXSystem:
		copying = []string{"xclip"}
		pasting = []string{"xclip", "-o"}
	case platform.Systems.LinuxWaylandSystem:
		copying = []string{"wl-copy"}
		pasting = []string{"wl-paste"}
	case platform.Systems.WindowsLinuxSystem:
		copying = []string{"clip.exe"}
		pasting = []string{"powershell.exe", "-command", "Get-Clipboard"}
	default:
		return Board{}, errors.New("clipboard is unavailable")
	}
	if setPaste {
		pasting = overridePaste
	}
	if setCopy {
		copying = overrideCopy
	}
	return newBoard(copying, pasting)
}

// CopyTo will copy to clipboard, if non-empty will clear later.
func (c Board) CopyTo(value string) error {
	cmd, args, err := c.Args(true)
	if err != nil {
		return err
	}
	pipeTo(cmd, value, args...)
	return nil
}

// Args returns clipboard args for execution.
func (c Board) Args(copying bool) (string, []string, error) {
	var using []string
	if copying {
		using = c.copying
	} else {
		using = c.pasting
	}
	if len(using) == 0 {
		return "", nil, fmt.Errorf("command is not set (copying? %v)", copying)
	}
	var args []string
	if len(using) > 1 {
		args = using[1:]
	}
	return using[0], args, nil
}

func pipeTo(command, value string, args ...string) error {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		if _, err := stdin.Write([]byte(value)); err != nil {
			fmt.Printf("failed writing to stdin: %v\n", err)
		}
	}()
	return cmd.Run()
}
