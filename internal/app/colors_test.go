package app_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestHasColoring(t *testing.T) {
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_NOCOLOR", "no")
	term, err := app.NewTerminal(app.Red)
	if err != nil {
		t.Errorf("color was valid: %v", err)
	}
	if term.Start != "\033[1;31m" || term.End != "\033[0m" {
		t.Error("bad resulting color")
	}
}

func TestBadColor(t *testing.T) {
	_, err := app.NewTerminal(app.Color(5))
	if err == nil || err.Error() != "bad color" {
		t.Errorf("invalid color error: %v", err)
	}
}

func TestNoColoring(t *testing.T) {
	os.Setenv("LOCKBOX_INTERACTIVE", "no")
	os.Setenv("LOCKBOX_NOCOLOR", "yes")
	term, err := app.NewTerminal(app.Red)
	if err != nil {
		t.Errorf("color was valid: %v", err)
	}
	if term.Start != "" || term.End != "" {
		t.Error("should have no color")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_NOCOLOR", "yes")
	term, err = app.NewTerminal(app.Red)
	if err != nil {
		t.Errorf("color was valid: %v", err)
	}
	if term.Start != "" || term.End != "" {
		t.Error("should have no color")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "no")
	os.Setenv("LOCKBOX_NOCOLOR", "no")
	term, err = app.NewTerminal(app.Red)
	if err != nil {
		t.Errorf("color was valid: %v", err)
	}
	if term.Start != "" || term.End != "" {
		t.Error("should have no color")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_NOCOLOR", "no")
	term, err = app.NewTerminal(app.Red)
	if err != nil {
		t.Errorf("color was valid: %v", err)
	}
	if term.Start == "" || term.End == "" {
		t.Error("should have color")
	}
}
