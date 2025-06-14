package config_test

import (
	"os"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

func TestNewEnvFiles(t *testing.T) {
	os.Clearenv()
	t.Setenv("LOCKBOX_CONFIG_TOML", "test")
	f := config.NewConfigFiles()
	if len(f) != 1 || f[0] != "test" {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("HOME", "test")
	t.Setenv("LOCKBOX_CONFIG_TOML", "")
	f = config.NewConfigFiles()
	if len(f) != 1 {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("LOCKBOX_CONFIG_TOML", "")
	t.Setenv("XDG_CONFIG_HOME", "test")
	f = config.NewConfigFiles()
	if len(f) != 2 {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("LOCKBOX_CONFIG_TOML", "")
	os.Unsetenv("HOME")
	f = config.NewConfigFiles()
	if len(f) != 1 {
		t.Errorf("invalid files: %v", f)
	}
}

func TestCanColor(t *testing.T) {
	store.Clear()
	defer store.Clear()
	if !config.CanColor() {
		t.Error("should be able to color")
	}
	store.SetBool("LOCKBOX_FEATURE_COLOR", false)
	if config.CanColor() {
		t.Error("should NOT be able to color")
	}
	store.Clear()
	t.Setenv("NO_COLOR", "1")
	if config.CanColor() {
		t.Error("should NOT be able to color")
	}
}

func TestTextFields(t *testing.T) {
	v := config.TextPositionFields()
	if v != "Text, Position.Start, Position.End" {
		t.Errorf("unexpected fields: %s", v)
	}
}

func TestNewFeatureError(t *testing.T) {
	err := config.NewFeatureError("abc")
	if err == nil || err.Error() != "abc feature is disabled" {
		t.Errorf("invalid error: %v", err)
	}
}
