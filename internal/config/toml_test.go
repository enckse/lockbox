package config_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/config"
)

func TestLoadIncludes(t *testing.T) {
	defer os.Clearenv()
	t.Setenv("TEST", "xyz")
	data := `include = ["$TEST/abc"]`
	r := strings.NewReader(data)
	if _, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"aaa\"]"), nil
		} else {
			return nil, errors.New("invalid path")
		}
	}); err == nil || err.Error() != "nested includes not allowed" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = ["abc"]`
	r = strings.NewReader(data)
	if _, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
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
	if _, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
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
	if _, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
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
	env, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("store = 'abc'"), nil
		} else {
			return nil, errors.New("invalid path")
		}
	})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(env) != 1 || env[0].Key != "LOCKBOX_STORE" || env[0].Value != "abc" {
		t.Errorf("invalid object: %v", env)
	}
}

func TestArrayLoad(t *testing.T) {
	defer os.Clearenv()
	t.Setenv("TEST", "abc")
	data := `store="xyz"
[clip]
copy = ["'xyz/$TEST'", "s", 1]
`
	r := strings.NewReader(data)
	_, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err == nil || err.Error() != "value is not string in array: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
store="xyz"
[clip]
copy = ["'xyz/$TEST'", "s"]
`
	r = strings.NewReader(data)
	env, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	slices.SortFunc(env, func(x, y config.ShellEnv) int {
		return strings.Compare(x.Key, y.Key)
	})
	if len(env) != 2 || env[1].Key != "LOCKBOX_STORE" || env[1].Value != "xyz" || env[0].Key != "LOCKBOX_CLIP_COPY" || env[0].Value != "'xyz/abc' s" {
		t.Errorf("invalid object: %v", env)
	}
	data = `include = []
store="xyz"
[clip]
copy = "'xyz/$TEST' s"
`
	r = strings.NewReader(data)
	env, err = config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	slices.SortFunc(env, func(x, y config.ShellEnv) int {
		return strings.Compare(x.Key, y.Key)
	})
	if len(env) != 2 || env[1].Key != "LOCKBOX_STORE" || env[1].Value != "xyz" || env[0].Key != "LOCKBOX_CLIP_COPY" || env[0].Value != "'xyz/abc' s" {
		t.Errorf("invalid object: %v", env)
	}
}

func TestRedirect(t *testing.T) {
	data := `include = []
[hook]
directory = "xyz"
`
	r := strings.NewReader(data)
	env, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(env) != 1 || env[0].Key != "LOCKBOX_HOOKDIR" || env[0].Value != "xyz" {
		t.Errorf("invalid object: %v", env)
	}
}

func TestReadInt(t *testing.T) {
	data := `
[clip]
max = true
`
	r := strings.NewReader(data)
	_, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err == nil || err.Error() != "non-int64 found where expected: true" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[clip]
max = 1
`
	r = strings.NewReader(data)
	env, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(env) != 1 || env[0].Key != "LOCKBOX_CLIP_MAX" || env[0].Value != "1" {
		t.Errorf("invalid object: %v", env)
	}
	data = `include = []
[clip]
max = -1
`
	r = strings.NewReader(data)
	env, err = config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err == nil || err.Error() != "-1 is negative (not allowed here)" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReadBool(t *testing.T) {
	defer os.Clearenv()
	t.Setenv("TEST", "abc")
	data := `
[totp]
enabled = 1
`
	r := strings.NewReader(data)
	_, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err == nil || err.Error() != "non-bool found where expected: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[totp]
enabled = true
`
	r = strings.NewReader(data)
	env, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(env) != 1 || env[0].Key != "LOCKBOX_NOTOTP" || env[0].Value != "yes" {
		t.Errorf("invalid object: %v", env)
	}
	data = `include = []
[totp]
enabled = false
`
	r = strings.NewReader(data)
	env, err = config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(env) != 1 || env[0].Key != "LOCKBOX_NOTOTP" || env[0].Value != "no" {
		t.Errorf("invalid object: %v", env)
	}
}

func TestBadValues(t *testing.T) {
	defer os.Clearenv()
	t.Setenv("TEST", "abc")
	data := `
[totsp]
enabled = "false"
`
	r := strings.NewReader(data)
	_, err := config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err == nil || err.Error() != "unknown key: totsp_enabled (LOCKBOX_TOTSP_ENABLED)" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[totp]
format = -1
`
	r = strings.NewReader(data)
	_, err = config.LoadConfig(r, func(p string) (io.Reader, error) {
		return nil, nil
	})
	if err == nil || err.Error() != "unknown field, can't determine type: totp_format (-1)" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestLoadFile(t *testing.T) {
	os.Mkdir("testdata", 0o755)
	// defer os.RemoveAll("testdata")
	defer os.Clearenv()
	file := filepath.Join("testdata", "config.toml")
	os.WriteFile(file, []byte(config.ExampleTOML), 0o644)
	if err := config.LoadConfigFile(file); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	count := 0
	for _, item := range os.Environ() {
		if strings.HasPrefix(item, "LOCKBOX_") {
			count++
		}
	}
	if count != 29 {
		t.Errorf("invalid environment after load: %d", count)
	}
}
