package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// MaxClipTime is the max time to let something stay in the clipboard.
	MaxClipTime     = 45
	pbClipMode      = "pb"
	waylandClipMode = "wayland"
	xClipMode       = "x11"
	wslMode         = "wsl"
)

// GetClipboardCommand will retrieve the commands to use for clipboard operations.
func GetClipboardCommand() ([]string, []string, error) {
	env := strings.TrimSpace(os.Getenv("LOCKBOX_CLIPMODE"))
	if env == "" {
		b, err := exec.Command("uname", "-a").Output()
		if err != nil {
			return nil, nil, err
		}
		raw := strings.TrimSpace(string(b))
		parts := strings.Split(raw, "")
		switch parts[0] {
		case "Darwin":
			env = pbClipMode
		case "Linux":
			if strings.Contains(raw, "microsoft-standard-WSL2") {
				env = wslMode
			} else {
				if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
					if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
						return nil, nil, NewLockboxError("unable to detect linux clipboard mode")
					}
					env = xClipMode
				} else {
					env = waylandClipMode
				}
			}
		default:
			return nil, nil, NewLockboxError("unable to detect clipboard mode")
		}
	}
	switch env {
	case pbClipMode:
		return []string{"pbcopy"}, []string{"pbpaste"}, nil
	case xClipMode:
		return []string{"xclip"}, []string{"xclip", "-o"}, nil
	case waylandClipMode:
		return []string{"wl-copy"}, []string{"wl-paste"}, nil
	case wslMode:
		return []string{"clip.exe"}, []string{"powershell.exe", "-command", "Get-Clipboard"}, nil
	case "off":
		return nil, nil, NewLockboxError("clipboard is turned off")
	}
	return nil, nil, NewLockboxError("unable to get clipboard command(s)")
}

// CopyToClipboard will copy to clipboard, if non-empty will clear later.
func CopyToClipboard(value string) {
	cp, _, err := GetClipboardCommand()
	if err != nil {
		fmt.Printf("unable to copy to clipboard: %v\n", err)
		return
	}
	var args []string
	if len(cp) > 1 {
		args = cp[1:]
	}
	pipeTo(cp[0], value, true, args...)
	if value != "" {
		fmt.Printf("clipboard will clear in %d seconds\n", MaxClipTime)
		exe, err := os.Executable()
		if err != nil {
			Die("unable to get executable", err)
		}
		pipeTo(filepath.Join(filepath.Dir(exe), "lb"), value, false, "clear")
	}
}

func pipeTo(command, value string, wait bool, args ...string) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		Die("unable to get stdin pipe", err)
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
		Die("failed to run command", ran)
	}
}
