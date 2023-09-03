package config_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/config"
)

func TestGetKey(t *testing.T) {
	os.Setenv("LOCKBOX_KEY", "aaa")
	os.Setenv("LOCKBOX_KEYMODE", "lak;jfea")
	if k, err := config.GetKey(false); err.Error() != "unknown keymode" || k != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_KEY", "")
	if k, err := config.GetKey(false); err != nil || k != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_KEY", "key")
	k, err := config.GetKey(false)
	if err != nil || k == nil || string(k.Key()) != "key" || k.Interactive() {
		t.Error("invalid key retrieval")
	}
	os.Setenv("LOCKBOX_KEY", "key")
	k, err = config.GetKey(true)
	if err != nil || k == nil || len(k.Key()) != 1 || k.Key()[0] != 0 || k.Interactive() {
		t.Error("invalid key retrieval")
	}
	os.Setenv("LOCKBOX_KEYMODE", "command")
	os.Setenv("LOCKBOX_KEY", "invalid command text is long and invalid via shlex")
	if k, err := config.GetKey(false); err == nil || k != nil {
		t.Error("should have failed")
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_KEYMODE", "ask")
	os.Setenv("LOCKBOX_KEY", "")
	if k, err := config.GetKey(false); err != nil || k == nil || !k.Interactive() {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_KEYMODE", "ask")
	os.Setenv("LOCKBOX_KEY", "")
	if k, err := config.GetKey(true); err != nil || k == nil || !k.Interactive() {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "no")
	os.Setenv("LOCKBOX_KEYMODE", "ask")
	os.Setenv("LOCKBOX_KEY", "")
	if k, err := config.GetKey(false); err == nil || err.Error() != "ask key mode requested in non-interactive mode" || k != nil {
		t.Errorf("invalid error: %v", err)
	}
	os.Setenv("LOCKBOX_INTERACTIVE", "yes")
	os.Setenv("LOCKBOX_KEYMODE", "ask")
	os.Setenv("LOCKBOX_KEY", "aaa")
	if k, err := config.GetKey(false); err == nil || err.Error() != "key can NOT be set in ask key mode" || k != nil {
		t.Errorf("invalid error: %v", err)
	}
}
