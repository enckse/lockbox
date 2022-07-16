package clipboard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/enckse/lockbox/internal/misc"
)

const (
	// MaxTime is the max time to let something stay in the clipboard.
	MaxTime         = 45
	pbClipMode      = "pb"
	waylandClipMode = "wayland"
	xClipMode       = "x11"
	wslMode         = "wsl"
)

type (
	// Commands represent system clipboard operations.
	Commands struct {
		Copy  []string
		Paste []string
	}
)

// NewCommands will retrieve the commands to use for clipboard operations.
func NewCommands() (Commands, error) {
	env := strings.TrimSpace(os.Getenv("LOCKBOX_CLIPMODE"))
	if env == "" {
		b, err := exec.Command("uname", "-a").Output()
		if err != nil {
			return Commands{}, err
		}
		raw := strings.TrimSpace(string(b))
		parts := strings.Split(raw, " ")
		switch parts[0] {
		case "Darwin":
			env = pbClipMode
		case "Linux":
			if strings.Contains(raw, "microsoft-standard-WSL2") {
				env = wslMode
			} else {
				if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
					if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
						return Commands{}, errors.New("unable to detect linux clipboard mode")
					}
					env = xClipMode
				} else {
					env = waylandClipMode
				}
			}
		default:
			return Commands{}, errors.New("unable to detect clipboard mode")
		}
	}
	switch env {
	case pbClipMode:
		return Commands{Copy: []string{"pbcopy"}, Paste: []string{"pbpaste"}}, nil
	case xClipMode:
		return Commands{Copy: []string{"xclip"}, Paste: []string{"xclip", "-o"}}, nil
	case waylandClipMode:
		return Commands{Copy: []string{"wl-copy"}, Paste: []string{"wl-paste"}}, nil
	case wslMode:
		return Commands{Copy: []string{"clip.exe"}, Paste: []string{"powershell.exe", "-command", "Get-Clipboard"}}, nil
	default:
		return Commands{}, errors.New("clipboard is unavailable")
	}
}

// CopyToClipboard will copy to clipboard, if non-empty will clear later.
func (c Commands) CopyToClipboard(value, executable string) {
	var args []string
	if len(c.Copy) > 1 {
		args = c.Copy[1:]
	}
	pipeTo(c.Copy[0], value, true, args...)
	if value != "" {
		fmt.Printf("clipboard will clear in %d seconds\n", MaxTime)
		pipeTo(filepath.Join(filepath.Dir(executable), "lb"), value, false, "clear")
	}
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
