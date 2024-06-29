package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/config"
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
	os.Setenv("LOCKBOX_STORE", "1")
	os.Setenv("LOCKBOX_2", "2")
	os.Setenv("LOCKBOX_KEY", "2")
	os.Setenv("LOCKBOX_ENV", "2")
	e = config.Environ()
	if len(e) != 2 || fmt.Sprintf("%v", e) != "[LOCKBOX_KEY=2 LOCKBOX_STORE=1]" {
		t.Errorf("invalid environ: %v", e)
	}
}

func TestExpandParsed(t *testing.T) {
	os.Clearenv()
	os.Setenv("TEST_ABC", "1")
	os.Setenv("LOCKBOX_ENV_EXPANDS", "a")
	_, err := config.ExpandParsed(nil)
	if err == nil || err.Error() != "invalid input variables" {
		t.Errorf("invalid error: %v", err)
	}
	r, err := config.ExpandParsed(make(map[string]string))
	if err != nil || len(r) != 0 {
		t.Errorf("invalid expand")
	}
	os.Setenv("LOCKBOX_ENV_EXPANDS", "a")
	ins := make(map[string]string)
	ins["TEST"] = "$TEST_ABC"
	_, err = config.ExpandParsed(ins)
	if err == nil || err.Error() != "strconv.Atoi: parsing \"a\": invalid syntax" {
		t.Errorf("invalid error: %v", err)
	}
	ins["LOCKBOX_ENV_EXPANDS"] = "2"
	r, err = config.ExpandParsed(ins)
	if err != nil || len(r) != 2 || r["TEST"] != "1" {
		t.Errorf("invalid expand: %v", r)
	}
	delete(ins, "LOCKBOX_ENV_EXPANDS")
	os.Setenv("LOCKBOX_ENV_EXPANDS", "2")
	r, err = config.ExpandParsed(ins)
	if err != nil || len(r) != 1 || r["TEST"] != "1" {
		t.Errorf("invalid expand: %v", r)
	}
	os.Setenv("LOCKBOX_ENV_EXPANDS", "2")
	r, err = config.ExpandParsed(ins)
	if err != nil || len(r) != 1 || r["TEST"] != "1" {
		t.Errorf("invalid expand: %v", r)
	}
	os.Setenv("TEST_ABC", "$OTHER_TEST")
	os.Setenv("OTHER_TEST", "$ANOTHER_TEST")
	os.Setenv("ANOTHER_TEST", "2")
	os.Setenv("LOCKBOX_ENV_EXPANDS", "1")
	if _, err = config.ExpandParsed(ins); err == nil || err.Error() != "reached maximum expand cycle count" {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_ENV_EXPANDS", "2")
	ins["OTHER_FIRST"] = "2"
	ins["OTHER_OTHER"] = "$ANOTHER_TEST|$TEST_ABC|$OTHER_TEST"
	ins["OTHER"] = "$OTHER_OTHER|$OTHER_FIRST"
	os.Setenv("LOCKBOX_ENV_EXPANDS", "20")
	r, err = config.ExpandParsed(ins)
	if err != nil || len(r) != 4 || r["TEST"] != "2" || r["OTHER"] != "2|2|2|2" || r["OTHER_OTHER"] != "2|2|2" {
		t.Errorf("invalid expand: %v", r)
	}
	os.Setenv("LOCKBOX_ENV_EXPANDS", "0")
	delete(ins, "OTHER_FIRST")
	delete(ins, "OTHER")
	delete(ins, "OTHER_OTHER")
	r, err = config.ExpandParsed(ins)
	if err != nil || len(r) != 1 || r["TEST"] != "$TEST_ABC" {
		t.Errorf("invalid expand: %v", r)
	}
	os.Unsetenv("LOCKBOX_ENV_EXPANDS")
	delete(ins, "OTHER_FIRST")
	delete(ins, "OTHER")
	delete(ins, "OTHER_OTHER")
	ins["LOCKBOX_ENV_EXPANDS"] = "0"
	r, err = config.ExpandParsed(ins)
	if err != nil || len(r) != 2 || r["TEST"] != "$TEST_ABC" {
		t.Errorf("invalid expand: %v", r)
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
