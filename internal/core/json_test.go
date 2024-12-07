package core_test

import (
	"testing"

	"github.com/seanenck/lockbox/internal/core"
)

func TestJSONList(t *testing.T) {
	if len(core.JSONOutputs.List()) != 3 {
		t.Errorf("invalid list result")
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
