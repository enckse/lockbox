package core_test

import (
	"fmt"
	"testing"

	"github.com/seanenck/lockbox/internal/core"
)

func TestJSONList(t *testing.T) {
	list := core.JSONOutputs.List()
	if len(list) != 3 || fmt.Sprintf("%v", list) != "[empty hash plaintext]" {
		t.Errorf("invalid list result: %v", list)
	}
}

func TestParseJSONMode(t *testing.T) {
	m, err := core.ParseJSONOutput("hAsH ")
	if m != core.JSONOutputs.Hash || err != nil {
		t.Error("invalid mode read")
	}
	m, err = core.ParseJSONOutput("EMPTY")
	if m != core.JSONOutputs.Blank || err != nil {
		t.Error("invalid mode read")
	}
	m, err = core.ParseJSONOutput(" PLAINtext ")
	if m != core.JSONOutputs.Raw || err != nil {
		t.Error("invalid mode read")
	}
	if _, err = core.ParseJSONOutput("a"); err == nil || err.Error() != "invalid JSON output mode: a" {
		t.Errorf("invalid error: %v", err)
	}
}
