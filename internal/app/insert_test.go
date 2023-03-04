package app_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/backend"
)

type (
	mockInsert struct {
		command *mockCommand
		noTOTP  func() (bool, error)
		input   func(bool, bool) ([]byte, error)
		pipe    func() bool
		token   func() string
	}
)

func newMockInsert(t *testing.T) *mockInsert {
	m := &mockInsert{}
	m.command = newMockCommand(t)
	return m
}

func (m *mockInsert) TOTPToken() string {
	return m.token()
}

func (m *mockInsert) IsPipe() bool {
	return m.pipe()
}

func (m *mockInsert) Input(pipe, multi bool) ([]byte, error) {
	return m.input(pipe, multi)
}

func (m *mockInsert) Args() []string {
	return m.command.Args()
}

func (m *mockInsert) Writer() io.Writer {
	return &m.command.buf
}

func (m *mockInsert) Confirm(p string) bool {
	return m.command.Confirm(p)
}

func (m *mockInsert) IsNoTOTP() (bool, error) {
	return m.noTOTP()
}

func (m *mockInsert) Transaction() *backend.Transaction {
	return m.command.Transaction()
}

func TestInsertArgs(t *testing.T) {
	m := newMockInsert(t)
	m.noTOTP = func() (bool, error) {
		return true, nil
	}
	if _, err := app.ReadArgs(m); err == nil || err.Error() != "insert requires an entry" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.args = []string{"test", "test", "test"}
	if _, err := app.ReadArgs(m); err == nil || err.Error() != "too many arguments" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.args = []string{"test"}
	r, err := app.ReadArgs(m)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if r.Multi || r.Entry != "test" {
		t.Error("invalid parse")
	}
	m.command.args = []string{"-t", "b"}
	if _, err := app.ReadArgs(m); err == nil || err.Error() != "unknown argument" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.args = []string{"-multi", "test3"}
	r, err = app.ReadArgs(m)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !r.Multi || r.Entry != "test3" {
		t.Error("invalid parse")
	}
	m.token = func() string {
		return "test3"
	}
	m.command.args = []string{"-multi", "test/test3"}
	r, err = app.ReadArgs(m)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	m.noTOTP = func() (bool, error) {
		return false, nil
	}
	if _, err := app.ReadArgs(m); err == nil || err.Error() != "can not insert totp entry without totp flag" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.args = []string{"test/test3"}
	if _, err := app.ReadArgs(m); err == nil || err.Error() != "can not insert totp entry without totp flag" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.args = []string{"-totp", "test/test3"}
	r, err = app.ReadArgs(m)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if r.Entry != "test/test3" {
		t.Error("invalid parse")
	}
	m.command.args = []string{"-totp", "test"}
	r, err = app.ReadArgs(m)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if r.Entry != "test/test3" {
		t.Error("invalid parse")
	}
}

func TestInsertDo(t *testing.T) {
	m := newMockInsert(t)
	args := app.InsertArgs{}
	m.pipe = func() bool {
		return false
	}
	args.Entry = "test/test2"
	m.command.confirm = false
	m.input = func(bool, bool) ([]byte, error) {
		return nil, errors.New("failure")
	}
	m.command.buf = bytes.Buffer{}
	if err := args.Do(m); err == nil || err.Error() != "invalid input: failure" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.confirm = false
	m.pipe = func() bool {
		return true
	}
	if err := args.Do(m); err == nil || err.Error() != "invalid input: failure" {
		t.Errorf("invalid error: %v", err)
	}
	m.input = func(bool, bool) ([]byte, error) {
		return []byte("TEST"), nil
	}
	m.command.confirm = true
	args.Entry = "a/b/c"
	if err := args.Do(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() != "" {
		t.Error("invalid insert")
	}
	m.pipe = func() bool {
		return false
	}
	m.command.buf = bytes.Buffer{}
	if err := args.Do(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" {
		t.Error("invalid insert")
	}
	m.command.buf = bytes.Buffer{}
	args.Entry = "test/test2/test1"
	if err := args.Do(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" {
		t.Error("invalid insert")
	}
	m.command.confirm = false
	m.command.buf = bytes.Buffer{}
	args.Entry = "test/test2/test1"
	if err := args.Do(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() != "" {
		t.Error("invalid insert")
	}
}
