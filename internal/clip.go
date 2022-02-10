package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"voidedtech.com/stock"
)

const (
	// MaxClipTime is the max time to let something stay in the clipboard.
	MaxClipTime     = 45
	pbClipMode      = "pb"
	waylandClipMode = "wayland"
	xClipMode       = "x11"
)

// GetClipboardCommand will retrieve the commands to use for clipboard operations.
func GetClipboardCommand() ([]string, []string, error) {
	env := strings.TrimSpace(os.Getenv("LOCKBOX_CLIPMODE"))
	if env == "" {
		b, err := exec.Command("uname").Output()
		if err != nil {
			return nil, nil, err
		}
		uname := strings.TrimSpace(string(b))
		switch uname {
		case "Darwin":
			env = pbClipMode
		case "Linux":
			if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
				if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
					return nil, nil, stock.NewBasicError("unable to detect linux clipboard mode")
				}
				env = xClipMode
			} else {
				env = waylandClipMode
			}
		default:
			return nil, nil, stock.NewBasicError("unable to detect clipboard mode")
		}
	}
	switch env {
	case pbClipMode:
		return []string{"pbcopy"}, []string{"pbpaste"}, nil
	case xClipMode:
		return []string{"xclip"}, []string{"xclip", "-o"}, nil
	case waylandClipMode:
		return []string{"wl-copy"}, []string{"wl-paste"}, nil
	case "off":
		return nil, nil, stock.NewBasicError("clipboard is turned off")
	}
	return nil, nil, stock.NewBasicError("unable to get clipboard command(s)")
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
		pipeTo("lb", value, false, "clear")
	}
}

func pipeTo(command, value string, wait bool, args ...string) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		stock.Die("unable to get stdin pipe", err)
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
		stock.Die("failed to run command", ran)
	}
}
