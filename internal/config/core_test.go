package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/config"
)

func TestPlatformSet(t *testing.T) {
	if len(config.Platforms) != 4 {
		t.Error("invalid platform set")
	}
}

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
	defer os.Clearenv()
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

func TestNewPlatform(t *testing.T) {
	for _, item := range config.Platforms {
		os.Setenv("LOCKBOX_PLATFORM", item)
		s, err := config.NewPlatform()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if s != config.SystemPlatform(item) {
			t.Error("mismatch on input and resulting detection")
		}
	}
}

func TestNewPlatformUnknown(t *testing.T) {
	os.Setenv("LOCKBOX_PLATFORM", "afleaj")
	_, err := config.NewPlatform()
	if err == nil || err.Error() != "unknown platform mode" {
		t.Errorf("error expected for platform: %v", err)
	}
}

func TestParseWindows(t *testing.T) {
	if _, err := config.ParseColorWindow(""); err.Error() != "invalid colorization rules for totp, none found" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",2"); err.Error() != "invalid colorization rule found: 2" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",1:200"); err.Error() != "invalid time found for colorization rule: 1:200" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",1:-1"); err.Error() != "invalid time found for colorization rule: 1:-1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",200:1"); err.Error() != "invalid time found for colorization rule: 200:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",-1:1"); err.Error() != "invalid time found for colorization rule: -1:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",2:1"); err.Error() != "invalid time found for colorization rule: 2:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",xxx:1"); err.Error() != "strconv.Atoi: parsing \"xxx\": invalid syntax" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",1:xxx"); err.Error() != "strconv.Atoi: parsing \"xxx\": invalid syntax" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := config.ParseColorWindow(",1:2,11:22"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestNewEnvFiles(t *testing.T) {
	os.Setenv("LOCKBOX_ENV", "none")
	f, err := config.NewEnvFiles()
	if len(f) != 0 || err != nil {
		t.Errorf("invalid files: %v %v", f, err)
	}
	os.Setenv("LOCKBOX_ENV", "test")
	f, err = config.NewEnvFiles()
	if len(f) != 1 || f[0] != "test" || err != nil {
		t.Errorf("invalid files: %v %v", f, err)
	}
	os.Setenv("HOME", "test")
	os.Setenv("LOCKBOX_ENV", "detect")
	f, err = config.NewEnvFiles()
	if len(f) != 2 || err != nil {
		t.Errorf("invalid files: %v %v", f, err)
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
	os.Setenv("UNSET_TEST", "abc")
	defer os.Clearenv()
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
	os.Setenv("LOCKBOX_1", "1")
	os.Setenv("LOCKBOX_2", "2")
	e = config.Environ()
	if len(e) != 2 || fmt.Sprintf("%v", e) != "[LOCKBOX_1=1 LOCKBOX_2=2]" {
		t.Errorf("invalid environ: %v", e)
	}
}
