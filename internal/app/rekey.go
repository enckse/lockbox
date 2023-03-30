package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/inputs"
)

type (
	// Keyer defines how rekeying happens
	Keyer interface {
		List() ([]string, error)
		Stats(string) ([]string, error)
		Show(string) ([]byte, error)
		Insert(ReKeyEntry) error
	}
	// ReKeyEntry is an entry that is being rekeyed
	ReKeyEntry struct {
		Path string
		Env  []string
		Data []byte
	}
	// DefaultKeyer is the default keyer for the application
	DefaultKeyer struct {
		exe string
	}
)

// NewDefaultKeyer initializes the default keyer
func NewDefaultKeyer() (DefaultKeyer, error) {
	exe, err := os.Executable()
	if err != nil {
		return DefaultKeyer{}, err
	}
	return DefaultKeyer{exe: exe}, nil
}

// List will get the list of keys in the store
func (r DefaultKeyer) List() ([]string, error) {
	return r.getCommandLines(cli.ListCommand)
}

// Stats will get stats for an entry
func (r DefaultKeyer) Stats(entry string) ([]string, error) {
	return r.getCommandLines(cli.StatsCommand, entry)
}

func (r DefaultKeyer) getCommandLines(args ...string) ([]string, error) {
	out, err := exec.Command(r.exe, args...).Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

// Show will get entry payload
func (r DefaultKeyer) Show(entry string) ([]byte, error) {
	return exec.Command(r.exe, cli.ShowCommand, entry).Output()
}

// Insert will insert the rekeying entry
func (r DefaultKeyer) Insert(entry ReKeyEntry) error {
	cmd := exec.Command(r.exe, cli.InsertCommand, entry.Path)
	cmd.Env = append(os.Environ(), entry.Env...)
	in, err := cmd.StdinPipe()
	if nil != err {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go func() {
		defer in.Close()
		in.Write(entry.Data)
	}()
	return cmd.Run()
}

// ReKey handles entry rekeying
func ReKey(writer io.Writer, r Keyer) error {
	env, err := inputs.GetReKey()
	if err != nil {
		return err
	}
	entries, err := r.List()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, err := fmt.Fprintf(writer, "rekeying: %s\n", entry); err != nil {
			return err
		}
		stats, err := r.Stats(entry)
		if err != nil {
			return fmt.Errorf("failed to get modtime, command failed: %w", err)
		}
		modTime := ""
		for _, stat := range stats {
			if strings.HasPrefix(stat, backend.ModTimeField) {
				if modTime != "" {
					return errors.New("unable to read modtime, too many values")
				}
				modTime = strings.TrimPrefix(stat, backend.ModTimeField)
			}
		}
		modTime = strings.TrimSpace(modTime)
		if modTime == "" {
			return errors.New("did not read modtime")
		}
		data, err := r.Show(entry)
		if err != nil {
			return err
		}
		var insertEnv []string
		insertEnv = append(insertEnv, env...)
		insertEnv = append(insertEnv, fmt.Sprintf("%s=%s", inputs.ModTimeEnv, modTime))
		if err := r.Insert(ReKeyEntry{Path: entry, Env: insertEnv, Data: data}); err != nil {
			return err
		}
	}
	return nil
}
