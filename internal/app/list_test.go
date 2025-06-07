package app_test

import (
	"os"
	"path/filepath"
	"strings"
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
	if err := app.List(m, false, false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("nothing listed")
	}
	m.args = []string{"test"}
	if err := app.List(m, false, false); err == nil || err.Error() != "arguments not supported" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestFind(t *testing.T) {
	m := newMockCommand(t)
	if err := app.List(m, true, false); err == nil || err.Error() != "find requires one argument" {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("something listed")
	}
	m.buf.Reset()
	m.args = []string{"["}
	if err := app.List(m, true, false); err == nil || !strings.Contains(err.Error(), "missing closing") {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("something listed")
	}
	m.buf.Reset()
	m.args = []string{"test", "1"}
	if err := app.List(m, true, false); err == nil || err.Error() != "find requires one argument" {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("something listed")
	}
	m.buf.Reset()
	m.args = []string{"[zzzzzz]"}
	if err := app.List(m, true, false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("something listed")
	}
	m.buf.Reset()
	m.args = []string{"test"}
	if err := app.List(m, true, false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("nothing listed")
	}
}

func TestGroups(t *testing.T) {
	m := newMockCommand(t)
	if err := app.List(m, false, true); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Errorf("nothing listed: %s", m.buf.String())
	}
	m.args = []string{"test"}
	if err := app.List(m, false, true); err == nil || err.Error() != "arguments not supported" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{}
	if err := app.List(m, true, true); err == nil || err.Error() != "groups+find not supported" {
		t.Errorf("invalid error: %v", err)
	}
}
