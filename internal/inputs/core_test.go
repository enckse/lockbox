package inputs_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/inputs"
)

func TestPlatformSet(t *testing.T) {
	if len(inputs.PlatformSet()) != 4 {
		t.Error("invalid platform set")
	}
}

func TestSet(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	inputs.EnvStore.Set("TEST")
	if inputs.EnvStore.Get() != "TEST" {
		t.Errorf("invalid set/get")
	}
}

func TestKeyValue(t *testing.T) {
	val := inputs.EnvStore.KeyValue("TEST")
	if val != "LOCKBOX_STORE=TEST" {
		t.Errorf("invalid keyvalue")
	}
}

func TestNewPlatform(t *testing.T) {
	for _, item := range inputs.PlatformSet() {
		os.Setenv("LOCKBOX_PLATFORM", item)
		s, err := inputs.NewPlatform()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if s != inputs.SystemPlatform(item) {
			t.Error("mismatch on input and resulting detection")
		}
	}
}

func TestNewPlatformUnknown(t *testing.T) {
	os.Setenv("LOCKBOX_PLATFORM", "afleaj")
	_, err := inputs.NewPlatform()
	if err == nil || err.Error() != "unknown platform mode" {
		t.Errorf("error expected for platform: %v", err)
	}
}

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
