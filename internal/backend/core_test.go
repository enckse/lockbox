package backend_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/backend"
)

func TestLoad(t *testing.T) {
	if _, err := backend.Load("garbage"); err.Error() != "should use a .kdbx extension" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := backend.Load("garbage.kdbx"); err.Error() != "invalid file, does not exist" {
		t.Errorf("invalid error: %v", err)
	}
}
