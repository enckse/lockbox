package app_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestNoInfo(t *testing.T) {
	var buf bytes.Buffer
	ok, err := app.Info(&buf, "", []string{})
	if ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestHelpInfo(t *testing.T) {
	os.Clearenv()
	var buf bytes.Buffer
	ok, err := app.Info(&buf, "help", []string{})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	old := buf.String()
	buf = bytes.Buffer{}
	ok, err = app.Info(&buf, "help", []string{"-verbose"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" || old == buf.String() {
		t.Error("nothing written")
	}
	if _, err = app.Info(&buf, "help", []string{"-verb"}); err.Error() != "invalid help option" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "help", []string{"-verbose", "A"}); err.Error() != "invalid help command" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestBashInfo(t *testing.T) {
	os.Clearenv()
	var buf bytes.Buffer
	ok, err := app.Info(&buf, "bash", []string{})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	buf = bytes.Buffer{}
	ok, err = app.Info(&buf, "bash", []string{"-defaults"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	if _, err = app.Info(&buf, "bash", []string{"-default"}); err.Error() != "invalid argument" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "bash", []string{"test", "-default"}); err.Error() != "invalid argument" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestEnvInfo(t *testing.T) {
	os.Clearenv()
	var buf bytes.Buffer
	ok, err := app.Info(&buf, "env", []string{})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	buf = bytes.Buffer{}
	ok, err = app.Info(&buf, "env", []string{"-defaults"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	if _, err = app.Info(&buf, "env", []string{"-default"}); err.Error() != "invalid argument" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "env", []string{"test", "-default"}); err.Error() != "invalid argument" {
		t.Errorf("invalid error: %v", err)
	}
}
