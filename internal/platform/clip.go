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
	Clipboard []string
	// ClipboardLoader handles how the system is detected
	ClipboardLoader interface {
		Name() (string, error)
		Runtime() string
	}
	// DefaultClipboardLoader is the default system detector
	DefaultClipboardLoader struct{}
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

// NewClipboard creates a new clipboard
func NewClipboard(loader ClipboardLoader) (Clipboard, error) {
	if !config.EnvFeatureClip.Get() {
		return Clipboard{}, config.NewFeatureError("clip")
	}
	overrideCopy := config.EnvClipCopy.Get()
	if len(overrideCopy) > 0 {
		return overrideCopy, nil
	}

	switch loader.Runtime() {
	case "darwin":
		return []string{"pbcopy"}, nil
	case "linux":
		name, err := loader.Name()
		if err != nil {
			return Clipboard{}, err
		}
		if strings.Contains(strings.ToLower(name), "microsoft") {
			return []string{"clip.exe"}, nil
		}
		if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) != "" {
			return []string{"wl-copy"}, nil
		}
		if strings.TrimSpace(os.Getenv("DISPLAY")) != "" {
			return []string{"xclip"}, nil
		}
		return Clipboard{}, errors.New("unable to detect linux clipboard")
	default:
		return Clipboard{}, errors.New("clipboard is unavailable")
	}
}

// CopyTo will copy to clipboard, if non-empty will clear later.
func (c Clipboard) CopyTo(value string) error {
	if len(c) == 0 {
		return errors.New("copy command is not set")
	}
	cmd := c[0]
	var args []string
	if len(c) > 1 {
		args = c[1:]
	}
	return pipeTo(cmd, value, args...)
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
