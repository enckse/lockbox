package commands_test

import (
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app/commands"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

func TestIsReadOnly(t *testing.T) {
	defer store.Clear()
	if res := commands.AllowedInReadOnly("insert", "xyz"); len(res) != 2 {
		t.Error("invalid, is not readonly")
	}
	store.SetBool("LOCKBOX_READONLY", true)
	if res := commands.AllowedInReadOnly("insert"); len(res) != 0 {
		t.Error("invalid, is not readonly")
	}
	if res := commands.AllowedInReadOnly("insert", "show"); len(res) != 1 {
		t.Error("invalid, is not readonly")
	}
	if res := commands.AllowedInReadOnly("show"); len(res) != 1 {
		t.Error("invalid, is not readonly")
	}
}
