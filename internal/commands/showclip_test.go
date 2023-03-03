package commands_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/commands"
)

func TestShowClip(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	tx := fullSetup(t, true)
	var b bytes.Buffer
	if err := commands.ShowClip(&b, tx, true, []string{}); err.Error() != "entry required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := commands.ShowClip(&b, tx, true, []string{"test/test2/test1"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if b.String() == "" {
		t.Error("no show")
	}
	b = bytes.Buffer{}
	if err := commands.ShowClip(&b, tx, true, []string{"tsest/test2/test1"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if b.String() != "" {
		t.Error("no show")
	}
	os.Clearenv()
	if err := commands.ShowClip(&b, tx, false, []string{"tsest/test2/test1"}); err == nil {
		t.Errorf("invalid error: %v", err)
	}
}
