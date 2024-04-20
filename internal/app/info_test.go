package app_test

import (
	"bytes"
	"os"
	"strings"
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

func TestCompletionInfo(t *testing.T) {
	defer os.Clearenv()
	for k, v := range map[string]string{
		"zsh":  "typeset -A opt_args",
		"fish": "set -l commands",
		"bash": "local cur opts",
	} {
		for _, b := range []bool{true, false} {
			os.Clearenv()
			sub := []string{k}
			os.Setenv("SHELL", "invalid")
			if b {
				sub = []string{}
				os.Setenv("SHELL", k)
			}
			var buf bytes.Buffer
			ok, err := app.Info(&buf, "completions", sub)
			if !ok || err != nil {
				t.Errorf("invalid error: %v", err)
			}
			s := buf.String()
			if s == "" {
				t.Error("nothing written")
			}
			if !strings.Contains(s, v) {
				t.Errorf("invalid completions for %s", k)
			}
		}
	}
	var buf bytes.Buffer
	ok, err := app.Info(&buf, "completions", []string{"help"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s := buf.String()
	if s == "" {
		t.Error("nothing written")
	}

	if _, err = app.Info(&buf, "completions", []string{"helps"}); err.Error() != "unknown completion type: helps" {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	os.Setenv("SHELL", "bad")
	if _, err = app.Info(&buf, "completions", []string{}); err.Error() != "unknown completion type: bad" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err = app.Info(&buf, "completions", []string{"bash"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
