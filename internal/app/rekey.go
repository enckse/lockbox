package app

import (
	"encoding/json"
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
		JSON() (map[string]backend.JSON, error)
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

// Show will get entry payload
func (r DefaultKeyer) Show(entry string) ([]byte, error) {
	return exec.Command(r.exe, cli.ShowCommand, entry).Output()
}

// JSON will get the JSON backing entries
func (r DefaultKeyer) JSON() (map[string]backend.JSON, error) {
	out, err := exec.Command(r.exe, cli.JSONCommand).Output()
	if err != nil {
		return nil, err
	}
	var j map[string]backend.JSON
	if err := json.Unmarshal(out, &j); err != nil {
		return nil, err
	}
	return j, nil
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
	entries, err := r.JSON()
	if err != nil {
		return err
	}
	for path, entry := range entries {
		if _, err := fmt.Fprintf(writer, "rekeying: %s\n", path); err != nil {
			return err
		}
		modTime := strings.TrimSpace(entry.ModTime)
		if modTime == "" {
			return errors.New("did not read modtime")
		}
		data, err := r.Show(path)
		if err != nil {
			return err
		}
		var insertEnv []string
		insertEnv = append(insertEnv, env...)
		insertEnv = append(insertEnv, fmt.Sprintf("%s=%s", inputs.ModTimeEnv, modTime))
		if err := r.Insert(ReKeyEntry{Path: path, Env: insertEnv, Data: data}); err != nil {
			return err
		}
	}
	return nil
}
