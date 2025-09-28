package platform_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/config/store"
	"github.com/enckse/lockbox/internal/platform"
)

func TestPathExist(t *testing.T) {
	testDir := filepath.Join("testdata", "exists")
	os.RemoveAll(testDir)
	if platform.PathExists(testDir) {
		t.Error("test dir SHOULD NOT exist")
	}
	os.MkdirAll(testDir, 0o755)
	if !platform.PathExists(testDir) {
		t.Error("test dir SHOULD exist")
	}
}

func TestLoadConfigFile(t *testing.T) {
	store.Clear()
	os.Mkdir("testdata", 0o755)
	defer os.RemoveAll("testdata")
	file := filepath.Join("testdata", "config.toml")
	loaded, err := config.DefaultTOML()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.WriteFile(file, []byte(loaded), 0o644)
	if err := platform.LoadConfigFile(file); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 16 {
		t.Errorf("invalid environment after load: %d", len(store.List()))
	}
}

func TestLoadConfigFileNoFile(t *testing.T) {
	store.Clear()
	os.Mkdir("testdata", 0o755)
	defer os.RemoveAll("testdata")
	file := filepath.Join("testdata", "config.toml")
	loaded, err := config.DefaultTOML()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	loaded = strings.Replace(loaded, "include = []", "include = ['invalid.toml']", 1)
	os.WriteFile(file, []byte(loaded), 0o644)
	if err := platform.LoadConfigFile(file); err == nil || err.Error() != "failed to load the included file: invalid.toml" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestLoadConfigFileNoFileNotRequired(t *testing.T) {
	store.Clear()
	os.Mkdir("testdata", 0o755)
	defer os.RemoveAll("testdata")
	file := filepath.Join("testdata", "config.toml")
	loaded, err := config.DefaultTOML()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	loaded = strings.Replace(loaded, "include = []", "include = [{file = 'invalid.toml', required = false}]", 1)
	os.WriteFile(file, []byte(loaded), 0o644)
	if err := platform.LoadConfigFile(file); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestLoadConfigFileNoFileRequired(t *testing.T) {
	store.Clear()
	os.Mkdir("testdata", 0o755)
	defer os.RemoveAll("testdata")
	file := filepath.Join("testdata", "config.toml")
	loaded, err := config.DefaultTOML()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	loaded = strings.Replace(loaded, "include = []", "include = [{file = 'invalid.toml'}]", 1)
	os.WriteFile(file, []byte(loaded), 0o644)
	if err := platform.LoadConfigFile(file); err == nil || err.Error() != "failed to load the included file: invalid.toml" {
		t.Errorf("invalid error: %v", err)
	}
}
