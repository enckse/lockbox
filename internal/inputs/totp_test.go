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
