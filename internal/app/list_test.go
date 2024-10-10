package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
	"github.com/seanenck/lockbox/internal/backend"
	"github.com/seanenck/lockbox/internal/platform"
)

func testFile() string {
	dir := "testdata"
	file := filepath.Join(dir, "test.kdbx")
	if !platform.PathExists(dir) {
		os.Mkdir(dir, 0o755)
	}
	return file
}

func fullSetup(t *testing.T, keep bool) *backend.Transaction {
	file := testFile()
	if !keep {
		os.Remove(file)
	}
	t.Setenv("LOCKBOX_READONLY", "no")
	t.Setenv("LOCKBOX_STORE", file)
	t.Setenv("LOCKBOX_KEY", "test")
	t.Setenv("LOCKBOX_KEYFILE", "")
	t.Setenv("LOCKBOX_KEYMODE", "plaintext")
	t.Setenv("LOCKBOX_TOTP", "totp")
	t.Setenv("LOCKBOX_HOOKDIR", "")
	t.Setenv("LOCKBOX_SET_MODTIME", "")
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	return tr
}

func setup(t *testing.T) *backend.Transaction {
	return fullSetup(t, false)
}

func TestList(t *testing.T) {
	m := newMockCommand(t)
	if err := app.List(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("nothing listed")
	}
	m.args = []string{"test"}
	if err := app.List(m); err.Error() != "list does not support any arguments" {
		t.Errorf("invalid error: %v", err)
	}
}
