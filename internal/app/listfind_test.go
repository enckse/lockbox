package app_test

import (
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/backend"
)

func fullSetup(t *testing.T, keep bool) *backend.Transaction {
	if !keep {
		os.Remove("test.kdbx")
	}
	os.Setenv("LOCKBOX_READONLY", "no")
	os.Setenv("LOCKBOX_STORE", "test.kdbx")
	os.Setenv("LOCKBOX_KEY", "test")
	os.Setenv("LOCKBOX_KEYFILE", "")
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_TOTP", "totp")
	os.Setenv("LOCKBOX_HOOKDIR", "")
	os.Setenv("LOCKBOX_SET_MODTIME", "")
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
	if err := app.ListFind(m, false); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("nothing listed")
	}
	m.args = []string{"test"}
	if err := app.ListFind(m, false); err.Error() != "list does not support any arguments" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestFind(t *testing.T) {
	m := newMockCommand(t)
	if err := app.ListFind(m, true); err.Error() != "find requires search term" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"test1"}
	if err := app.ListFind(m, true); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" || strings.Contains(m.buf.String(), "test3") {
		t.Error("wrong find")
	}
}
