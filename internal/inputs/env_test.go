package inputs_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/inputs"
)

func TestParseWindows(t *testing.T) {
	if _, err := inputs.ParseColorWindow(""); err.Error() != "invalid colorization rules for totp, none found" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",2"); err.Error() != "invalid colorization rule found: 2" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",1:200"); err.Error() != "invalid time found for colorization rule: 1:200" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",1:-1"); err.Error() != "invalid time found for colorization rule: 1:-1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",200:1"); err.Error() != "invalid time found for colorization rule: 200:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",-1:1"); err.Error() != "invalid time found for colorization rule: -1:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",2:1"); err.Error() != "invalid time found for colorization rule: 2:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",xxx:1"); err.Error() != "strconv.Atoi: parsing \"xxx\": invalid syntax" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",1:xxx"); err.Error() != "strconv.Atoi: parsing \"xxx\": invalid syntax" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := inputs.ParseColorWindow(",1:2,11:22"); err != nil {
		t.Errorf("invalid error: %v", err)
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

func checkYesNo(key string, t *testing.T, cb func() (bool, error), onEmpty bool) {
	os.Setenv(key, "yes")
	c, err := cb()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	os.Setenv(key, "")
	c, err = cb()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c != onEmpty {
		t.Error("invalid setting")
	}
	os.Setenv(key, "no")
	c, err = cb()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	os.Setenv(key, "lkaj;f")
	_, err = cb()
	if err == nil || err.Error() != fmt.Sprintf("invalid yes/no env value for %s", key) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestColorSetting(t *testing.T) {
	checkYesNo("LOCKBOX_NOCOLOR", t, inputs.IsNoColorEnabled, false)
}

func TestInteractiveSetting(t *testing.T) {
	checkYesNo("LOCKBOX_INTERACTIVE", t, inputs.IsInteractive, true)
}

func TestIsReadOnly(t *testing.T) {
	checkYesNo("LOCKBOX_READONLY", t, inputs.IsReadOnly, false)
}

func TestIsOSC52(t *testing.T) {
	checkYesNo("LOCKBOX_CLIP_OSC52", t, inputs.IsClipOSC52, false)
}

func TestIsNoTOTP(t *testing.T) {
	checkYesNo("LOCKBOX_NOTOTP", t, inputs.IsNoTOTP, false)
}

func TestIsNoClip(t *testing.T) {
	checkYesNo("LOCKBOX_NOCLIP", t, inputs.IsNoClipEnabled, false)
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
	if l != 20 {
		t.Errorf("invalid env count, outdated? %d", l)
	}
}

func TestReKey(t *testing.T) {
	os.Setenv("LOCKBOX_STORE_NEW", "")
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_KEYFILE_NEW", "")
	err := inputs.SetReKey()
	if err == nil || err.Error() != "missing required environment variables for rekey" {
		t.Errorf("failed: %v", err)
	}
	os.Setenv("LOCKBOX_STORE_NEW", "abc")
	err = inputs.SetReKey()
	if err == nil || err.Error() != "missing required environment variables for rekey" {
		t.Errorf("failed: %v", err)
	}
	if os.Getenv("LOCKBOX_STORE") != "abc" {
		t.Error("not set")
	}
	os.Setenv("LOCKBOX_KEY_NEW", "aaa")
	err = inputs.SetReKey()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if os.Getenv("LOCKBOX_KEY") != "aaa" && os.Getenv("LOCKBOX_KEYFILE") == "" {
		t.Error("not set")
	}
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_KEYFILE_NEW", "xxx")
	err = inputs.SetReKey()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if os.Getenv("LOCKBOX_KEYFILE") != "xxx" && os.Getenv("LOCKBOX_KEY") == "" {
		t.Error("not set")
	}
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_STORE_NEW", "")
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_KEYFILE_NEW", "")
}
