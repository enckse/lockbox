package inputs

import (
	"os"
	"testing"
)

func TestColorSetting(t *testing.T) {
	os.Setenv("LOCKBOX_NOCOLOR", "yes")
	c, err := IsNoColorEnabled()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_NOCOLOR", "")
	c, err = IsNoColorEnabled()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_NOCOLOR", "no")
	c, err = IsNoColorEnabled()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_NOCOLOR", "lkaj;f")
	_, err = IsNoColorEnabled()
	if err == nil || err.Error() != "invalid yes/no env value for LOCKBOX_NOCOLOR" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInteractiveSetting(t *testing.T) {
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	c, err := IsInteractive()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "no")
	c, err = IsInteractive()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "")
	c, err = IsInteractive()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yaojia")
	_, err = IsInteractive()
	if err == nil || err.Error() != "invalid yes/no env value for LOCKBOX_INTERACTIVE" {
		t.Errorf("unexpected error: %v", err)
	}
}
