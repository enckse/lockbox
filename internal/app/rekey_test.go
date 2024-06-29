package app_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
	"github.com/seanenck/lockbox/internal/backend"
)

type (
	mockKeyer struct {
		data   map[string][]byte
		err    error
		rekeys int
		items  map[string]backend.JSON
	}
)

func (m *mockKeyer) JSON() (map[string]backend.JSON, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func (m *mockKeyer) Insert(entry app.ReKeyEntry) error {
	m.rekeys++
	if entry.Path == "error" {
		return errors.New("bad insert")
	}
	return nil
}

func TestErrors(t *testing.T) {
	cmd := &mockCommand{}
	cmd.confirm = false
	cmd.buf = bytes.Buffer{}
	m := &mockKeyer{}
	cmd.args = []string{"-store", "store", "-key", "abc"}
	if err := app.ReKey(cmd, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	cmd.confirm = true
	m.err = errors.New("invalid call")
	if err := app.ReKey(cmd, m); err == nil || err.Error() != "invalid call" {
		t.Errorf("invalid error: %v", err)
	}
	m.err = nil
	m.items = map[string]backend.JSON{"test": {ModTime: ""}}
	if err := app.ReKey(cmd, m); err == nil || err.Error() != "did not read modtime" {
		t.Errorf("invalid error: %v", err)
	}
	m.data = make(map[string][]byte)
	m.data["test1"] = []byte{1}
	m.data["error"] = []byte{2}
	m.items = map[string]backend.JSON{"error": {ModTime: "2"}}
	if err := app.ReKey(cmd, m); err == nil || err.Error() != "bad insert" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReKey(t *testing.T) {
	cmd := &mockCommand{}
	cmd.confirm = true
	cmd.buf = bytes.Buffer{}
	cmd.args = []string{"-store", "store", "-key", "abc"}
	if err := app.ReKey(cmd, &mockKeyer{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if cmd.buf.String() != "" {
		t.Error("no data")
	}
	m := &mockKeyer{}
	m.items = map[string]backend.JSON{
		"test1": {ModTime: "1"},
		"test2": {ModTime: "2"},
	}
	m.data = make(map[string][]byte)
	m.data["test1"] = []byte{1}
	m.data["test2"] = []byte{2}
	cmd.buf = bytes.Buffer{}
	if err := app.ReKey(cmd, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if cmd.buf.String() == "" {
		t.Error("invalid data")
	}
	if m.rekeys != 2 {
		t.Errorf("invalid results")
	}
}
