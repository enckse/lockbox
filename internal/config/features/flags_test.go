package features_test

import (
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/config/features"
)

func TestCanTOTP(t *testing.T) {
	if !features.CanTOTP() {
		t.Error("flag should be on by default")
	}
}

func TestCanClip(t *testing.T) {
	if !features.CanClip() {
		t.Error("flag should be on by default")
	}
}

func TestNewError(t *testing.T) {
	err := features.NewError("abc")
	if err == nil || err.Error() != "abc feature is disabled" {
		t.Errorf("invalid error: %v", err)
	}
}
