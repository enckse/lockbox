package app_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

func setupGenScript() string {
	store.Clear()
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
	store.SetInt64("LOCKBOX_PWGEN_WORD_COUNT", 0)
	if err := app.GeneratePassword(m); err == nil || err.Error() != "word count must be > 0" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetInt64("LOCKBOX_PWGEN_WORD_COUNT", 1)
	if err := app.GeneratePassword(m); err == nil || err.Error() != "word list command must set" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{"1 x"})
	if err := app.GeneratePassword(m); err == nil || !strings.Contains(err.Error(), "exec: \"1 x\":") {
		t.Errorf("invalid error: %v", err)
	}
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath})
	if err := app.GeneratePassword(m); err == nil || err.Error() != "no sources given" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath, "1"})
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath, "aloj", "1"})
	if err := app.GeneratePassword(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	store.SetBool("LOCKBOX_PWGEN_ENABLED", false)
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
	store.SetInt64("LOCKBOX_PWGEN_WORD_COUNT", 1)
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath, "1"})
	testPasswordGen(t, "1")
	store.SetInt64("LOCKBOX_PWGEN_WORD_COUNT", 10)
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath, "1 1 1 1 1 1 1 1 1 1 1 1"})
	testPasswordGen(t, "1-1-1-1-1-1-1-1-1-1")
	store.SetInt64("LOCKBOX_PWGEN_WORD_COUNT", 4)
	store.SetBool("LOCKBOX_PWGEN_TITLE", true)
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath, "a a a a a a a a a a a a a a a a a a a a a a"})
	testPasswordGen(t, "A-A-A-A")
	store.SetString("LOCKBOX_PWGEN_CHARACTERS", "bc")
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath, "abc abc abc abc abc abc aaa aa aaa a"})
	testPasswordGen(t, "Bc-Bc-Bc-Bc")
	store.SetString("LOCKBOX_PWGEN_CHARACTERS", "")
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath, "a a a a a a a a a a a a a a a a a a a a a a"})
	store.SetBool("LOCKBOX_PWGEN_TITLE", false)
	testPasswordGen(t, "a-a-a-a")
	// NOTE: this allows templating below in golang
	store.SetString("LOCKBOX_PWGEN_TEMPLATE", "{{range $idx, $val := .}}{{if lt $idx 5}}-{{end}}{{ $val.Text }}{{ $val.Position.Start }}{{ $val.Position.End }}{{end}}")
	testPasswordGen(t, "-a01-a12-a23-a34")
	store.SetString("LOCKBOX_PWGEN_TEMPLATE", "{{range $idx, $val := .}}{{if lt $idx 5}}-{{end}}{{ $val.Text }}{{end}}")
	testPasswordGen(t, "-a-a-a-a")
	store.Clear()
	store.SetBool("LOCKBOX_PWGEN_TITLE", true)
	store.SetArray("LOCKBOX_PWGEN_WORDS_COMMAND", []string{pwgenPath, "abc axy axY aZZZ aoijafea aoiajfoea afeafa"})
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
