package config_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

var emptyRead = func(_ string) (io.Reader, error) {
	return nil, nil
}

func TestLoadIncludes(t *testing.T) {
	store.Clear()
	defer os.Clearenv()
	t.Setenv("TEST", "xyz")
	data := `include = ["$TEST/abc"]`
	r := strings.NewReader(data)
	if err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"$TEST/abc\"]"), nil
		} else {
			return nil, errors.New("invalid path")
		}
	}); err == nil || err.Error() != "too many nested includes (11 > 10)" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = ["abc"]`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"aaa\"]"), nil
		} else {
			return nil, errors.New("invalid path")
		}
	}); err == nil || err.Error() != "invalid path" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = 1`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"aaa\"]"), nil
		} else {
			return nil, errors.New("invalid path")
		}
	}); err == nil || err.Error() != "value is not of array type: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = [1]`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"aaa\"]"), nil
		} else {
			return nil, errors.New("invalid path")
		}
	}); err == nil || err.Error() != "value is not string in array: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = ["$TEST/abc"]
store="xyz"
`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("store = 'abc'"), nil
		} else {
			return nil, errors.New("invalid path")
		}
	}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 1 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetString("LOCKBOX_STORE")
	if val != "abc" || !ok {
		t.Errorf("invalid object: %v", val)
	}
}

func TestArrayLoad(t *testing.T) {
	store.Clear()
	defer os.Clearenv()
	t.Setenv("TEST", "abc")
	data := `store="xyz"
[clip]
copy_command = ["'xyz/$TEST'", "s", 1]
`
	r := strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err == nil || err.Error() != "value is not string in array: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
store="xyz"
[clip]
copy_command = ["'xyz/$TEST'", "s"]
`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 2 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetString("LOCKBOX_STORE")
	if val != "xyz" || !ok {
		t.Errorf("invalid object: %v", val)
	}
	a, ok := store.GetArray("LOCKBOX_CLIP_COPY_COMMAND")
	if fmt.Sprintf("%v", a) != "['xyz/abc' s]" || !ok {
		t.Errorf("invalid object: %v", a)
	}
	data = `include = []
store="xyz"
[clip]
copy_command = ["'xyz/$TEST'", "s"]
`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 2 {
		t.Errorf("invalid store")
	}
	val, ok = store.GetString("LOCKBOX_STORE")
	if val != "xyz" || !ok {
		t.Errorf("invalid object: %v", val)
	}
	a, ok = store.GetArray("LOCKBOX_CLIP_COPY_COMMAND")
	if fmt.Sprintf("%v", a) != "['xyz/abc' s]" || !ok {
		t.Errorf("invalid object: %v", val)
	}
}

func TestReadInt(t *testing.T) {
	store.Clear()
	data := `
[clip]
timeout = true
`
	r := strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err == nil || err.Error() != "non-int64 found where expected: true" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[clip]
timeout = 1
`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 1 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetInt64("LOCKBOX_CLIP_TIMEOUT")
	if val != 1 || !ok {
		t.Errorf("invalid object: %v", val)
	}
	data = `include = []
[clip]
timeout = -1
`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err == nil || err.Error() != "-1 is negative (not allowed here)" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReadBool(t *testing.T) {
	store.Clear()
	data := `
[totp]
enabled = 1
`
	r := strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err == nil || err.Error() != "non-bool found where expected: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[totp]
enabled = true
`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 1 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetBool("LOCKBOX_TOTP_ENABLED")
	if !val || !ok {
		t.Errorf("invalid object: %v", val)
	}
	data = `include = []
[totp]
enabled = false
`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 1 {
		t.Errorf("invalid store")
	}
	val, ok = store.GetBool("LOCKBOX_TOTP_ENABLED")
	if val || !ok {
		t.Errorf("invalid object: %v", val)
	}
}

func TestBadValues(t *testing.T) {
	store.Clear()
	data := `
[totsp]
enabled = "false"
`
	r := strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err == nil || err.Error() != "unknown key: totsp_enabled (LOCKBOX_TOTSP_ENABLED)" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[totp]
otp_format = -1
`
	r = strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err == nil || err.Error() != "non-string found where expected: -1" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestDefaultTOMLToLoadFile(t *testing.T) {
	store.Clear()
	os.Mkdir("testdata", 0o755)
	defer os.RemoveAll("testdata")
	file := filepath.Join("testdata", "config.toml")
	loaded, err := config.DefaultTOML()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.WriteFile(file, []byte(loaded), 0o644)
	if err := config.LoadConfigFile(file); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 25 {
		t.Errorf("invalid environment after load")
	}
}

func TestExpands(t *testing.T) {
	store.Clear()
	t.Setenv("TEST", "1")
	data := `include = []
store = "$TEST"
clip.copy_command = ["$TEST", "$TEST"]
[totp]
otp_format = "$TEST"
`
	r := strings.NewReader(data)
	if err := config.LoadConfig(r, emptyRead); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 3 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetString("LOCKBOX_TOTP_OTP_FORMAT")
	if val != "$TEST" || !ok {
		t.Errorf("invalid object: %v", val)
	}
	val, ok = store.GetString("LOCKBOX_STORE")
	if val != "1" || !ok {
		t.Errorf("invalid object: %v", val)
	}
	a, ok := store.GetArray("LOCKBOX_CLIP_COPY_COMMAND")
	if fmt.Sprintf("%v", a) != "[1 1]" || !ok {
		t.Errorf("invalid object: %v", a)
	}
}
