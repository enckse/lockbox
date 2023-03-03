package app_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/backend"
)

type (
	mockConfirm struct {
		called bool
	}
)

func (m *mockConfirm) prompt(string) bool {
	m.called = true
	return true
}

func TestMove(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	m := mockConfirm{}
	if err := app.Move(fullSetup(t, true), []string{}, m.prompt); err.Error() != "src/dst required for move" {
		t.Errorf("invalid error: %v", err)
	}
	if err := app.Move(fullSetup(t, true), []string{"a", "b"}, m.prompt); err.Error() != "unable to get source entry" {
		t.Errorf("invalid error: %v", err)
	}
	m.called = false
	if err := app.Move(fullSetup(t, true), []string{"test/test2/test1", "test/test2/test3"}, m.prompt); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !m.called {
		t.Error("no move")
	}
}
