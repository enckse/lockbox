// Package config handles user inputs/UI elements.
package config

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type (
	// Key is a wrapper to help manage the returned key
	Key struct {
		key []byte
	}
)

// Interactive indicates if the key requires interactive input
func (e *Key) Interactive() bool {
	return e.key == nil
}

// Key returns the key data
func (e *Key) Key() []byte {
	return e.key
}

// GetKey will get the encryption key setup for lb
func GetKey(dryrun bool) (*Key, error) {
	useKey := envKey.Get()
	keyMode := envKeyMode.Get()
	if keyMode == askKeyMode {
		isInteractive, err := EnvInteractive.Get()
		if err != nil {
			return nil, err
		}
		if !isInteractive {
			return nil, errors.New("ask key mode requested in non-interactive mode")
		}
		if useKey != "" {
			return nil, errors.New("key can NOT be set in ask key mode")
		}
		return &Key{}, nil
	}
	if useKey == "" {
		return nil, nil
	}
	if dryrun {
		return &Key{key: []byte{0}}, nil
	}
	var data []byte
	switch keyMode {
	case commandKeyMode:
		parts, err := shlex(useKey)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		b, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("key command failed: %w", err)
		}
		data = b
	case plainKeyMode:
		data = []byte(useKey)
	default:
		return nil, errors.New("unknown keymode")
	}
	b := []byte(strings.TrimSpace(string(data)))
	if len(b) == 0 {
		return nil, errors.New("key is empty")
	}
	return &Key{key: b}, nil
}
