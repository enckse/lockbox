package clip

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/enckse/lockbox/internal/misc"
)

const (
	maxTime         = 45
	pbClipMode      = "pb"
	waylandClipMode = "wayland"
	xClipMode       = "x11"
	wslMode         = "wsl"
)

type (
	// Commands represent system clipboard operations.
	Commands struct {
		copying []string
		pasting []string
		MaxTime int
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
	max := maxTime
	useMax := os.Getenv("LOCKBOX_CLIPMAX")
	if useMax != "" {
		i, err := strconv.Atoi(useMax)
		if err != nil {
			return Commands{}, err
		}
		if i < 1 {
			return Commands{}, errors.New("clipboard max time must be greater than 0")
		}
		max = i
	}
	var copying []string
	var pasting []string
	switch env {
	case pbClipMode:
		copying = []string{"pbcopy"}
		pasting = []string{"pbpaste"}
	case xClipMode:
		copying = []string{"xclip"}
		pasting = []string{"xclip", "-o"}
	case waylandClipMode:
		copying = []string{"wl-copy"}
		pasting = []string{"wl-paste"}
	case wslMode:
		copying = []string{"clip.exe"}
		pasting = []string{"powershell.exe", "-command", "Get-Clipboard"}
	default:
		return Commands{}, errors.New("clipboard is unavailable")
	}
	return Commands{copying: copying, pasting: pasting, MaxTime: max}, nil
}

// CopyTo will copy to clipboard, if non-empty will clear later.
func (c Commands) CopyTo(value, executable string) {
	cmd, args := c.Args(true)
	pipeTo(cmd, value, true, args...)
	if value != "" {
		fmt.Printf("clipboard will clear in %d seconds\n", c.MaxTime)
		pipeTo(filepath.Join(filepath.Dir(executable), "lb"), value, false, "clear")
	}
}

// Args returns clipboard args for execution.
func (c Commands) Args(copying bool) (string, []string) {
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
