package app_test

import (
	"bytes"
	"encoding/json"
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
	m.args = []string{}
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" || m.buf.String() == "{}\n" {
		t.Error("no data")
	}
	m.buf = bytes.Buffer{}
	m.args = []string{"test"}
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" || m.buf.String() == "{}\n" {
		t.Error("no data")
	}
	m.buf = bytes.Buffer{}
	m.args = []string{"test/test2"}
	if err := app.JSON(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() == "" || m.buf.String() == "{}\n" {
		t.Error("no data")
	}
	m.buf = bytes.Buffer{}
	m.args = []string{"test/test2/*"}
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
	var check any
	if err := json.Unmarshal(m.buf.Bytes(), &check); err != nil {
		t.Errorf("invalid json: %v", err)
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
