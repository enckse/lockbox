package config_test

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/config/store"
)

func TestParse(t *testing.T) {
	store.Clear()
	defer store.Clear()
	os.Clearenv()
	defer os.Clearenv()
	mock := mockReader{}
	mock.home = func() (string, error) {
		return "", errors.New("invalid home")
	}
	mock.check = func(string) bool {
		return false
	}
	if err := config.Parse(mock); err == nil || err.Error() != "invalid home" {
		t.Errorf("invalid error: %v", err)
	}
	mock.home = func() (string, error) {
		return "hometest", nil
	}
	mock.read = func(p string) (io.Reader, error) {
		if p == "hometest/.config/lockbox/config.toml" {
			return strings.NewReader("[feature]\ncolor = true"), nil
		}
		return nil, nil
	}
	mock.check = func(string) bool {
		return true
	}
	store.SetBool("LOCKBOX_FEATURE_COLOR", false)
	if err := config.Parse(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if ok, _ := store.GetBool("LOCKBOX_FEATURE_COLOR"); !ok {
		t.Error("should have set")
	}
	t.Setenv("XDG_CONFIG_HOME", "xdghome")
	mock.read = func(p string) (io.Reader, error) {
		if p == "xdghome/lockbox/config.toml" {
			return strings.NewReader("[feature]\ncolor = false"), nil
		}
		return nil, nil
	}
	if err := config.Parse(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if ok, _ := store.GetBool("LOCKBOX_FEATURE_COLOR"); ok {
		t.Error("should have unset")
	}
	store.SetBool("LOCKBOX_FEATURE_COLOR", true)
	t.Setenv("LOCKBOX_CONFIG_TOML", "test")
	mock.check = func(p string) bool {
		return p != "test"
	}
	if err := config.Parse(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if ok, _ := store.GetBool("LOCKBOX_FEATURE_COLOR"); !ok {
		t.Error("should have set")
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
