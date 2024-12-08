package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/seanenck/lockbox/internal/config"
	"github.com/seanenck/lockbox/internal/config/store"
)

func TestNewEnvFiles(t *testing.T) {
	os.Clearenv()
	t.Setenv("LOCKBOX_CONFIG_TOML", "test")
	f := config.NewConfigFiles()
	if len(f) != 1 || f[0] != "test" {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("HOME", "test")
	t.Setenv("LOCKBOX_CONFIG_TOML", "detect")
	f = config.NewConfigFiles()
	if len(f) != 2 {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("LOCKBOX_CONFIG_TOML", "detect")
	t.Setenv("XDG_CONFIG_HOME", "test")
	f = config.NewConfigFiles()
	if len(f) != 4 {
		t.Errorf("invalid files: %v", f)
	}
	t.Setenv("LOCKBOX_CONFIG_TOML", "detect")
	os.Unsetenv("HOME")
	f = config.NewConfigFiles()
	if len(f) != 2 {
		t.Errorf("invalid files: %v", f)
	}
}

func TestCanColor(t *testing.T) {
	store.Clear()
	if can, _ := config.CanColor(); !can {
		t.Error("should be able to color")
	}
	for raw, expect := range map[string]bool{
		"INTERACTIVE":   true,
		"COLOR_ENABLED": true,
	} {
		store.Clear()
		key := fmt.Sprintf("LOCKBOX_%s", raw)
		store.SetBool(key, true)
		if can, _ := config.CanColor(); can != expect {
			t.Errorf("expect != actual: %s", key)
		}
		store.SetBool(key, false)
		if can, _ := config.CanColor(); can == expect {
			t.Errorf("expect == actual: %s", key)
		}
	}
	store.Clear()
	t.Setenv("NO_COLOR", "1")
	if can, _ := config.CanColor(); can {
		t.Error("should NOT be able to color")
	}
}
