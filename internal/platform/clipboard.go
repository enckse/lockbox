// Package platform handles platform-specific operations around clipboards.
package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	osc "github.com/aymanbagabas/go-osc52"
	"github.com/enckse/lockbox/internal/inputs"
)

type (
	// Clipboard represent system clipboard operations.
	Clipboard struct {
		copying []string
		pasting []string
		MaxTime int
		isOSC52 bool
	}
)

func newClipboard(copying, pasting []string) (Clipboard, error) {
	max, err := inputs.GetClipboardMax()
	if err != nil {
		return Clipboard{}, err
	}
	return Clipboard{copying: copying, pasting: pasting, MaxTime: max, isOSC52: false}, nil
}

func overrideCommand(v string) ([]string, error) {
	value := inputs.EnvOrDefault(v, "")
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	return inputs.Shlex(value)
}

// NewClipboard will retrieve the commands to use for clipboard operations.
func NewClipboard() (Clipboard, error) {
	noClip, err := inputs.IsNoClipEnabled()
	if err != nil {
		return Clipboard{}, err
	}
	if noClip {
		return Clipboard{}, errors.New("clipboard is off")
	}
	overridePaste, err := overrideCommand(inputs.ClipPasteEnv)
	if err != nil {
		return Clipboard{}, err
	}
	overrideCopy, err := overrideCommand(inputs.ClipCopyEnv)
	if err != nil {
		return Clipboard{}, err
	}
	if overrideCopy != nil && overridePaste != nil {
		return newClipboard(overrideCopy, overridePaste)
	}
	isOSC, err := inputs.IsClipOSC52()
	if err != nil {
		return Clipboard{}, err
	}
	if isOSC {
		c := Clipboard{isOSC52: true}
		return c, nil
	}
	sys, err := NewPlatform()
	if err != nil {
		return Clipboard{}, err
	}

	var copying []string
	var pasting []string
	switch sys {
	case inputs.MacOSPlatform:
		copying = []string{"pbcopy"}
		pasting = []string{"pbpaste"}
	case inputs.LinuxXPlatform:
		copying = []string{"xclip"}
		pasting = []string{"xclip", "-o"}
	case inputs.LinuxWaylandPlatform:
		copying = []string{"wl-copy"}
		pasting = []string{"wl-paste"}
	case inputs.WindowsLinuxPlatform:
		copying = []string{"clip.exe"}
		pasting = []string{"powershell.exe", "-command", "Get-Clipboard"}
	default:
		return Clipboard{}, errors.New("clipboard is unavailable")
	}
	if overridePaste != nil {
		pasting = overridePaste
	}
	if overrideCopy != nil {
		copying = overrideCopy
	}
	return newClipboard(copying, pasting)
}

// CopyTo will copy to clipboard, if non-empty will clear later.
func (c Clipboard) CopyTo(value string) error {
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
func (c Clipboard) Args(copying bool) (string, []string, bool) {
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
