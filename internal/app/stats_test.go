package app_test

import (
	"bytes"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestStats(t *testing.T) {
	m := newMockCommand(t)
	if err := app.Stats(m); err.Error() != "entry required" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"test/test2/test1"}
	if err := app.Stats(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("no stats")
	}
	m.buf = bytes.Buffer{}
	m.args = []string{"tsest/test2/test1"}
	if err := app.Stats(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("no stats")
	}
}
