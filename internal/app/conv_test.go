package app_test

import (
	"bytes"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestConv(t *testing.T) {
	c := newMockCommand(t)
	if err := app.Conv(c); err.Error() != "hash requires a file" {
		t.Errorf("invalid error: %v", err)
	}
	c.buf = bytes.Buffer{}
	c.args = []string{"test.kdbx"}
	if err := app.Conv(c); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.buf.String() == "" {
		t.Error("nothing hashed")
	}
}
