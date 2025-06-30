package clip

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"git.sr.ht/~enckse/lockbox/internal/app/commands"
)

type (
	// Daemon is the manager interface
	Daemon interface {
		WriteFile(string, string)
		ReadFile(string) ([]byte, error)
		Output(string, ...string) ([]byte, error)
		Start(string, ...string) error
		Getpid() int
		Copy(Board, string)
		Sleep()
		Loader() Loader
	}
	// DefaultDaemon is the default functioning daemon
	DefaultDaemon struct{}
)

// WriteFile will write the necessary file to backing filesystem
func (d DefaultDaemon) WriteFile(file, data string) {
	os.WriteFile(file, []byte(data), 0o644)
}

// ReadFile will read a file from the filesystem
func (d DefaultDaemon) ReadFile(file string) ([]byte, error) {
	return os.ReadFile(file)
}

// Output will run a command and get output
func (d DefaultDaemon) Output(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}

// Start will start an disconnected execution
func (d DefaultDaemon) Start(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Start()
}

// Getpid will return the pid
func (d DefaultDaemon) Getpid() int {
	return os.Getpid()
}

// Copy will copy data to the clipboard
func (d DefaultDaemon) Copy(c Board, val string) {
	c.CopyTo(val)
}

// Sleep will cause a pause/delay/wait
func (d DefaultDaemon) Sleep() {
	time.Sleep(1 * time.Second)
}

// Loader will get the backing loader to use
func (d DefaultDaemon) Loader() Loader {
	return DefaultLoader{Full: true}
}

// Manager handles the daemon runner
func Manager(daemon bool, manager Daemon) error {
	if manager == nil {
		return errors.New("manager is nil")
	}
	clipboard, err := New(manager.Loader())
	if err != nil {
		return err
	}
	if clipboard.PIDFile == "" {
		return errors.New("pidfile is unset")
	}
	if !daemon {
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
		b, err := manager.ReadFile(clipboard.PIDFile)
		if err != nil {
			return false, err
		}
		val := strings.TrimSpace(string(b))
		return val == pid, nil
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

func wait(val string, clip Board, mgr Daemon, isCurrent func() (bool, error), pasteFxn func() (string, error)) (bool, error) {
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
