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
		pass       string
		secondPass string
		confirm    bool
		args       []string
		buf        bytes.Buffer
		t          *testing.T
		pipe       bool
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

func (m *mockKeyer) ReadLine() (string, error) {
	return m.Password()
}

func (m *mockKeyer) Password() (string, error) {
	p := m.pass
	m.pass = m.secondPass
	m.secondPass = ""
	return p, nil
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
	if err := app.ReKey(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	mock.confirm = true
	if err := app.ReKey(mock); err == nil || err.Error() != "password required but not given" {
		t.Errorf("invalid error: %v", err)
	}
	mock.pass = "abc"
	if err := app.ReKey(mock); err == nil || err.Error() != "rekey passwords do not match" {
		t.Errorf("invalid error: %v", err)
	}
	mock.pass = "xyz"
	mock.secondPass = "xyz"
	if err := app.ReKey(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReKeyPipe(t *testing.T) {
	newMockCommand(t)
	mock := &mockKeyer{}
	mock.pipe = true
	if err := app.ReKey(mock); err == nil || err.Error() != "password required but not given" {
		t.Errorf("invalid error: %v", err)
	}
	mock.pass = "abc"
	if err := app.ReKey(mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
