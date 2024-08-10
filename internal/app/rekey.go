package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/seanenck/lockbox/internal/backend"
	"github.com/seanenck/lockbox/internal/config"
)

type (
	// Keyer defines how rekeying happens
	Keyer interface {
		JSON() (map[string]backend.JSON, error)
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

// JSON will get the JSON backing entries
func (r DefaultKeyer) JSON() (map[string]backend.JSON, error) {
	out, err := exec.Command(r.exe, JSONCommand).Output()
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
	cmd := exec.Command(r.exe, InsertCommand, entry.Path)
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
func ReKey(cmd CommandOptions, r Keyer) error {
	args := cmd.Args()
	vars, err := config.GetReKey(args)
	if err != nil {
		return err
	}
	if !cmd.Confirm("proceed with rekey") {
		return nil
	}
	if err := config.EnvJSONDataOutput.Set(string(config.JSONDataOutputRaw)); err != nil {
		return err
	}
	entries, err := r.JSON()
	if err != nil {
		return err
	}
	writer := cmd.Writer()
	for path, entry := range entries {
		if _, err := fmt.Fprintf(writer, "rekeying: %s\n", path); err != nil {
			return err
		}
		var modTime string
		if vars.ModMode != config.ReKeyModModeNone {
			modTime = strings.TrimSpace(entry.ModTime)
			if modTime == "" {
				switch vars.ModMode {
				case config.ReKeyModModeSkip:
				case config.ReKeyModModeError:
					return errors.New("did not read modtime")
				default:
					return errors.New("unknown modtime control")
				}
			}
		}

		var insertEnv []string
		insertEnv = append(insertEnv, vars.Env...)
		if modTime != "" {
			insertEnv = append(insertEnv, config.EnvModTime.KeyValue(modTime))
		}
		if err := r.Insert(ReKeyEntry{Path: path, Env: insertEnv, Data: []byte(entry.Data)}); err != nil {
			return err
		}
	}
	return nil
}
