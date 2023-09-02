package platform_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/enckse/lockbox/internal/config"
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

func TestReadKey(t *testing.T) {
	o, err := platform.ReadKey(nil, nil)
	if o != "" || err == nil || err.Error() != "invalid function given" {
		t.Errorf("invalid error: %v", err)
	}
	fxn := func() (string, error) {
		return "", nil
	}
	o, err = platform.ReadKey(nil, fxn)
	if o != "" || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	o, err = platform.ReadKey(&config.Key{}, fxn)
	if o != "" || err == nil || err.Error() != "interactive password can NOT be empty" {
		t.Errorf("invalid error: %v", err)
	}
	fxn = func() (string, error) {
		return "abc", errors.New("test error")
	}
	o, err = platform.ReadKey(&config.Key{}, fxn)
	if o != "" || err == nil || err.Error() != "test error" {
		t.Errorf("invalid error: %v", err)
	}
	fxn = func() (string, error) {
		return "abc", nil
	}
	o, err = platform.ReadKey(&config.Key{}, fxn)
	if o != "abc" || err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
