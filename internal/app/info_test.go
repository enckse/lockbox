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
	ok, err = app.Info(&buf, "help", []string{"verbose"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" || old == buf.String() {
		t.Error("nothing written")
	}
	if _, err = app.Info(&buf, "help", []string{"-verb"}); err.Error() != "invalid help option" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "help", []string{"verbose", "A"}); err.Error() != "invalid help command" {
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
	ok, err = app.Info(&buf, "bash", []string{"help"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	if _, err = app.Info(&buf, "bash", []string{"defaults"}); err.Error() != "invalid bash subcommand" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "bash", []string{"test", "default"}); err.Error() != "invalid bash command" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "bash", []string{"short"}); err.Error() != "invalid bash subcommand" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestEnvInfo(t *testing.T) {
	os.Clearenv()
	var buf bytes.Buffer
	ok, err := app.Info(&buf, "env", []string{})
	if ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() != "" {
		t.Error("nothing written")
	}
	os.Setenv("LOCKBOX_STORE", "1")
	ok, err = app.Info(&buf, "env", []string{})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	if _, err = app.Info(&buf, "env", []string{"defaults"}); err.Error() != "invalid env command" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "env", []string{"test", "default"}); err.Error() != "invalid env command" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestZshInfo(t *testing.T) {
	os.Clearenv()
	var buf bytes.Buffer
	ok, err := app.Info(&buf, "zsh", []string{})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	buf = bytes.Buffer{}
	ok, err = app.Info(&buf, "zsh", []string{"help"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing written")
	}
	if _, err = app.Info(&buf, "zsh", []string{"defaults"}); err.Error() != "invalid zsh subcommand" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "zsh", []string{"test", "default"}); err.Error() != "invalid zsh command" {
		t.Errorf("invalid error: %v", err)
	}
}
