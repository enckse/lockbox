package system_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/system"
)

func TestEnvDefault(t *testing.T) {
	os.Clearenv()
	val := system.EnvironOrDefault("TEST", "value")
	if val != "value" {
		t.Error("invalid read")
	}
	os.Setenv("TEST", "  ")
	val = system.EnvironOrDefault("TEST", "value")
	if val != "value" {
		t.Error("invalid read")
	}
	os.Setenv("TEST", " a")
	val = system.EnvironOrDefault("TEST", "value")
	if val != " a" {
		t.Error("invalid read")
	}
}

func TestReadValue(t *testing.T) {
	os.Clearenv()
	val := system.EnvironValue("test")
	if val != system.EmptyValue {
		t.Error("bad read")
	}
	os.Setenv("TEST", "a")
	val = system.EnvironValue("TEST")
	if val != system.UnknownValue {
		t.Error("bad read")
	}
	os.Setenv("TEST", " YeS ")
	val = system.EnvironValue("TEST")
	if val != system.YesValue {
		t.Error("bad read")
	}
	os.Setenv("TEST", " NO ")
	val = system.EnvironValue("TEST")
	if val != system.NoValue {
		t.Error("bad read")
	}
	os.Setenv("TEST", "FALSESSS")
	val = system.EnvironValue("TEST")
	if val != system.UnknownValue {
		t.Error("bad read")
	}
}
