// Package clip handles platform-specific operations around clipboards.
package clip

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	osc "github.com/aymanbagabas/go-osc52"
	"github.com/seanenck/lockbox/internal/config"
	"github.com/seanenck/lockbox/internal/platform"
)

type (
	// Board represent system clipboard operations.
	Board struct {
		copying []string
		pasting []string
		MaxTime int64
		isOSC52 bool
	}
)

func newBoard(copying, pasting []string) (Board, error) {
	maximum, err := config.EnvClipTimeout.Get()
	if err != nil {
		return Board{}, err
	}
	return Board{copying: copying, pasting: pasting, MaxTime: maximum, isOSC52: false}, nil
}

// New will retrieve the commands to use for clipboard operations.
func New() (Board, error) {
	if !config.EnvClipEnabled.Get() {
		return Board{}, errors.New("clipboard is off")
	}
	overridePaste := config.EnvClipPaste.Get()
	overrideCopy := config.EnvClipCopy.Get()
	setPaste := len(overridePaste) > 0
	setCopy := len(overrideCopy) > 0
	if setPaste && setCopy {
		return newBoard(overrideCopy, overridePaste)
	}
	if config.EnvClipOSC52.Get() {
		c := Board{isOSC52: true}
		return c, nil
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
	if c.isOSC52 {
		osc.Copy(value)
		return nil
	}
	cmd, args, _ := c.Args(true)
	pipeTo(cmd, value, true, args...)
	if value != "" {
		fmt.Printf("clipboard will clear in %d seconds\n", c.MaxTime)
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		pipeTo(exe, value, false, "clear")
	}
	return nil
}

// Args returns clipboard args for execution.
func (c Board) Args(copying bool) (string, []string, bool) {
	if c.isOSC52 {
		return "", []string{}, false
	}
	var using []string
	if copying {
		using = c.copying
	} else {
		using = c.pasting
	}
	var args []string
	if len(using) > 1 {
		args = using[1:]
	}
	return using[0], args, true
}

func pipeTo(command, value string, wait bool, args ...string) error {
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
	var ran error
	if wait {
		ran = cmd.Run()
	} else {
		ran = cmd.Start()
	}
	if ran != nil {
		return errors.New("failed to run command")
	}
	return nil
}
