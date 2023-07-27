package inputs_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/inputs"
)

func TestPlatformSet(t *testing.T) {
	if len(inputs.PlatformSet()) != 4 {
		t.Error("invalid platform set")
	}
}

func TestSet(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	inputs.EnvStore.Set("TEST")
	if inputs.EnvStore.Get() != "TEST" {
		t.Errorf("invalid set/get")
	}
}

func TestKeyValue(t *testing.T) {
	val := inputs.EnvStore.KeyValue("TEST")
	if val != "LOCKBOX_STORE=TEST" {
		t.Errorf("invalid keyvalue")
	}
}
