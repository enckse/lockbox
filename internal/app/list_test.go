package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
	"git.sr.ht/~enckse/lockbox/internal/kdbx"
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

func fullSetup(t *testing.T, keep bool) *kdbx.Transaction {
	file := testFile()
	if !keep {
		os.Remove(file)
	}
	store.SetString("LOCKBOX_STORE", file)
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetString("LOCKBOX_TOTP_ENTRY", "totp")
	tr, err := kdbx.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	return tr
}

func setup(t *testing.T) *kdbx.Transaction {
	return fullSetup(t, false)
}

func TestList(t *testing.T) {
	m := newMockCommand(t)
	if err := app.List(m, false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("nothing listed")
	}
	m.args = []string{"test", "test2"}
	if err := app.List(m, false); err == nil || err.Error() != "too many arguments (none or filter)" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestFind(t *testing.T) {
	m := newMockCommand(t)
	m.args = []string{"["}
	if err := app.List(m, false); err == nil || !strings.Contains(err.Error(), "missing closing") {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("something listed")
	}
	m.buf.Reset()
	m.args = []string{"[zzzzzz]"}
	if err := app.List(m, false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("something listed")
	}
	m.buf.Reset()
	m.args = []string{"test"}
	if err := app.List(m, false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("nothing listed")
	}
	m.buf.Reset()
	m.args = []string{"test"}
	if err := app.List(m, true); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("nothing listed")
	}
	m.buf.Reset()
	m.args = []string{"[zzzz]"}
	if err := app.List(m, true); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("something listed")
	}
}

func TestGroups(t *testing.T) {
	m := newMockCommand(t)
	if err := app.List(m, true); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Errorf("nothing listed: %s", m.buf.String())
	}
	m.args = []string{"test", "test2"}
	if err := app.List(m, true); err == nil || err.Error() != "too many arguments (none or filter)" {
		t.Errorf("invalid error: %v", err)
	}
}
