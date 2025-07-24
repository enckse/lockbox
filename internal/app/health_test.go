package app_test

import (
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/config/store"
)

func TestHealth(t *testing.T) {
	m := newMockCommand(t)
	database, _ := store.GetString("LOCKBOX_STORE")
	store.Clear()
	store.SetBool("LOCKBOX_FEATURE_CLIP", false)
	if err := app.Health(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s := m.buf.String()
	if strings.Count(s, "ok") != 1 || !strings.Contains(s, "key MUST be set in this key mode") || !strings.Contains(s, "clip feature is disabled") || !strings.Contains(s, "store not set") {
		t.Errorf("invalid health: %s", s)
	}
	m.buf.Reset()
	store.SetBool("LOCKBOX_FEATURE_CLIP", true)
	store.SetArray("LOCKBOX_CLIP_COPY", []string{"x"})
	if err := app.Health(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s = m.buf.String()
	if strings.Count(s, "ok") != 2 || !strings.Contains(s, "key MUST be set in this key mode") || !strings.Contains(s, "store not set") {
		t.Errorf("invalid health: %s", s)
	}
	m.buf.Reset()
	store.SetString("LOCKBOX_STORE", "xxxxxx")
	if err := app.Health(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s = m.buf.String()
	if strings.Count(s, "ok") != 2 || !strings.Contains(s, "key MUST be set in this key mode") || !strings.Contains(s, "store does not exist") {
		t.Errorf("invalid health: %s", s)
	}
	m.buf.Reset()
	store.SetString("LOCKBOX_STORE", database)
	if err := app.Health(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s = m.buf.String()
	if strings.Count(s, "ok") != 3 || !strings.Contains(s, "key MUST be set in this key mode") {
		t.Errorf("invalid health: %s", s)
	}
	m.buf.Reset()
	store.SetString("LOCKBOX_CREDENTIALS_KEY_FILE", "xxxxx")
	if err := app.Health(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s = m.buf.String()
	if strings.Count(s, "ok") != 2 || !strings.Contains(s, "key MUST be set in this key mode") || !strings.Contains(s, "key file set, does not exist") {
		t.Errorf("invalid health: %s", s)
	}
	m.buf.Reset()
	store.SetString("LOCKBOX_CREDENTIALS_KEY_FILE", database)
	if err := app.Health(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s = m.buf.String()
	if strings.Count(s, "ok") != 3 || !strings.Contains(s, "key MUST be set in this key mode") {
		t.Errorf("invalid health: %s", s)
	}
	m.buf.Reset()
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "command")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"zoiajfoaijfoeaijo1j091"})
	if err := app.Health(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s = m.buf.String()
	if strings.Count(s, "ok") != 3 || !strings.Contains(s, "key command failed") {
		t.Errorf("invalid health: %s", s)
	}
	m.buf.Reset()
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"pass"})
	if err := app.Health(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	s = m.buf.String()
	if strings.Count(s, "ok") != 4 {
		t.Errorf("invalid health: %s", s)
	}
}
