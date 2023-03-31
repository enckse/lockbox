package app_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/backend"
)

type (
	mockKeyer struct {
		data   map[string][]byte
		err    error
		rekeys []app.ReKeyEntry
		items  []backend.JSON
	}
)

func (m *mockKeyer) JSON() ([]backend.JSON, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func (m *mockKeyer) Show(entry string) ([]byte, error) {
	val, ok := m.data[entry]
	if !ok {
		return nil, errors.New("no data")
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
	m.err = errors.New("invalid call")
	if err := app.ReKey(&buf, m); err == nil || err.Error() != "invalid call" {
		t.Errorf("invalid error: %v", err)
	}
	m.err = nil
	m.items = []backend.JSON{{Path: "test1", ModTime: ""}}
	if err := app.ReKey(&buf, m); err == nil || err.Error() != "did not read modtime" {
		t.Errorf("invalid error: %v", err)
	}
	m.items = []backend.JSON{{Path: "test1", ModTime: "2"}}
	if err := app.ReKey(&buf, m); err == nil || err.Error() != "no data" {
		t.Errorf("invalid error: %v", err)
	}
	m.data = make(map[string][]byte)
	m.data["test1"] = []byte{1}
	m.data["error"] = []byte{2}
	m.items = []backend.JSON{{Path: "error", ModTime: "2"}}
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
	m.items = []backend.JSON{{Path: "test1", ModTime: "2"}, {Path: "test1", ModTime: "2"}}
	m.data = make(map[string][]byte)
	m.data["test1"] = []byte{1}
	m.data["test2"] = []byte{2}
	if err := app.ReKey(&buf, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("invalid data")
	}
	if len(m.rekeys) != 2 {
		t.Error("invalid rekeys")
	}
	if fmt.Sprintf("%v", m.rekeys) != `[{test1 [LOCKBOX_KEYMODE= LOCKBOX_KEY=abc LOCKBOX_KEYFILE= LOCKBOX_STORE=store LOCKBOX_SET_MODTIME=2] [1]} {test1 [LOCKBOX_KEYMODE= LOCKBOX_KEY=abc LOCKBOX_KEYFILE= LOCKBOX_STORE=store LOCKBOX_SET_MODTIME=2] [1]}]` {
		t.Errorf("invalid results: %v", m.rekeys)
	}
}
