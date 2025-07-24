package app_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/kdbx"
)

type (
	mockKeyer struct {
		t       *testing.T
		pass    string
		args    []string
		buf     bytes.Buffer
		confirm bool
		pipe    bool
	}
)

func (m *mockKeyer) Confirm(string) bool {
	return m.confirm
}

func (m *mockKeyer) Transaction() *kdbx.Transaction {
	return fullSetup(m.t, true)
}

func (m *mockKeyer) Args() []string {
	return m.args
}

func (m *mockKeyer) Input(_, pass bool, _ string) ([]byte, error) {
	if !pass {
		return nil, errors.New("invalid request, always password")
	}
	return []byte(m.pass), nil
}

func (m *mockKeyer) IsPipe() bool {
	return m.pipe
}

func (m *mockKeyer) Writer() io.Writer {
	return &m.buf
}

func TestReKey(t *testing.T) {
	newMockCommand(t)
	mock := &mockKeyer{}
	mock.t = t
	if err := app.ReKey(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	mock.confirm = true
	mock.pipe = false
	if err := app.ReKey(mock); err == nil || err.Error() != "key and/or keyfile must be set" {
		t.Errorf("invalid error: %v", err)
	}
	mock.pass = "abc"
	if err := app.ReKey(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReKeyPipe(t *testing.T) {
	newMockCommand(t)
	mock := &mockKeyer{}
	mock.t = t
	mock.pipe = true
	if err := app.ReKey(mock); err == nil || err.Error() != "key and/or keyfile must be set" {
		t.Errorf("invalid error: %v", err)
	}
	mock.pass = "abc"
	if err := app.ReKey(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReKeyFlags(t *testing.T) {
	newMockCommand(t)
	mock := &mockKeyer{}
	mock.t = t
	mock.args = []string{"-nokey"}
	if err := app.ReKey(mock); err == nil || err.Error() != "a key or keyfile must be passed for rekey" {
		t.Errorf("invalid error: %v", err)
	}
	mock.args = []string{"-nokey", "-keyfile", "blla"}
	mock.confirm = true
	mock.pipe = false
	if err := app.ReKey(mock); err == nil || err.Error() != "no keyfile found on disk" {
		t.Errorf("invalid error: %v", err)
	}
}
