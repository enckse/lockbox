package encrypt_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/store"
)

func setupData(t *testing.T) string {
	os.Setenv("LOCKBOX_KEYMODE", "")
	os.Setenv("LOCKBOX_KEY", "")
	if store.PathExists("bin") {
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
	e, err := encrypt.NewLockbox(encrypt.LockboxOptions{Key: "echo test", KeyMode: inputs.CommandKeyMode, File: setupData(t)})
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
	_, err := encrypt.NewLockbox(encrypt.LockboxOptions{})
	if err == nil || err.Error() != "no key given" {
		t.Errorf("invalid error: %v", err)
	}
	_, err = encrypt.NewLockbox(encrypt.LockboxOptions{KeyMode: inputs.CommandKeyMode, Key: "echo"})
	if err == nil || err.Error() != "key is empty" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestKeyLength(t *testing.T) {
	val := ""
	for i := 0; i < 42; i++ {
		val = fmt.Sprintf("a%s", val)
		_, err := encrypt.NewLockbox(encrypt.LockboxOptions{KeyMode: inputs.PlainKeyMode, Key: val})
		if err != nil {
			t.Error("no error expected")
		}
	}
}

func TestUnknownMode(t *testing.T) {
	_, err := encrypt.NewLockbox(encrypt.LockboxOptions{KeyMode: "aaa", Key: "echo"})
	if err == nil || err.Error() != "unknown keymode" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestEncryptDecryptPlainText(t *testing.T) {
	e, err := encrypt.NewLockbox(encrypt.LockboxOptions{Key: "plain", KeyMode: inputs.PlainKeyMode, File: setupData(t)})
	if err != nil {
		t.Errorf("failed to create lockbox: %v", err)
	}
	data := []byte("datum")
	if err := e.Encrypt(data); err != nil {
		t.Errorf("failed to encrypt: %v", err)
	}
	d, err := e.Decrypt()
	if err != nil {
		t.Errorf("failed to decrypt: %v", err)
	}
	if string(d) != string(data) {
		t.Error("data mismatch")
	}
}

func TestEncryptDecryptSecretBox(t *testing.T) {
	e, err := encrypt.NewLockbox(encrypt.LockboxOptions{Key: "plain", KeyMode: inputs.PlainKeyMode, File: setupData(t), Algorithm: "secretbox"})
	if err != nil {
		t.Errorf("failed to create lockbox: %v", err)
	}
	data := []byte("datum")
	if err := e.Encrypt(data); err != nil {
		t.Errorf("failed to encrypt: %v", err)
	}
	d, err := e.Decrypt()
	if err != nil {
		t.Errorf("failed to decrypt: %v", err)
	}
	if string(d) != string(data) {
		t.Error("data mismatch")
	}
}

func TestEncryptDecryptAESBox(t *testing.T) {
	e, err := encrypt.NewLockbox(encrypt.LockboxOptions{Key: "plain", KeyMode: inputs.PlainKeyMode, File: setupData(t), Algorithm: "aesgcm"})
	if err != nil {
		t.Errorf("failed to create lockbox: %v", err)
	}
	data := []byte("datum")
	if err := e.Encrypt(data); err != nil {
		t.Errorf("failed to encrypt: %v", err)
	}
	d, err := e.Decrypt()
	if err != nil {
		t.Errorf("failed to decrypt: %v", err)
	}
	if string(d) != string(data) {
		t.Error("data mismatch")
	}
}
