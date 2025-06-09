package config_test

import (
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

func TestDefaultKey(t *testing.T) {
	store.Clear()
	if _, err := config.NewKey(config.DefaultKeyMode); err == nil || err.Error() != "key MUST be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
	store.Clear()
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	if _, err := config.NewKey(config.IgnoreKeyMode); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestNewKeyErrors(t *testing.T) {
	store.Clear()
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "invalid")
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "unknown key mode: invalid" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "none")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "key can NOT be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "command")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{})
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "key MUST be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"  "})
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "key MUST be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestIgnoreKey(t *testing.T) {
	store.Clear()
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "ignore")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	if _, err := config.NewKey(config.IgnoreKeyMode); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "ignore")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{})
	if _, err := config.NewKey(config.IgnoreKeyMode); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReadErrors(t *testing.T) {
	store.Clear()
	k := config.Key{}
	if _, err := k.Read(); err == nil || err.Error() != "invalid key given" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestPlainKey(t *testing.T) {
	store.Clear()
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"  test "})
	k, err := config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	val, err := k.Read()
	if err != nil || val != "test" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReadIgnoreOrNoKey(t *testing.T) {
	store.Clear()
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "ignore")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	k, err := config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	val, err := k.Read()
	if err != nil || val != "" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "ignore")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{})
	k, err = config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	val, err = k.Read()
	if err != nil || val != "" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "none")
	k, err = config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	val, err = k.Read()
	if err != nil || val != "" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestCommandKey(t *testing.T) {
	store.Clear()
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "command")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"thisisagarbagekey"})
	k, err := config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	_, err = k.Read()
	if err == nil || !strings.HasPrefix(err.Error(), "key command failed:") {
		t.Errorf("invalid error: %v", err)
	}
}
