package inputs_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/inputs"
)

func TestEnvDefault(t *testing.T) {
	os.Clearenv()
	val := inputs.EnvironOrDefault("TEST", "value")
	if val != "value" {
		t.Error("invalid read")
	}
	os.Setenv("TEST", "  ")
	val = inputs.EnvironOrDefault("TEST", "value")
	if val != "value" {
		t.Error("invalid read")
	}
	os.Setenv("TEST", " a")
	val = inputs.EnvironOrDefault("TEST", "value")
	if val != " a" {
		t.Error("invalid read")
	}
}

func TestPlatformSet(t *testing.T) {
	if len(inputs.PlatformSet()) != 4 {
		t.Error("invalid platform set")
	}
}
