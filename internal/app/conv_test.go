package app_test

import (
	"bytes"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
)

func TestConv(t *testing.T) {
	fullSetup(t, false)
	c := newMockCommand(t)
	if err := app.Conv(c); err.Error() != "conv requires a file" {
		t.Errorf("invalid error: %v", err)
	}
	c.buf = bytes.Buffer{}
	c.args = []string{testFile()}
	if err := app.Conv(c); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.buf.String() == "" {
		t.Error("nothing converted")
	}
}
