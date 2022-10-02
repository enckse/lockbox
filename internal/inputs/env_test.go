package inputs_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/inputs"
)

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

func TestColorSetting(t *testing.T) {
	os.Setenv("LOCKBOX_NOCOLOR", "yes")
	c, err := inputs.IsNoColorEnabled()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_NOCOLOR", "")
	c, err = inputs.IsNoColorEnabled()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_NOCOLOR", "no")
	c, err = inputs.IsNoColorEnabled()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_NOCOLOR", "lkaj;f")
	_, err = inputs.IsNoColorEnabled()
	if err == nil || err.Error() != "invalid yes/no env value for LOCKBOX_NOCOLOR" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInteractiveSetting(t *testing.T) {
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	c, err := inputs.IsInteractive()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "no")
	c, err = inputs.IsInteractive()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "")
	c, err = inputs.IsInteractive()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yaojia")
	_, err = inputs.IsInteractive()
	if err == nil || err.Error() != "invalid yes/no env value for LOCKBOX_INTERACTIVE" {
		t.Errorf("unexpected error: %v", err)
	}
}
