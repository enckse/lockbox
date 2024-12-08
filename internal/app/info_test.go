package app_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
	"github.com/seanenck/lockbox/internal/config/store"
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
	store.Clear()
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
	ok, err = app.Info(&buf, "help", []string{"config"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" || old == buf.String() {
		t.Error("nothing written")
	}
}

func TestEnvInfo(t *testing.T) {
	os.Clearenv()
	store.Clear()
	var buf bytes.Buffer
	ok, err := app.Info(&buf, "var", []string{})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() != "\n" {
		t.Error("nothing written")
	}
	buf = bytes.Buffer{}
	store.SetString("LOCKBOX_STORE", "1")
	ok, err = app.Info(&buf, "var", []string{})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "LOCKBOX_STORE=1" {
		t.Error("nothing written")
	}
	buf = bytes.Buffer{}
	ok, err = app.Info(&buf, "var", []string{"completions"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() != "\n" {
		t.Error("nothing written")
	}
	store.SetString("LOCKBOX_READONLY", "true")
	buf = bytes.Buffer{}
	ok, err = app.Info(&buf, "var", []string{"completions"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "LOCKBOX_READONLY=true" {
		t.Error("nothing written")
	}
	buf = bytes.Buffer{}
	ok, err = app.Info(&buf, "var", []string{"LOCKBOX_READONLY"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "LOCKBOX_READONLY=true" {
		t.Error("nothing written")
	}
	buf = bytes.Buffer{}
	ok, err = app.Info(&buf, "var", []string{"garbage"})
	if !ok || err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() != "\n" {
		t.Error("nothing written")
	}
	if _, err = app.Info(&buf, "var", []string{"test", "default"}); err.Error() != "invalid env command, too many arguments" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestCompletionInfo(t *testing.T) {
	for k, v := range map[string]string{
		"zsh":  "typeset -A opt_args",
		"fish": "set -f commands",
		"bash": "local cur opts",
	} {
		for _, b := range []bool{true, false} {
			store.Clear()
			os.Clearenv()
			sub := []string{k}
			t.Setenv("SHELL", "invalid")
			if b {
				sub = []string{}
				t.Setenv("SHELL", k)
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
	if _, err := app.Info(&buf, "completions", []string{"helps"}); err.Error() != "unknown completion type: helps" {
		t.Errorf("invalid error: %v", err)
	}
	os.Clearenv()
	store.Clear()
	t.Setenv("SHELL", "bad")
	if _, err := app.Info(&buf, "completions", []string{}); err.Error() != "unknown completion type: bad" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.Info(&buf, "completions", []string{"bash"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
