package app_test

import (
	"bytes"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
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
	m.args = []string{"test2/test1"}
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" || m.buf.String() == "{}\n" {
		t.Error("no data")
	}
	m.buf = bytes.Buffer{}
	m.args = []string{"test/test2/test1"}
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" || m.buf.String() == "{}\n" {
		t.Error("no data")
	}
	m.buf = bytes.Buffer{}
	m.args = []string{"tsest/test2/test1"}
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "{}\n" {
		t.Error("no data")
	}
}
