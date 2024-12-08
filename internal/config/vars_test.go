package config_test

import (
	"fmt"
	"testing"

	"github.com/seanenck/lockbox/internal/config"
	"github.com/seanenck/lockbox/internal/config/store"
)

func checkYesNo(key string, t *testing.T, obj config.EnvironmentBool, onEmpty bool) {
	store.Clear()
	if obj.Get() != onEmpty {
		t.Error("invalid setting")
	}
	store.SetBool(key, true)
	if !obj.Get() {
		t.Error("invalid setting")
	}
	store.SetBool(key, false)
	if obj.Get() {
		t.Error("invalid setting")
	}
}

func TestColorSetting(t *testing.T) {
	checkYesNo("LOCKBOX_COLOR_ENABLED", t, config.EnvColorEnabled, true)
}

func TestNoHook(t *testing.T) {
	checkYesNo("LOCKBOX_HOOKS_ENABLED", t, config.EnvHooksEnabled, true)
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
	checkYesNo("LOCKBOX_TOTP_ENABLED", t, config.EnvTOTPEnabled, true)
}

func TestIsNoClip(t *testing.T) {
	checkYesNo("LOCKBOX_CLIP_ENABLED", t, config.EnvClipEnabled, true)
}

func TestIsNoGeneratePassword(t *testing.T) {
	checkYesNo("LOCKBOX_PWGEN_ENABLED", t, config.EnvPasswordGenEnabled, true)
}

func TestIsTitle(t *testing.T) {
	checkYesNo("LOCKBOX_PWGEN_TITLE", t, config.EnvPasswordGenTitle, true)
}

func TestTOTP(t *testing.T) {
	store.Clear()
	if config.EnvTOTPEntry.Get() != "totp" {
		t.Error("invalid totp token field")
	}
	store.SetString("LOCKBOX_TOTP_ENTRY", "abc")
	if config.EnvTOTPEntry.Get() != "abc" {
		t.Error("invalid totp token field")
	}
}

func TestFormatTOTP(t *testing.T) {
	store.Clear()
	otp := config.EnvTOTPFormat.Get("otpauth://abc")
	if otp != "otpauth://abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	otp = config.EnvTOTPFormat.Get("abc")
	if otp != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	otp = config.EnvTOTPFormat.Get("abc")
	if otp != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	store.SetString("LOCKBOX_TOTP_OTP_FORMAT", "test/%s")
	otp = config.EnvTOTPFormat.Get("abc")
	if otp != "test/abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
}

func TestClipboardMax(t *testing.T) {
	checkInt(config.EnvClipTimeout, "LOCKBOX_CLIP_TIMEOUT", "clipboard max time", 45, false, t)
}

func TestHashLength(t *testing.T) {
	checkInt(config.EnvJSONHashLength, "LOCKBOX_JSON_HASH_LENGTH", "hash length", 0, true, t)
}

func TestMaxTOTP(t *testing.T) {
	checkInt(config.EnvTOTPTimeout, "LOCKBOX_TOTP_TIMEOUT", "max totp time", 120, false, t)
}

func TestWordCount(t *testing.T) {
	checkInt(config.EnvPasswordGenWordCount, "LOCKBOX_PWGEN_WORD_COUNT", "word count", 8, false, t)
}

func checkInt(e config.EnvironmentInt, key, text string, def int64, allowZero bool, t *testing.T) {
	store.Clear()
	val, err := e.Get()
	if err != nil || val != def {
		t.Error("invalid read")
	}
	store.SetInt64(key, 1)
	val, err = e.Get()
	if err != nil || val != 1 {
		t.Error("invalid read")
	}
	store.SetInt64(key, -1)
	zero := ""
	if allowZero {
		zero = "="
	}
	if _, err := e.Get(); err == nil || err.Error() != fmt.Sprintf("%s must be >%s 0", text, zero) {
		t.Errorf("invalid err: %v", err)
	}
	store.SetInt64(key, 0)
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
