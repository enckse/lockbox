package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/config"
)

func isSet(key string) bool {
	for _, item := range os.Environ() {
		if strings.HasPrefix(item, fmt.Sprintf("%s=", key)) {
			return true
		}
	}
	return false
}

func TestSet(t *testing.T) {
	os.Clearenv()
	config.EnvStore.Set("TEST")
	if config.EnvStore.Get() != "TEST" {
		t.Errorf("invalid set/get")
	}
	if !isSet("LOCKBOX_STORE") {
		t.Error("should be set")
	}
	config.EnvStore.Set("")
	if isSet("LOCKBOX_STORE") {
		t.Error("should be set")
	}
}

func TestKeyValue(t *testing.T) {
	val := config.EnvStore.KeyValue("TEST")
	if val != "LOCKBOX_STORE=TEST" {
		t.Errorf("invalid keyvalue")
	}
}

func TestNewEnvFiles(t *testing.T) {
	os.Clearenv()
	t.Setenv("LOCKBOX_CONFIG_TOML", "none")
	f := config.NewConfigFiles()
	if len(f) != 0 {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("LOCKBOX_CONFIG_TOML", "test")
	f = config.NewConfigFiles()
	if len(f) != 1 || f[0] != "test" {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("HOME", "test")
	t.Setenv("LOCKBOX_CONFIG_TOML", "detect")
	f = config.NewConfigFiles()
	if len(f) != 2 {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("LOCKBOX_CONFIG_TOML", "detect")
	t.Setenv("XDG_CONFIG_HOME", "test")
	f = config.NewConfigFiles()
	if len(f) != 4 {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("LOCKBOX_CONFIG_TOML", "detect")
	os.Unsetenv("HOME")
	f = config.NewConfigFiles()
	if len(f) != 2 {
		t.Errorf("invalid files: %v", f)
	}
}

func TestIsUnset(t *testing.T) {
	os.Clearenv()
	o, err := config.IsUnset("test", "   ")
	if err != nil || !o {
		t.Error("was unset")
	}
	o, err = config.IsUnset("test", "")
	if err != nil || !o {
		t.Error("was unset")
	}
	o, err = config.IsUnset("test", "a")
	if err != nil || o {
		t.Error("was set")
	}
	t.Setenv("UNSET_TEST", "abc")
	config.IsUnset("UNSET_TEST", "")
	if isSet("UNSET_TEST") {
		t.Error("found unset var")
	}
}

func TestEnviron(t *testing.T) {
	os.Clearenv()
	e := config.Environ()
	if len(e) != 0 {
		t.Error("invalid environ")
	}
	t.Setenv("LOCKBOX_STORE", "1")
	t.Setenv("LOCKBOX_2", "2")
	t.Setenv("LOCKBOX_CREDENTIALS_PASSWORD", "2")
	t.Setenv("LOCKBOX_ENV", "2")
	e = config.Environ()
	if len(e) != 2 || fmt.Sprintf("%v", e) != "[LOCKBOX_CREDENTIALS_PASSWORD=2 LOCKBOX_STORE=1]" {
		t.Errorf("invalid environ: %v", e)
	}
}

func TestWrap(t *testing.T) {
	w := config.Wrap(0, "")
	if w != "" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = config.Wrap(0, "abc\n\nabc\nxyz\n")
	if w != "abc\n\nabc xyz\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = config.Wrap(0, "abc\n\nabc\nxyz\n\nx")
	if w != "abc\n\nabc xyz\n\nx\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = config.Wrap(5, "abc\n\nabc\nxyz\n\nx")
	if w != "     abc\n\n     abc xyz\n\n     x\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
}

func TestCanColor(t *testing.T) {
	os.Clearenv()
	if can, _ := config.CanColor(); !can {
		t.Error("should be able to color")
	}
	for raw, expect := range map[string]bool{
		"INTERACTIVE":   true,
		"COLOR_ENABLED": true,
	} {
		os.Clearenv()
		key := fmt.Sprintf("LOCKBOX_%s", raw)
		t.Setenv(key, "true")
		if can, _ := config.CanColor(); can != expect {
			t.Errorf("expect != actual: %s", key)
		}
		t.Setenv(key, "false")
		if can, _ := config.CanColor(); can == expect {
			t.Errorf("expect == actual: %s", key)
		}
	}
	os.Clearenv()
	t.Setenv("NO_COLOR", "1")
	if can, _ := config.CanColor(); can {
		t.Error("should NOT be able to color")
	}
}
