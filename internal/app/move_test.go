package app_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
	"github.com/seanenck/lockbox/internal/backend"
)

type (
	mockCommand struct {
		confirmed bool
		confirm   bool
		args      []string
		t         *testing.T
		buf       bytes.Buffer
	}
)

func newMockCommand(t *testing.T) *mockCommand {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test2"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test3", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test3", "test2"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test4", "test5"), "pass")
	return &mockCommand{t: t, confirmed: false, confirm: true}
}

func (m *mockCommand) Confirm(string) bool {
	m.confirmed = true
	return m.confirm
}

func (m *mockCommand) Transaction() *backend.Transaction {
	return fullSetup(m.t, true)
}

func (m *mockCommand) Args() []string {
	return m.args
}

func (m *mockCommand) Writer() io.Writer {
	return &m.buf
}

func TestMove(t *testing.T) {
	m := newMockCommand(t)
	if err := app.Move(m); err.Error() != "src/dst required for move" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"a/c", "b"}
	if err := app.Move(m); err.Error() != "no source entries matched" {
		t.Errorf("invalid error: %v", err)
	}
	m.confirmed = false
	m.args = []string{"test/test2/test1", "test/test2/test3"}
	if err := app.Move(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.confirmed {
		t.Error("no move")
	}
	m.args = []string{"test/test3/*", "test/test2/test3"}
	if err := app.Move(m); err.Error() != "test/test2/test3 must be a path, not an entry" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"test/*", "test/test2/"}
	if err := app.Move(m); err.Error() != "multiple moves can only be done at a leaf level" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"test/test3/*", "test/test2/"}
	if err := app.Move(m); err.Error() != "unable to overwrite entries when moving multiple items" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"test/test3/*", "test/test4/"}
	if err := app.Move(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
