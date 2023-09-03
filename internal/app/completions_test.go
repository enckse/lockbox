package app_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestCompletions(t *testing.T) {
	testCompletion(t, true)
	testCompletion(t, false)
}

func testCompletion(t *testing.T, bash bool) {
	v, err := app.GenerateCompletions(bash, true, "lb")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(v) < 2 {
		t.Errorf("invalid result")
	}
	defer os.Clearenv()
	os.Setenv("LOCKBOX_COMPLETION_FUNCTION", "A")
	o, err := app.GenerateCompletions(bash, true, "lb")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(o) < 2 || len(o) != len(v) {
		t.Errorf("invalid result")
	}
	os.Setenv("LOCKBOX_COMPLETION_FUNCTION", "")
	v, err = app.GenerateCompletions(bash, false, "lb")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(v) != 1 {
		t.Errorf("invalid result")
	}
	os.Setenv("LOCKBOX_COMPLETION_FUNCTION", "ZZZ")
	_, err = app.GenerateCompletions(bash, false, "lb")
	if err == nil || err.Error() != "no profiles loaded, invalid environment setting?" {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_COMPLETION_FUNCTION", "NOCLIP-READONLY")
	n, err := app.GenerateCompletions(bash, false, "lb")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", n) == fmt.Sprintf("%v", v) {
		t.Errorf("invalid result, should filter")
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
