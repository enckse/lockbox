package system_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/enckse/lockbox/internal/system"
)

func TestPathExist(t *testing.T) {
	testDir := filepath.Join("testdata", "exists")
	os.RemoveAll(testDir)
	if system.PathExists(testDir) {
		t.Error("test dir SHOULD NOT exist")
	}
	os.MkdirAll(testDir, 0o755)
	if !system.PathExists(testDir) {
		t.Error("test dir SHOULD exist")
	}
}
