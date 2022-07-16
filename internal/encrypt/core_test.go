package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/enckse/lockbox/internal/misc"
)

func setupData(t *testing.T) string {
	os.Setenv("LOCKBOX_KEYMODE", "")
	os.Setenv("LOCKBOX_KEY", "")
	if misc.PathExists("bin") {
		if err := os.RemoveAll("bin"); err != nil {
			t.Errorf("unable to cleanup dir: %v", err)
		}
	}

	if err := os.MkdirAll("bin", 0755); err != nil {
		t.Errorf("failed to setup bin directory: %v", err)
	}
	return filepath.Join("bin", "test.lb")
}

func TestEncryptDecryptCommand(t *testing.T) {
	e, err := NewLockbox(LockboxOptions{Key: "echo test", KeyMode: CommandKeyMode, File: setupData(t)})
	if err != nil {
		t.Errorf("failed to create lockbox: %v", err)
	}
	data := []byte("datum")
	if err := e.Encrypt(data); err != nil {
		t.Errorf("failed to encrypt: %v", err)
	}
	d, err := e.Decrypt()
	if err != nil {
		t.Errorf("failed to encrypt: %v", err)
	}
	if string(d) != string(data) {
		t.Error("data mismatch")
	}
}

func TestEmptyKey(t *testing.T) {
	setupData(t)
	_, err := NewLockbox(LockboxOptions{})
	if err == nil || err.Error() != "no key given" {
		t.Errorf("invalid error: %v", err)
	}
	_, err = NewLockbox(LockboxOptions{KeyMode: CommandKeyMode, Key: "echo"})
	if err == nil || err.Error() != "key is empty" {
		t.Errorf("invalid error: %v", err)
	}
	_, err = NewLockbox(LockboxOptions{KeyMode: CommandKeyMode, Key: "echo aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
	if err == nil || err.Error() != "key is too large for use" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestUnknownMode(t *testing.T) {
	_, err := NewLockbox(LockboxOptions{KeyMode: "aaa", Key: "echo"})
	if err == nil || err.Error() != "unknown keymode" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestEncryptDecryptPlainText(t *testing.T) {
	e, err := NewLockbox(LockboxOptions{Key: "plain", KeyMode: PlainKeyMode, File: setupData(t)})
	if err != nil {
		t.Errorf("failed to create lockbox: %v", err)
	}
	data := []byte("datum")
	if err := e.Encrypt(data); err != nil {
		t.Errorf("failed to encrypt: %v", err)
	}
	d, err := e.Decrypt()
	if err != nil {
		t.Errorf("failed to encrypt: %v", err)
	}
	if string(d) != string(data) {
		t.Error("data mismatch")
	}
}
