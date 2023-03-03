package commands_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/commands"
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
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	tx := fullSetup(t, true)
	var buf bytes.Buffer
	if err := commands.ListFind(tx, &buf, "list", []string{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing listed")
	}
	if err := commands.ListFind(tx, &buf, "list", []string{"test"}); err.Error() != "list does not support any arguments" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestFind(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	tx := fullSetup(t, true)
	var buf bytes.Buffer
	if err := commands.ListFind(tx, &buf, "find", []string{}); err.Error() != "find requires search term" {
		t.Errorf("invalid error: %v", err)
	}
	if err := commands.ListFind(tx, &buf, "find", []string{"test1"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" || strings.Contains(buf.String(), "test3") {
		t.Error("wrong find")
	}
}
