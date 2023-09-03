package app_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestBashCompletion(t *testing.T) {
	v, err := app.GenerateCompletions(true, false, "lb")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(v) != 1 {
		t.Errorf("invalid result")
	}
}

func TestZshCompletion(t *testing.T) {
	v, err := app.GenerateCompletions(false, false, "lb")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(v) != 1 {
		t.Errorf("invalid result")
	}
}

func TestProfileDisplay(t *testing.T) {
	p := app.Profile{Name: "_abc-test-awera-zzz"}
	if p.Display() != "TEST-AWERA-ZZZ" {
		t.Errorf("invalid display: %s", p.Display())
	}
}

func TestProfileEnv(t *testing.T) {
	p := app.Profile{Name: "_abc-test-awera-zzz"}
	if p.Env() != "LOCKBOX_COMPLETION_FUNCTION=TEST-AWERA-ZZZ" {
		t.Error("invalid env")
	}
}

func TestProfileOptions(t *testing.T) {
	p := app.Profile{Name: "_abc-test-awera-zzz"}
	p.CanClip = true
	p.CanTOTP = true
	if len(p.Options()) != 12 {
		t.Errorf("invalid options: %v", p.Options())
	}
	p.CanClip = false
	if len(p.Options()) != 11 {
		t.Errorf("invalid options: %v", p.Options())
	}
	p.CanClip = true
	p.CanTOTP = false
	if len(p.Options()) != 11 {
		t.Errorf("invalid options: %v", p.Options())
	}
	p.CanTOTP = true
	p.ReadOnly = true
	if len(p.Options()) != 8 {
		t.Errorf("invalid options: %v", p.Options())
	}
}

func TestProfileTOTPSubOptions(t *testing.T) {
	p := app.Profile{Name: "_abc-test-awera-zzz"}
	p.CanClip = true
	if len(p.TOTPSubCommands()) != 5 {
		t.Errorf("invalid options: %v", p.TOTPSubCommands())
	}
	p.CanClip = false
	if len(p.TOTPSubCommands()) != 4 {
		t.Errorf("invalid options: %v", p.TOTPSubCommands())
	}
	p.CanClip = true
	p.ReadOnly = true
	if len(p.TOTPSubCommands()) != 4 {
		t.Errorf("invalid options: %v", p.TOTPSubCommands())
	}
}
