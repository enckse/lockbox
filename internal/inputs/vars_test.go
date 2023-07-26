package inputs_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/inputs"
)

func checkYesNo(key string, t *testing.T, obj inputs.EnvironmentBool, onEmpty bool) {
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
	checkYesNo("LOCKBOX_NOCOLOR", t, inputs.EnvNoColor, false)
}

func TestInteractiveSetting(t *testing.T) {
	checkYesNo("LOCKBOX_INTERACTIVE", t, inputs.EnvInteractive, true)
}

func TestIsReadOnly(t *testing.T) {
	checkYesNo("LOCKBOX_READONLY", t, inputs.EnvReadOnly, false)
}

func TestIsOSC52(t *testing.T) {
	checkYesNo("LOCKBOX_CLIP_OSC52", t, inputs.EnvClipOSC52, false)
}

func TestIsNoTOTP(t *testing.T) {
	checkYesNo("LOCKBOX_NOTOTP", t, inputs.EnvNoTOTP, false)
}

func TestIsNoClip(t *testing.T) {
	checkYesNo("LOCKBOX_NOCLIP", t, inputs.EnvNoClip, false)
}

func TestTOTP(t *testing.T) {
	os.Setenv("LOCKBOX_TOTP", "abc")
	if inputs.TOTPToken() != "abc" {
		t.Error("invalid totp token field")
	}
	os.Setenv("LOCKBOX_TOTP", "")
	if inputs.TOTPToken() != "totp" {
		t.Error("invalid totp token field")
	}
}

func TestGetKey(t *testing.T) {
	os.Setenv("LOCKBOX_KEY", "aaa")
	os.Setenv("LOCKBOX_KEYMODE", "lak;jfea")
	if _, err := inputs.GetKey(); err.Error() != "unknown keymode" {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_KEY", "")
	if _, err := inputs.GetKey(); err.Error() != "no key given" {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_KEY", "key")
	k, err := inputs.GetKey()
	if err != nil || string(k) != "key" {
		t.Error("invalid key retrieval")
	}
	os.Setenv("LOCKBOX_KEYMODE", "command")
	os.Setenv("LOCKBOX_KEY", "invalid command text is long and invalid via shlex")
	if _, err := inputs.GetKey(); err == nil {
		t.Error("should have failed")
	}
}

func TestListVariables(t *testing.T) {
	known := make(map[string]struct{})
	for _, v := range inputs.ListEnvironmentVariables(false) {
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
	if l != 22 {
		t.Errorf("invalid env count, outdated? %d", l)
	}
}

func TestReKey(t *testing.T) {
	_, err := inputs.GetReKey([]string{})
	if err == nil || err.Error() != "missing required arguments for rekey: LOCKBOX_KEY= LOCKBOX_KEYFILE= LOCKBOX_KEYMODE= LOCKBOX_STORE=" {
		t.Errorf("failed: %v", err)
	}
	_, err = inputs.GetReKey([]string{"-store", "abc"})
	if err == nil || err.Error() != "missing required arguments for rekey: LOCKBOX_KEY= LOCKBOX_KEYFILE= LOCKBOX_KEYMODE= LOCKBOX_STORE=abc" {
		t.Errorf("failed: %v", err)
	}
	out, err := inputs.GetReKey([]string{"-store", "abc", "-key", "aaa"})
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if fmt.Sprintf("%v", out) != "[LOCKBOX_KEY=aaa LOCKBOX_KEYFILE= LOCKBOX_KEYMODE= LOCKBOX_STORE=abc]" {
		t.Errorf("invalid env: %v", out)
	}
	out, err = inputs.GetReKey([]string{"-store", "abc", "-keyfile", "aaa"})
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

func TestGetClipboardMax(t *testing.T) {
	os.Setenv("LOCKBOX_CLIP_MAX", "")
	defer os.Clearenv()
	max, err := inputs.EnvClipboardMax.Get()
	if err != nil || max != 45 {
		t.Error("invalid clipboard read")
	}
	os.Setenv("LOCKBOX_CLIP_MAX", "1")
	max, err = inputs.EnvClipboardMax.Get()
	if err != nil || max != 1 {
		t.Error("invalid clipboard read")
	}
	os.Setenv("LOCKBOX_CLIP_MAX", "-1")
	if _, err := inputs.EnvClipboardMax.Get(); err == nil || err.Error() != "clipboard max time must be > 0" {
		t.Errorf("invalid err: %v", err)
	}
	os.Setenv("LOCKBOX_CLIP_MAX", "alk;ja")
	if _, err := inputs.EnvClipboardMax.Get(); err == nil || err.Error() != "strconv.Atoi: parsing \"alk;ja\": invalid syntax" {
		t.Errorf("invalid err: %v", err)
	}
	os.Setenv("LOCKBOX_CLIP_MAX", "0")
	if _, err := inputs.EnvClipboardMax.Get(); err == nil || err.Error() != "clipboard max time must be > 0" {
		t.Errorf("invalid err: %v", err)
	}
}

func TestGetHashLength(t *testing.T) {
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT_HASH_LENGTH", "")
	defer os.Clearenv()
	val, err := inputs.EnvHashLength.Get()
	if err != nil || val != 0 {
		t.Error("invalid hash read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT_HASH_LENGTH", "1")
	val, err = inputs.EnvHashLength.Get()
	if err != nil || val != 1 {
		t.Error("invalid hash read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT_HASH_LENGTH", "0")
	val, err = inputs.EnvHashLength.Get()
	if err != nil || val != 0 {
		t.Error("invalid hash read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT_HASH_LENGTH", "-1")
	if _, err := inputs.EnvHashLength.Get(); err == nil || err.Error() != "hash length must be >= 0" {
		t.Errorf("invalid err: %v", err)
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT_HASH_LENGTH", "-aoaofaij;p1")
	if _, err := inputs.EnvHashLength.Get(); err == nil || err.Error() != "strconv.Atoi: parsing \"-aoaofaij;p1\": invalid syntax" {
		t.Errorf("invalid err: %v", err)
	}
}

func TestFormatTOTP(t *testing.T) {
	otp := inputs.FormatTOTP("otpauth://abc")
	if otp != "otpauth://abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	otp = inputs.FormatTOTP("abc")
	if otp != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	os.Setenv("LOCKBOX_TOTP_FORMAT", "test/%s")
	otp = inputs.FormatTOTP("abc")
	if otp != "test/abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	os.Setenv("LOCKBOX_TOTP_FORMAT", "")
	otp = inputs.FormatTOTP("abc")
	if otp != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
}

func TestParseJSONMode(t *testing.T) {
	defer os.Clearenv()
	m, err := inputs.ParseJSONOutput()
	if m != inputs.JSONDataOutputHash || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT", "hAsH ")
	m, err = inputs.ParseJSONOutput()
	if m != inputs.JSONDataOutputHash || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT", "EMPTY")
	m, err = inputs.ParseJSONOutput()
	if m != inputs.JSONDataOutputBlank || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT", " PLAINtext ")
	m, err = inputs.ParseJSONOutput()
	if m != inputs.JSONDataOutputRaw || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT", "a")
	if _, err = inputs.ParseJSONOutput(); err == nil || err.Error() != "invalid JSON output mode: a" {
		t.Errorf("invalid error: %v", err)
	}
}
