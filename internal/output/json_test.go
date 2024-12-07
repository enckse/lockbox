package output_test

import (
	"fmt"
	"testing"

	"github.com/seanenck/lockbox/internal/output"
)

func TestJSONList(t *testing.T) {
	list := output.JSONModes.List()
	if len(list) != 3 || fmt.Sprintf("%v", list) != "[empty hash plaintext]" {
		t.Errorf("invalid list result: %v", list)
	}
}

func TestParseJSONMode(t *testing.T) {
	m, err := output.ParseJSONMode("hAsH ")
	if m != output.JSONModes.Hash || err != nil {
		t.Error("invalid mode read")
	}
	m, err = output.ParseJSONMode("EMPTY")
	if m != output.JSONModes.Blank || err != nil {
		t.Error("invalid mode read")
	}
	m, err = output.ParseJSONMode(" PLAINtext ")
	if m != output.JSONModes.Raw || err != nil {
		t.Error("invalid mode read")
	}
	if _, err = output.ParseJSONMode("a"); err == nil || err.Error() != "invalid JSON output mode: a" {
		t.Errorf("invalid error: %v", err)
	}
}
