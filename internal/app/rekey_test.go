package app_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

type (
	mockKeyer struct {
		entries []string
		data    map[string][]byte
		stats   map[string][]byte
		err     error
		rekeys  []app.ReKeyEntry
	}
)

func (m *mockKeyer) List() ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.entries, nil
}

func (m *mockKeyer) Show(entry string) ([]byte, error) {
	val, ok := m.data[entry]
	if !ok {
		return nil, errors.New("no data")
	}
	return val, nil
}

func (m *mockKeyer) Stats(entry string) ([]byte, error) {
	val, ok := m.stats[entry]
	if !ok {
		return nil, errors.New("no stats")
	}
	return val, nil
}

func (m *mockKeyer) Insert(entry app.ReKeyEntry) error {
	m.rekeys = append(m.rekeys, entry)
	if entry.Path == "error" {
		return errors.New("bad insert")
	}
	return nil
}

func setupReKey() {
	os.Setenv("LOCKBOX_KEY_NEW", "abc")
	os.Setenv("LOCKBOX_STORE_NEW", "store")
}

func TestErrors(t *testing.T) {
	setupReKey()
	var buf bytes.Buffer
	m := &mockKeyer{}
	m.err = errors.New("invalid ls")
	if err := app.ReKey(&buf, m); err == nil || err.Error() != "invalid ls" {
		t.Errorf("invalid error: %v", err)
	}
	m.err = nil
	m.entries = []string{"test1", "error"}
	if err := app.ReKey(&buf, m); err == nil || err.Error() != "failed to get modtime: no stats" {
		t.Errorf("invalid error: %v", err)
	}
	m.stats = make(map[string][]byte)
	m.stats["test1"] = []byte("modtime")
	if err := app.ReKey(&buf, m); err == nil || !strings.HasPrefix(err.Error(), "invalid stats json:") {
		t.Errorf("invalid error: %v", err)
	}
	m.stats = make(map[string][]byte)
	m.stats["test1"] = []byte("{\"modtime\": \"\"}")
	if err := app.ReKey(&buf, m); err == nil || err.Error() != "did not read modtime" {
		t.Errorf("invalid error: %v", err)
	}
	m.stats["test1"] = []byte("{\"modtime\": \"1\"}")
	m.stats["error"] = []byte("{\"modtime\": \"1\"}")
	if err := app.ReKey(&buf, m); err == nil || err.Error() != "no data" {
		t.Errorf("invalid error: %v", err)
	}
	m.data = make(map[string][]byte)
	m.data["test1"] = []byte{1}
	m.data["error"] = []byte{2}
	if err := app.ReKey(&buf, m); err == nil || err.Error() != "bad insert" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReKey(t *testing.T) {
	setupReKey()
	var buf bytes.Buffer
	if err := app.ReKey(&buf, &mockKeyer{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() != "" {
		t.Error("no data")
	}
	m := &mockKeyer{}
	m.entries = []string{"test1", "test2"}
	m.data = make(map[string][]byte)
	m.data["test1"] = []byte{1}
	m.data["test2"] = []byte{2}
	m.stats = make(map[string][]byte)
	m.stats["test1"] = []byte("{\"modtime\": \"1\"}")
	m.stats["test2"] = []byte("{\"modtime\": \"2\"}")
	if err := app.ReKey(&buf, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("invalid data")
	}
	if len(m.rekeys) != 2 {
		t.Error("invalid rekeys")
	}
	if fmt.Sprintf("%v", m.rekeys) != `[{test1 [LOCKBOX_KEYMODE= LOCKBOX_KEY=abc LOCKBOX_KEYFILE= LOCKBOX_STORE=store LOCKBOX_SET_MODTIME=1] [1]} {test2 [LOCKBOX_KEYMODE= LOCKBOX_KEY=abc LOCKBOX_KEYFILE= LOCKBOX_STORE=store LOCKBOX_SET_MODTIME=2] [2]}]` {
		t.Errorf("invalid results: %v", m.rekeys)
	}
}
