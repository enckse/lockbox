package app_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/kdbx"
)

func TestUnset(t *testing.T) {
	m := newMockCommand(t)
	fullSetup(t, true).Insert(kdbx.NewPath("test", "test2", "testz"), map[string]string{"notes": "something"})
	if err := app.Unset(m); err == nil || err.Error() != "invalid unset, no entry given" {
		t.Errorf("invalid error: %v", err)
	}
	m.args = []string{"a/y/z"}
	if err := app.Unset(m); err == nil || err.Error() != "a/y/z does not exist" {
		t.Errorf("invalid error: %v", err)
	}
	m.confirm = false
	m.args = []string{"test/test2/test1/otp"}
	if err := app.Unset(m); err == nil || err.Error() != "unable to unset: test/test2/test1/otp" {
		t.Errorf("invalid error: %v", err)
	}
	m.confirm = false
	m.args = []string{"test/test2/test1/password"}
	if err := app.Unset(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Errorf("invalid operation")
	}
	m.buf.Reset()
	m.confirm = true
	m.args = []string{"test/test2/test1/password"}
	if err := app.Unset(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Errorf("invalid operation")
	}
	m.buf.Reset()
	m.args = []string{"test/test2/testz/notes"}
	if err := app.Unset(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" {
		t.Errorf("invalid operation")
	}
}
