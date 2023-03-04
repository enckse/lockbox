package app_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/backend"
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
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
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
	m.args = []string{"a", "b"}
	if err := app.Move(m); err.Error() != "unable to get source entry" {
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
}
