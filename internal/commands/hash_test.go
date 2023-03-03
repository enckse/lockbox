package commands_test

import (
	"bytes"
	"testing"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/commands"
)

func TestHash(t *testing.T) {
	var buf bytes.Buffer
	if err := commands.Hash(&buf, []string{}); err.Error() != "hash requires a file" {
		t.Errorf("invalid error: %v", err)
	}
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	if err := commands.Hash(&buf, []string{"test.kdbx"}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("nothing hashed")
	}
}
