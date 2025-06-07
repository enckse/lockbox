package app_test

import (
	"bytes"
	"os"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app"
)

func TestShowClip(t *testing.T) {
	m := newMockCommand(t)
	if err := app.ShowClip(m, true); err.Error() != "only one argument supported" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"test/test2/test1/password"}
	if err := app.ShowClip(m, true); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Error("no show")
	}
	m.buf = bytes.Buffer{}
	m.args = []string{"test211/test2/test"}
	if err := app.ShowClip(m, true); err == nil || err.Error() != "entry does not exist" {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Error("no show")
	}
	os.Clearenv()
	m.args = []string{"tsest/test2/test1"}
	if err := app.ShowClip(m, false); err == nil {
		t.Errorf("invalid error: %v", err)
	}
}
