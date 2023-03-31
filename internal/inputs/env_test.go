package inputs_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/inputs"
)

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
	_, err := inputs.GetReKey()
	if err == nil || err.Error() != "missing required environment variables for rekey: LOCKBOX_KEYMODE= LOCKBOX_KEY= LOCKBOX_KEYFILE= LOCKBOX_STORE=" {
		t.Errorf("failed: %v", err)
	}
	os.Setenv("LOCKBOX_STORE_NEW", "abc")
	_, err = inputs.GetReKey()
	if err == nil || err.Error() != "missing required environment variables for rekey: LOCKBOX_KEYMODE= LOCKBOX_KEY= LOCKBOX_KEYFILE= LOCKBOX_STORE=abc" {
		t.Errorf("failed: %v", err)
	}
	os.Setenv("LOCKBOX_KEY_NEW", "aaa")
	out, err := inputs.GetReKey()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if fmt.Sprintf("%v", out) != "[LOCKBOX_KEYMODE= LOCKBOX_KEY=aaa LOCKBOX_KEYFILE= LOCKBOX_STORE=abc]" {
		t.Errorf("invalid env: %v", out)
	}
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_KEYFILE_NEW", "xxx")
	out, err = inputs.GetReKey()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if fmt.Sprintf("%v", out) != "[LOCKBOX_KEYMODE= LOCKBOX_KEY= LOCKBOX_KEYFILE=xxx LOCKBOX_STORE=abc]" {
		t.Errorf("invalid env: %v", out)
	}
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_STORE_NEW", "")
	os.Setenv("LOCKBOX_KEY_NEW", "")
	os.Setenv("LOCKBOX_KEYFILE_NEW", "")
}
