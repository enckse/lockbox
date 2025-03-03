package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app"
	"git.sr.ht/~enckse/lockbox/internal/backend"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
	"git.sr.ht/~enckse/lockbox/internal/platform"
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
	store.SetString("LOCKBOX_STORE", file)
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetString("LOCKBOX_TOTP_ENTRY", "totp")
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
