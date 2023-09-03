package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/config"
)

func checkYesNo(key string, t *testing.T, obj config.EnvironmentBool, onEmpty bool) {
	os.Setenv(key, "yes")
	c, err := obj.Get()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	os.Setenv(key, "")
	c, err = obj.Get()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c != onEmpty {
		t.Error("invalid setting")
	}
	os.Setenv(key, "no")
	c, err = obj.Get()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	os.Setenv(key, "afoieae")
	_, err = obj.Get()
	if err == nil || err.Error() != fmt.Sprintf("invalid yes/no env value for %s", key) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestColorSetting(t *testing.T) {
	checkYesNo("LOCKBOX_NOCOLOR", t, config.EnvNoColor, false)
}

func TestInteractiveSetting(t *testing.T) {
	checkYesNo("LOCKBOX_INTERACTIVE", t, config.EnvInteractive, true)
}

func TestIsReadOnly(t *testing.T) {
	checkYesNo("LOCKBOX_READONLY", t, config.EnvReadOnly, false)
}

func TestIsOSC52(t *testing.T) {
	checkYesNo("LOCKBOX_CLIP_OSC52", t, config.EnvClipOSC52, false)
}

func TestIsNoTOTP(t *testing.T) {
	checkYesNo("LOCKBOX_NOTOTP", t, config.EnvNoTOTP, false)
}

func TestIsNoClip(t *testing.T) {
	checkYesNo("LOCKBOX_NOCLIP", t, config.EnvNoClip, false)
}

func TestTOTP(t *testing.T) {
	os.Setenv("LOCKBOX_TOTP", "abc")
	if config.EnvTOTPToken.Get() != "abc" {
		t.Error("invalid totp token field")
	}
	os.Setenv("LOCKBOX_TOTP", "")
	if config.EnvTOTPToken.Get() != "totp" {
		t.Error("invalid totp token field")
	}
}

func TestGetKey(t *testing.T) {
	os.Setenv("LOCKBOX_KEY", "aaa")
	os.Setenv("LOCKBOX_KEYMODE", "lak;jfea")
	if k, err := config.GetKey(false); err.Error() != "unknown keymode" || k != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_KEY", "")
	if k, err := config.GetKey(false); err != nil || k != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_KEY", "key")
	k, err := config.GetKey(false)
	if err != nil || k == nil || string(k.Key()) != "key" || k.Interactive() {
		t.Error("invalid key retrieval")
	}
	os.Setenv("LOCKBOX_KEY", "key")
	k, err = config.GetKey(true)
	if err != nil || k == nil || len(k.Key()) != 1 || k.Key()[0] != 0 || k.Interactive() {
		t.Error("invalid key retrieval")
	}
	os.Setenv("LOCKBOX_KEYMODE", "command")
	os.Setenv("LOCKBOX_KEY", "invalid command text is long and invalid via shlex")
	if k, err := config.GetKey(false); err == nil || k != nil {
		t.Error("should have failed")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_KEYMODE", "interactive")
	os.Setenv("LOCKBOX_KEY", "")
	if k, err := config.GetKey(false); err != nil || k == nil || !k.Interactive() {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_KEYMODE", "interactive")
	os.Setenv("LOCKBOX_KEY", "")
	if k, err := config.GetKey(true); err != nil || k == nil || !k.Interactive() {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "no")
	os.Setenv("LOCKBOX_KEYMODE", "interactive")
	os.Setenv("LOCKBOX_KEY", "")
	if k, err := config.GetKey(false); err == nil || err.Error() != "interactive key mode requested in non-interactive mode" || k != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_KEYMODE", "interactive")
	os.Setenv("LOCKBOX_KEY", "aaa")
	if k, err := config.GetKey(false); err == nil || err.Error() != "key can NOT be set in interactive mode" || k != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestListVariables(t *testing.T) {
	known := make(map[string]struct{})
	for _, v := range config.ListEnvironmentVariables() {
		trim := strings.Split(strings.TrimSpace(v), " ")[0]
		if !strings.HasPrefix(trim, "LOCKBOX_") {
			t.Errorf("invalid env: %s", v)
		}
		if _, ok := known[trim]; ok {
			t.Errorf("invalid re-used env: %s", trim)
		}
		known[trim] = struct{}{}
	}
	l := len(known)
	if l != 24 {
		t.Errorf("invalid env count, outdated? %d", l)
	}
}

func TestReKey(t *testing.T) {
	_, err := config.GetReKey([]string{})
	if err == nil || !strings.HasPrefix(err.Error(), "missing required arguments for rekey") {
		t.Errorf("failed: %v", err)
	}
	_, err = config.GetReKey([]string{"-store", "abc"})
	if err == nil || !strings.HasPrefix(err.Error(), "missing required arguments for rekey") {
		t.Errorf("failed: %v", err)
	}
	out, err := config.GetReKey([]string{"-store", "abc", "-key", "aaa"})
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if fmt.Sprintf("%v", out) != "[LOCKBOX_KEY=aaa LOCKBOX_KEYFILE= LOCKBOX_KEYMODE= LOCKBOX_STORE=abc]" {
		t.Errorf("invalid env: %v", out)
	}
	out, err = config.GetReKey([]string{"-store", "abc", "-keyfile", "aaa"})
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if fmt.Sprintf("%v", out) != "[LOCKBOX_KEY= LOCKBOX_KEYFILE=aaa LOCKBOX_KEYMODE= LOCKBOX_STORE=abc]" {
		t.Errorf("invalid env: %v", out)
	}
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_STORE_NEW", "")
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_KEYFILE_NEW", "")
}

func TestFormatTOTP(t *testing.T) {
	otp := config.EnvFormatTOTP.Get("otpauth://abc")
	if otp != "otpauth://abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	otp = config.EnvFormatTOTP.Get("abc")
	if otp != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	os.Setenv("LOCKBOX_TOTP_FORMAT", "test/%s")
	otp = config.EnvFormatTOTP.Get("abc")
	if otp != "test/abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	os.Setenv("LOCKBOX_TOTP_FORMAT", "")
	otp = config.EnvFormatTOTP.Get("abc")
	if otp != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
}

func TestParseJSONMode(t *testing.T) {
	defer os.Clearenv()
	m, err := config.ParseJSONOutput()
	if m != config.JSONDataOutputHash || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA", "hAsH ")
	m, err = config.ParseJSONOutput()
	if m != config.JSONDataOutputHash || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA", "EMPTY")
	m, err = config.ParseJSONOutput()
	if m != config.JSONDataOutputBlank || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA", " PLAINtext ")
	m, err = config.ParseJSONOutput()
	if m != config.JSONDataOutputRaw || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA", "a")
	if _, err = config.ParseJSONOutput(); err == nil || err.Error() != "invalid JSON output mode: a" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestClipboardMax(t *testing.T) {
	checkInt(config.EnvClipMax, "LOCKBOX_CLIP_MAX", "clipboard max time", 45, false, t)
}

func TestHashLength(t *testing.T) {
	checkInt(config.EnvHashLength, "LOCKBOX_JSON_DATA_HASH_LENGTH", "hash length", 0, true, t)
}

func TestMaxTOTP(t *testing.T) {
	checkInt(config.EnvMaxTOTP, "LOCKBOX_TOTP_MAX", "max totp time", 120, false, t)
}

func checkInt(e config.EnvironmentInt, key, text string, def int, allowZero bool, t *testing.T) {
	os.Setenv(key, "")
	defer os.Clearenv()
	val, err := e.Get()
	if err != nil || val != def {
		t.Error("invalid read")
	}
	os.Setenv(key, "1")
	val, err = e.Get()
	if err != nil || val != 1 {
		t.Error("invalid read")
	}
	os.Setenv(key, "-1")
	zero := ""
	if allowZero {
		zero = "="
	}
	if _, err := e.Get(); err == nil || err.Error() != fmt.Sprintf("%s must be >%s 0", text, zero) {
		t.Errorf("invalid err: %v", err)
	}
	os.Setenv(key, "alk;ja")
	if _, err := e.Get(); err == nil || err.Error() != "strconv.Atoi: parsing \"alk;ja\": invalid syntax" {
		t.Errorf("invalid err: %v", err)
	}
	os.Setenv(key, "0")
	if allowZero {
		val, err = e.Get()
		if err != nil || val != 0 {
			t.Error("invalid read")
		}
	} else {
		if _, err := e.Get(); err == nil || err.Error() != fmt.Sprintf("%s must be > 0", text) {
			t.Errorf("invalid err: %v", err)
		}
	}
}

func TestEnvironDefinitions(t *testing.T) {
	b, err := os.ReadFile("vars.go")
	if err != nil {
		t.Errorf("invalid err: %v", err)
	}
	count := 0
	inVars := false
	for _, line := range strings.Split(string(b), "\n") {
		if line == "var (" {
			inVars = true
			continue
		}
		if inVars {
			if strings.Contains(line, "= Environment") {
				count++
			} else {
				if line == ")" {
					inVars = false
					break
				}
			}
		}
	}
	if count == 0 || inVars {
		t.Errorf("invalid simple parse: %d", count)
	}
	os.Clearenv()
	vals := config.ListEnvironmentVariables()
	if len(vals) != count {
		t.Errorf("invalid environment variable info: %d != %d", count, len(vals))
	}
	os.Clearenv()
	expect := make(map[string]struct{})
	for _, val := range vals {
		env := strings.Split(strings.TrimSpace(val), "\n")[0]
		if !strings.HasPrefix(env, "LOCKBOX_") {
			t.Errorf("invalid env var: %s", env)
		}
		if env == "LOCKBOX_ENV" {
			continue
		}
		os.Setenv(env, "test")
		expect[env] = struct{}{}
	}
	read := config.Environ()
	if len(read) != len(expect) {
		t.Errorf("invalid environment variable info: %d != %d", len(expect), len(read))
	}
	for k := range expect {
		found := false
		for _, r := range read {
			if r == fmt.Sprintf("%s=test", k) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unable to find env: %s", k)
		}
	}
}
