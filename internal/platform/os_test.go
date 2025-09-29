package platform_test

import (
	"os"
	"path/filepath"
	"testing"

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
