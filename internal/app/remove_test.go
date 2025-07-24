package app_test

import (
	"bytes"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestRemove(t *testing.T) {
	m := newMockCommand(t)
	m.buf = bytes.Buffer{}
	if err := app.Remove(m); err.Error() != "remove requires an entry" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"a", "b"}
	if err := app.Remove(m); err.Error() != "remove requires an entry" {
		t.Errorf("invalid error: %v", err)
	}
	m.confirmed = false
	m.args = []string{"tzzzest/test2/test1"}
	if err := app.Remove(m); err.Error() != "no entities matching: tzzzest/test2/test1" {
		t.Errorf("invalid error: %v", err)
	}
	if m.confirmed {
		t.Error("no remove")
	}
	m.confirmed = false
	m.args = []string{"test/test2/test1"}
	if err := app.Remove(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.confirmed {
		t.Error("no remove")
	}
}
