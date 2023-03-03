package commands_test

import (
	"bytes"
	"testing"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/commands"
)

func TestRemove(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	m := mockConfirm{}
	var buf bytes.Buffer
	if err := commands.Remove(&buf, fullSetup(t, true), []string{}, m.prompt); err.Error() != "remove requires an entry" {
		t.Errorf("invalid error: %v", err)
	}
	if err := commands.Remove(&buf, fullSetup(t, true), []string{"a", "b"}, m.prompt); err.Error() != "remove requires an entry" {
		t.Errorf("invalid error: %v", err)
	}
	m.called = false
	if err := commands.Remove(&buf, fullSetup(t, true), []string{"tzzzest/test2/test1"}, m.prompt); err.Error() != "unable to remove: no entities given" {
		t.Errorf("invalid error: %v", err)
	}
	if !m.called {
		t.Error("no remove")
	}
}