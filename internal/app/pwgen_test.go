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
	defer os.Clearenv()
	m := newMockCommand(t)
	pwgenPath := setupGenScript()
	os.Setenv("LOCKBOX_PWGEN_COUNT", "0")
	if err := app.GeneratePassword(m); err == nil || err.Error() != "word count must be > 0" {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_PWGEN_COUNT", "1")
	if err := app.GeneratePassword(m); err == nil || err.Error() != "word list command must set" {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_PWGEN_WORDLIST", "1 x")
	if err := app.GeneratePassword(m); err == nil || !strings.Contains(err.Error(), "exec: \"1\":") {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_PWGEN_WORDLIST", pwgenPath)
	if err := app.GeneratePassword(m); err == nil || err.Error() != "choices <= word count requested" {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_PWGEN_WORDLIST", fmt.Sprintf("%s 1", pwgenPath))
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_PWGEN_WORDLIST", fmt.Sprintf("%s aloj 1", pwgenPath))
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func testPasswordGen(t *testing.T, expect string) {
	m := newMockCommand(t)
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s := m.buf.String()
	if s != expect {
		t.Errorf("invalid generated: %s (expected: %s)", s, expect)
	}
}

func TestGenerate(t *testing.T) {
	defer os.Clearenv()
	pwgenPath := setupGenScript()
	os.Setenv("LOCKBOX_PWGEN_COUNT", "1")
	os.Setenv("LOCKBOX_PWGEN_WORDLIST", fmt.Sprintf("%s 1", pwgenPath))
	testPasswordGen(t, "1")
	os.Setenv("LOCKBOX_PWGEN_COUNT", "10")
	os.Setenv("LOCKBOX_PWGEN_WORDLIST", fmt.Sprintf("%s 1 1 1 1 1 1 1 1 1 1 1 1", pwgenPath))
	testPasswordGen(t, "1-1-1-1-1-1-1-1-1-1")
	os.Setenv("LOCKBOX_PWGEN_COUNT", "4")
	os.Setenv("LOCKBOX_PWGEN_TITLE", "yes")
	os.Setenv("LOCKBOX_PWGEN_WORDLIST", fmt.Sprintf("%s a a a a a a a a a a a a a a a", pwgenPath))
	testPasswordGen(t, "A-A-A-A")
	os.Setenv("LOCKBOX_PWGEN_TITLE", "no")
	testPasswordGen(t, "a-a-a-a")
	// NOTE: this allows templating below in golang
	os.Setenv("DOLLAR", "$")
	os.Setenv("LOCKBOX_PWGEN_TEMPLATE", "{{range ${DOLLAR}idx, ${DOLLAR}val := .}}{{if lt ${DOLLAR}idx 5}}-{{end}}{{ ${DOLLAR}val }}{{end}}")
	testPasswordGen(t, "-a-a-a-a")
	os.Unsetenv("LOCKBOX_PWGEN_TEMPLATE")
	os.Setenv("LOCKBOX_PWGEN_TITLE", "yes")
	os.Setenv("LOCKBOX_PWGEN_WORDLIST", fmt.Sprintf("%s abc axy axY aZZZ aoijafea aoiajfoea afaeoa", pwgenPath))
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
