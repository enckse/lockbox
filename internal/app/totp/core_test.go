package totp_test

import (
	"bytes"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app/totp"
)

func TestPrint(t *testing.T) {
	generator, _ := totp.New("5ae472abqdekjqykoyxk7hvc2leklq5n")
	var buf bytes.Buffer
	generator.Print(&buf, false)
	if strings.TrimSpace(buf.String()) != "5ae472abqdekjqykoyxk7hvc2leklq5n" {
		t.Errorf("invalid buffer: %s", buf.String())
	}
	buf = bytes.Buffer{}
	generator.Print(&buf, true)
	count := 0
	hasBlank := false
	for line := range strings.SplitSeq(buf.String(), "\n") {
		if line == "" {
			if hasBlank {
				t.Errorf("already have blank line")
			}
			hasBlank = true
			continue
		}
		count++
		if !strings.Contains(line, ":") {
			t.Errorf("line missing colon: %s", line)
		}
	}
	if count != 5 {
		t.Errorf("invalid buffer: %s", buf.String())
	}
}

func TestCode(t *testing.T) {
	generator, _ := totp.New("5ae472abqdekjqykoyxk7hvc2leklq5n")
	code, err := generator.Code()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("invalid code: %s", code)
	}
}

func TestNew(t *testing.T) {
	if _, err := totp.New("5ae472abqdekjqykoyxk7hvc2leklq5n"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
