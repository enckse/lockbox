package commands_test

import (
	"bytes"
	"testing"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/commands"
)

func TestStats(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	tx := fullSetup(t, true)
	var b bytes.Buffer
	if err := commands.Stats(&b, tx, []string{}); err.Error() != "entry required" {
		t.Errorf("invalid error: %v", err)
	}
	if err := commands.Stats(&b, tx, []string{"test/test2/test1"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if b.String() == "" {
		t.Error("no stats")
	}
	b = bytes.Buffer{}
	if err := commands.Stats(&b, tx, []string{"tsest/test2/test1"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if b.String() != "" {
		t.Error("no stats")
	}
}
