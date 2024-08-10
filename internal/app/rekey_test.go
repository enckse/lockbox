package app_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
	"github.com/seanenck/lockbox/internal/backend"
)

type (
	mockKeyer struct {
		data   map[string][]byte
		err    error
		items  map[string]backend.JSON
		rekeys [][]string
	}
)

func (m *mockKeyer) JSON() (map[string]backend.JSON, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func (m *mockKeyer) Insert(entry app.ReKeyEntry) error {
	m.rekeys = append(m.rekeys, entry.Env)
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
	if len(m.rekeys) != 2 {
		t.Errorf("invalid results")
	}
	obj := fmt.Sprintf("%v", m.rekeys)
	for _, idx := range []int{1, 2} {
		if !strings.Contains(obj, fmt.Sprintf("LOCKBOX_SET_MODTIME=%d", idx)) {
			t.Errorf("missing converted modtime: %s", obj)
		}
	}
}

func modTimeKey(t *testing.T, mode string, modSet int) {
	cmd := &mockCommand{}
	cmd.confirm = true
	cmd.buf = bytes.Buffer{}
	cmd.args = []string{"-store", "store", "-key", "abc"}
	m := &mockKeyer{}
	m.items = map[string]backend.JSON{
		"test1": {ModTime: "1"},
		"test2": {ModTime: ""},
	}
	m.data = make(map[string][]byte)
	m.data["test1"] = []byte{1}
	m.data["test2"] = []byte{2}
	cmd.buf = bytes.Buffer{}
	cmd.args = []string{"-store", "store", "-key", "abc", "-modtime", mode}
	if err := app.ReKey(cmd, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(m.rekeys) != 2 {
		t.Errorf("invalid results")
	}
	if strings.Count(fmt.Sprintf("%v", m.rekeys), "LOCKBOX_SET_MODTIME=") != modSet {
		t.Errorf("invalid object: %v", m.rekeys)
	}
}

func TestReKeyModTime(t *testing.T) {
	cmd := &mockCommand{}
	cmd.confirm = true
	cmd.buf = bytes.Buffer{}
	cmd.args = []string{"-store", "store", "-key", "abc"}
	m := &mockKeyer{}
	m.items = map[string]backend.JSON{
		"test1": {ModTime: "1"},
		"test3": {ModTime: "a"},
		"test2": {ModTime: ""},
	}
	m.data = make(map[string][]byte)
	m.data["test1"] = []byte{1}
	m.data["test2"] = []byte{2}
	m.data["test3"] = []byte{4}
	cmd.buf = bytes.Buffer{}
	if err := app.ReKey(cmd, m); err == nil || err.Error() != "did not read modtime" {
		t.Errorf("invalid error: %v", err)
	}
	cmd.args = []string{"-store", "store", "-key", "abc", "-modtime", "xyz"}
	if err := app.ReKey(cmd, m); err == nil || err.Error() != "unknown modtime setting for import: xyz" {
		t.Errorf("invalid error: %v", err)
	}
	modTimeKey(t, "none", 0)
	modTimeKey(t, "skip", 1)
}
