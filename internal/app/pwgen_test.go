package app_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
)

func setupGenScript() string {
	os.Clearenv()
	const pwgenScript = "pwgen.sh"
	pwgenPath := filepath.Join("testdata", pwgenScript)
	os.WriteFile(pwgenPath, []byte(`#!/bin/sh
for f in $@; do
  echo $f
done
`), 0o755)
	return pwgenPath
}

func TestGenerateError(t *testing.T) {
	m := newMockCommand(t)
	pwgenPath := setupGenScript()
	t.Setenv("LOCKBOX_PWGEN_WORD_COUNT", "0")
	if err := app.GeneratePassword(m); err == nil || err.Error() != "word count must be > 0" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_PWGEN_WORD_COUNT", "1")
	if err := app.GeneratePassword(m); err == nil || err.Error() != "word list command must set" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", "1 x")
	if err := app.GeneratePassword(m); err == nil || !strings.Contains(err.Error(), "exec: \"1\":") {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", pwgenPath)
	if err := app.GeneratePassword(m); err == nil || err.Error() != "no sources given" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", fmt.Sprintf("%s 1", pwgenPath))
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", fmt.Sprintf("%s aloj 1", pwgenPath))
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_PWGEN_ENABLED", "no")
	if err := app.GeneratePassword(m); err == nil || err.Error() != "password generation is disabled" {
		t.Errorf("invalid error: %v", err)
	}
}

func testPasswordGen(t *testing.T, expect string) {
	m := newMockCommand(t)
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s := m.buf.String()
	if s != fmt.Sprintf("%s\n", expect) {
		t.Errorf("invalid generated: %s (expected: %s)", s, expect)
	}
}

func TestGenerate(t *testing.T) {
	pwgenPath := setupGenScript()
	t.Setenv("LOCKBOX_PWGEN_WORD_COUNT", "1")
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", fmt.Sprintf("%s 1", pwgenPath))
	testPasswordGen(t, "1")
	t.Setenv("LOCKBOX_PWGEN_WORD_COUNT", "10")
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", fmt.Sprintf("%s 1 1 1 1 1 1 1 1 1 1 1 1", pwgenPath))
	testPasswordGen(t, "1-1-1-1-1-1-1-1-1-1")
	t.Setenv("LOCKBOX_PWGEN_WORD_COUNT", "4")
	t.Setenv("LOCKBOX_PWGEN_TITLE", "yes")
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", fmt.Sprintf("%s a a a a a a a a a a a a a a a", pwgenPath))
	testPasswordGen(t, "A-A-A-A")
	t.Setenv("LOCKBOX_PWGEN_CHARACTERS", "bc")
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", fmt.Sprintf("%s abc abc abc abc abc aaa aaa aa a", pwgenPath))
	testPasswordGen(t, "Bc-Bc-Bc-Bc")
	os.Unsetenv("LOCKBOX_PWGEN_CHARACTERS")
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", fmt.Sprintf("%s a a a a a a a a a a a a a a a", pwgenPath))
	t.Setenv("LOCKBOX_PWGEN_TITLE", "no")
	t.Setenv("LOCKBOX_PWGEN_TITLE", "no")
	testPasswordGen(t, "a-a-a-a")
	// NOTE: this allows templating below in golang
	t.Setenv("DOLLAR", "$")
	t.Setenv("LOCKBOX_PWGEN_TEMPLATE", "{{range ${DOLLAR}idx, ${DOLLAR}val := .}}{{if lt ${DOLLAR}idx 5}}-{{end}}{{ ${DOLLAR}val.Text }}{{ ${DOLLAR}val.Position.Start }}{{ ${DOLLAR}val.Position.End }}{{end}}")
	testPasswordGen(t, "-a01-a12-a23-a34")
	t.Setenv("LOCKBOX_PWGEN_TEMPLATE", "{{range [%]idx, [%]val := .}}{{if lt [%]idx 5}}-{{end}}{{ [%]val.Text }}{{end}}")
	testPasswordGen(t, "-a-a-a-a")
	os.Unsetenv("LOCKBOX_PWGEN_TEMPLATE")
	t.Setenv("LOCKBOX_PWGEN_TITLE", "yes")
	t.Setenv("LOCKBOX_PWGEN_WORDS_COMMAND", fmt.Sprintf("%s abc axy axY aZZZ aoijafea aoiajfoea afaeoa", pwgenPath))
	m := newMockCommand(t)
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s := m.buf.String()
	if s[0] != 'A' {
		t.Errorf("no title: %s", s)
	}
	if len(s) < 5 {
		t.Errorf("bad result: %s", s)
	}
}
