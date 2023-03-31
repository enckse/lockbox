package inputs_test

import (
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
