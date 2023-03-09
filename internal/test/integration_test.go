package test_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/enckse/lockbox/internal/test"
	"github.com/enckse/pgl/os/paths"
)

const (
	testDir = "bin"
)

func runTest(t *testing.T, keyFile bool) {
	os.RemoveAll(testDir)
	os.Mkdir(testDir, 0o755)
	binary := filepath.Join("..", "..", testDir, "lb")
	if !paths.Exist(binary) {
		t.Error("no binary to test")
		return
	}
	logFile := filepath.Join(testDir, "actual.log")
	if err := test.Execute(keyFile, binary, testDir, logFile); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := test.Cleanup(logFile); err != nil {
		t.Errorf("cleanup failed: %v", err)
	}
	diff := exec.Command("diff", "-u", logFile, "expected.log")
	diff.Stdout = os.Stdout
	diff.Stderr = os.Stderr
	if err := diff.Run(); err != nil {
		t.Errorf("diff failed: %v", err)
	}
}

func TestKeyFile(t *testing.T) {
	runTest(t, true)
}

func TestPasswordOnly(t *testing.T) {
	runTest(t, false)
}
