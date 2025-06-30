package platform

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"git.sr.ht/~enckse/lockbox/internal/app/commands"
)

type (
	// ClipboardDaemon is the manager interface
	ClipboardDaemon interface {
		WriteFile(string, string)
		ReadFile(string) ([]byte, error)
		Output(string, ...string) ([]byte, error)
		Start(string, ...string) error
		Getpid() int
		Copy(Clipboard, string)
		Sleep()
		Loader() ClipboardLoader
		Checkpid(int) error
	}
	// DefaultClipboardDaemon is the default functioning daemon
	DefaultClipboardDaemon struct{}
)

// WriteFile will write the necessary file to backing filesystem
func (d DefaultClipboardDaemon) WriteFile(file, data string) {
	os.WriteFile(file, []byte(data), 0o644)
}

// ReadFile will read a file from the filesystem
func (d DefaultClipboardDaemon) ReadFile(file string) ([]byte, error) {
	return os.ReadFile(file)
}

// Output will run a command and get output
func (d DefaultClipboardDaemon) Output(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}

// Start will start an disconnected execution
func (d DefaultClipboardDaemon) Start(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Start()
}

// Getpid will return the pid
func (d DefaultClipboardDaemon) Getpid() int {
	return os.Getpid()
}

// Copy will copy data to the clipboard
func (d DefaultClipboardDaemon) Copy(c Clipboard, val string) {
	c.CopyTo(val)
}

// Sleep will cause a pause/delay/wait
func (d DefaultClipboardDaemon) Sleep() {
	time.Sleep(1 * time.Second)
}

// Loader will get the backing loader to use
func (d DefaultClipboardDaemon) Loader() ClipboardLoader {
	return DefaultClipboardLoader{Full: true}
}

// Checkpid will check if a pid is still active
func (d DefaultClipboardDaemon) Checkpid(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.Signal(0))
}

// ClipboardManager handles the daemon runner
func ClipboardManager(daemon bool, manager ClipboardDaemon) error {
	if manager == nil {
		return errors.New("manager is nil")
	}
	clipboard, err := NewClipboard(manager.Loader())
	if err != nil {
		return err
	}
	if clipboard.PIDFile == "" {
		return errors.New("pidfile is unset")
	}
	getProcess := func() (string, error) {
		b, err := manager.ReadFile(clipboard.PIDFile)
		if err != nil {
			return "", err
		}
		val := strings.TrimSpace(string(b))
		return val, nil
	}
	if !daemon {
		if PathExists(clipboard.PIDFile) {
			p, err := getProcess()
			if err != nil {
				return err
			}
			pid, err := strconv.Atoi(p)
			if err != nil {
				return err
			}
			if err := manager.Checkpid(pid); err == nil {
				return nil
			}
		}
		return manager.Start(commands.Executable, commands.ClipManagerDaemon)
	}
	paste, pasteArgs, err := clipboard.Args(false)
	if err != nil {
		return err
	}
	pasteFxn := func() (string, error) {
		b, err := manager.Output(paste, pasteArgs...)
		if err != nil {
			return "", err
		}
		val := strings.TrimSpace(string(b))
		if val == "" {
			return "", nil
		}
		hash := sha256.New()
		if _, err := hash.Write([]byte(val)); err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", hash.Sum(nil)), nil
	}
	pid := strings.TrimSpace(fmt.Sprintf("%d", manager.Getpid()))
	isCurrentProcess := func() (bool, error) {
		val, err := getProcess()
		return val == pid, err
	}
	manager.WriteFile(clipboard.PIDFile, pid)
	var errs []error
	for {
		if len(errs) > 5 {
			return errors.Join(errs...)
		}
		manager.Sleep()
		ok, err := isCurrentProcess()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !ok {
			return nil
		}
		current, err := pasteFxn()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if current != "" {
			ok, err := wait(current, clipboard, manager, isCurrentProcess, pasteFxn)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if !ok {
				return nil
			}
		}
		errs = []error{}
	}
}

func wait(val string, clip Clipboard, mgr ClipboardDaemon, isCurrent func() (bool, error), pasteFxn func() (string, error)) (bool, error) {
	var count int64
	for count < clip.MaxTime {
		ok, err := isCurrent()
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		cur, err := pasteFxn()
		if err != nil {
			return false, err
		}
		if cur != val {
			return true, nil
		}
		mgr.Sleep()
		count++
	}
	mgr.Copy(clip, "")
	return true, nil
}
