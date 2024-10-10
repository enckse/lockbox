package app_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
	"github.com/seanenck/lockbox/internal/backend"
)

type (
	mockKeyer struct {
		pass    string
		confirm bool
		args    []string
		buf     bytes.Buffer
		t       *testing.T
		pipe    bool
	}
)

func (m *mockKeyer) Confirm(string) bool {
	return m.confirm
}

func (m *mockKeyer) Transaction() *backend.Transaction {
	return fullSetup(m.t, true)
}

func (m *mockKeyer) Args() []string {
	return m.args
}

func (m *mockKeyer) Input(bool) ([]byte, error) {
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
