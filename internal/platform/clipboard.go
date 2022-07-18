package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/misc"
)

const (
	maxTime = 45
)

type (
	// Clipboard represent system clipboard operations.
	Clipboard struct {
		copying []string
		pasting []string
		MaxTime int
	}
)

// NewClipboard will retrieve the commands to use for clipboard operations.
func NewClipboard() (Clipboard, error) {
	noClip, err := inputs.IsNoClipEnabled()
	if err != nil {
		return Clipboard{}, err
	}
	if noClip {
		return Clipboard{}, errors.New("clipboard is off")
	}
	sys, err := NewPlatform()
	if err != nil {
		return Clipboard{}, err
	}
	max := maxTime
	useMax := os.Getenv(inputs.ClipMaxEnv)
	if useMax != "" {
		i, err := strconv.Atoi(useMax)
		if err != nil {
			return Clipboard{}, err
		}
		if i < 1 {
			return Clipboard{}, errors.New("clipboard max time must be greater than 0")
		}
		max = i
	}
	var copying []string
	var pasting []string
	switch sys {
	case MacOS:
		copying = []string{"pbcopy"}
		pasting = []string{"pbpaste"}
	case LinuxX:
		copying = []string{"xclip"}
		pasting = []string{"xclip", "-o"}
	case LinuxWayland:
		copying = []string{"wl-copy"}
		pasting = []string{"wl-paste"}
	case WindowsLinux:
		copying = []string{"clip.exe"}
		pasting = []string{"powershell.exe", "-command", "Get-Clipboard"}
	default:
		return Clipboard{}, errors.New("clipboard is unavailable")
	}
	return Clipboard{copying: copying, pasting: pasting, MaxTime: max}, nil
}

// CopyTo will copy to clipboard, if non-empty will clear later.
func (c Clipboard) CopyTo(value, executable string) {
	cmd, args := c.Args(true)
	pipeTo(cmd, value, true, args...)
	if value != "" {
		fmt.Printf("clipboard will clear in %d seconds\n", c.MaxTime)
		pipeTo(filepath.Join(filepath.Dir(executable), "lb"), value, false, "clear")
	}
}

// Args returns clipboard args for execution.
func (c Clipboard) Args(copying bool) (string, []string) {
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
	return using[0], args
}

func pipeTo(command, value string, wait bool, args ...string) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		misc.Die("unable to get stdin pipe", err)
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
		misc.Die("failed to run command", ran)
	}
}
