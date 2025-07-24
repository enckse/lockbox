package commands_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/enckse/lockbox/internal/app/commands"
	"github.com/enckse/lockbox/internal/config/store"
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

func TestReadOnlyOrder(t *testing.T) {
	had := fmt.Sprintf("%v", commands.ReadOnly)
	v := commands.ReadOnly
	sort.Strings(v)
	if had != fmt.Sprintf("%v", v) {
		t.Error("invalid readonly sort")
	}
}
