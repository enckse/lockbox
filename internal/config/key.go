// Package config handles user inputs/UI elements.
package config

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type (
	// KeyModeType are valid ways to get the key
	KeyModeType string
	// Key is a wrapper to help manage the returned key
	Key struct {
		mode     KeyModeType
		inputKey []string
		valid    bool
	}
)

const (
	plainKeyMode KeyModeType = "plaintext"
	noKeyMode    KeyModeType = "none"
	// IgnoreKeyMode will ignore the value set in the key (acts like no key)
	IgnoreKeyMode  KeyModeType = "ignore"
	commandKeyMode KeyModeType = "command"
	// DefaultKeyMode is the default operating keymode if NOT set
	DefaultKeyMode = commandKeyMode
)

// NewKey will create a new key
func NewKey(defaultKeyModeType KeyModeType) (Key, error) {
	keyMode := EnvPasswordMode.Get()
	if keyMode == "" {
		keyMode = string(defaultKeyModeType)
	}
	requireEmptyKey := false
	switch keyMode {
	case string(IgnoreKeyMode):
		return Key{mode: IgnoreKeyMode, inputKey: []string{}, valid: true}, nil
	case string(noKeyMode):
		requireEmptyKey = true
	case string(commandKeyMode), string(plainKeyMode):
	default:
		return Key{}, fmt.Errorf("unknown key mode: %s", keyMode)
	}
	useKey := envPassword.Get()
	isEmpty := len(useKey) == 0
	if !isEmpty {
		if strings.TrimSpace(useKey[0]) == "" {
			isEmpty = true
		}
	}
	if requireEmptyKey {
		if !isEmpty {
			return Key{}, errors.New("key can NOT be set in this key mode")
		}
	} else {
		if isEmpty {
			return Key{}, errors.New("key MUST be set in this key mode")
		}
	}
	return Key{mode: KeyModeType(keyMode), inputKey: useKey, valid: true}, nil
}

func (k Key) empty() bool {
	return k.valid && len(k.inputKey) == 0
}

// Read will read the key as configured by the mode
func (k Key) Read() (string, error) {
	if !k.valid {
		return "", errors.New("invalid key given")
	}
	if k.empty() {
		return "", nil
	}
	var useKey string
	if len(k.inputKey) > 0 {
		useKey = k.inputKey[0]
	}
	switch k.mode {
	case commandKeyMode:
		exe := k.inputKey[0]
		var args []string
		for idx, k := range k.inputKey {
			if idx == 0 {
				continue
			}
			args = append(args, k)
		}
		cmd := exec.Command(exe, args...)
		b, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("key command failed: %w", err)
		}
		useKey = string(b)
	}
	key := strings.TrimSpace(useKey)
	if key == "" {
		return "", errors.New("key is empty")
	}
	return key, nil
}
