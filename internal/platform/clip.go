// Package platform handles platform-specific operations around clipboards.
package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/config"
)

type (
	// Clipboard represent system clipboard operations.
	Clipboard struct {
		PIDFile string
		copying []string
		pasting []string
		MaxTime int64
	}
	// ClipboardLoader handles how the system is detected
	ClipboardLoader interface {
		Name() (string, error)
		Runtime() string
		Complete() bool
	}
	// DefaultClipboardLoader is the default system detector
	DefaultClipboardLoader struct {
		Full bool
	}
)

// Name will get the uname results
func (l DefaultClipboardLoader) Name() (string, error) {
	b, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Runtime will return the GOOS runtime
func (l DefaultClipboardLoader) Runtime() string {
	return runtime.GOOS
}

// Complete indicates if the loader needs a full complete
func (l DefaultClipboardLoader) Complete() bool {
	return l.Full
}

func newBoard(copying, pasting []string) (Clipboard, error) {
	maximum, err := config.EnvClipTimeout.Get()
	if err != nil {
		return Clipboard{}, err
	}
	pid := config.EnvClipProcessFile.Get()
	return Clipboard{copying: copying, pasting: pasting, MaxTime: maximum, PIDFile: pid}, nil
}

// NewClipboard creates a new clipboard
func NewClipboard(loader ClipboardLoader) (Clipboard, error) {
	if !config.EnvFeatureClip.Get() {
		return Clipboard{}, config.NewFeatureError("clip")
	}
	overridePaste := config.EnvClipPaste.Get()
	overrideCopy := config.EnvClipCopy.Get()
	setPaste := len(overridePaste) > 0
	setCopy := len(overrideCopy) > 0
	if setPaste && setCopy {
		return newBoard(overrideCopy, overridePaste)
	}
	if setCopy && !loader.Complete() {
		return newBoard(overrideCopy, []string{})
	}

	var copying []string
	var pasting []string
	switch loader.Runtime() {
	case "darwin":
		copying = []string{"pbcopy"}
		pasting = []string{"pbpaste"}
	case "linux":
		name, err := loader.Name()
		if err != nil {
			return Clipboard{}, err
		}
		if strings.Contains(strings.ToLower(name), "microsoft") {
			copying = []string{"clip.exe"}
			pasting = []string{"powershell.exe", "-command", "Get-Clipboard"}
		} else {
			if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) != "" {
				copying = []string{"wl-copy"}
				pasting = []string{"wl-paste"}
			} else {
				if strings.TrimSpace(os.Getenv("DISPLAY")) != "" {
					copying = []string{"xclip"}
					pasting = []string{"xclip", "-o"}
				} else {
					return Clipboard{}, errors.New("unable to detect linux clipboard")
				}
			}
		}
	default:
		return Clipboard{}, errors.New("clipboard is unavailable")
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
func (c Clipboard) CopyTo(value string) error {
	cmd, args, err := c.Args(true)
	if err != nil {
		return err
	}
	pipeTo(cmd, value, args...)
	return nil
}

// Args returns clipboard args for execution.
func (c Clipboard) Args(copying bool) (string, []string, error) {
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
