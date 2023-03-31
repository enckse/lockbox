package app_test

import (
	"bytes"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestJSON(t *testing.T) {
	m := newMockCommand(t)
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"test", "test2"}
	if err := app.JSON(m); err.Error() != "invalid arguments" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"test/test2/test1"}
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("no stats")
	}
	m.buf = bytes.Buffer{}
	m.args = []string{"tsest/test2/test1"}
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("no stats")
	}
}
