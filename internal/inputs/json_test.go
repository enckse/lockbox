package inputs_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/inputs"
)

func TestParseJSONMode(t *testing.T) {
	defer os.Clearenv()
	m, err := inputs.ParseJSONOutput()
	if m != inputs.JSONDataOutputHash || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT", "hAsH ")
	m, err = inputs.ParseJSONOutput()
	if m != inputs.JSONDataOutputHash || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT", "EMPTY")
	m, err = inputs.ParseJSONOutput()
	if m != inputs.JSONDataOutputBlank || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT", " PLAINtext ")
	m, err = inputs.ParseJSONOutput()
	if m != inputs.JSONDataOutputRaw || err != nil {
		t.Error("invalid mode read")
	}
	os.Setenv("LOCKBOX_JSON_DATA_OUTPUT", "a")
	if _, err = inputs.ParseJSONOutput(); err == nil || err.Error() != "invalid JSON output mode: a" {
		t.Errorf("invalid error: %v", err)
	}
}
