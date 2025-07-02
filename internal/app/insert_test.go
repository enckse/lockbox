package app_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
	"git.sr.ht/~enckse/lockbox/internal/kdbx"
)

type (
	mockInsert struct {
		command     *mockCommand
		noTOTP      func() (bool, error)
		input       func() ([]byte, error)
		pipe        func() bool
		token       func() string
		prompt      string
		isPass      bool
		interactive bool
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

func (m *mockInsert) Input(interactive, isPass bool, prompt string) ([]byte, error) {
	m.interactive = interactive
	m.prompt = prompt
	m.isPass = isPass
	return m.input()
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

func (m *mockInsert) Transaction() *kdbx.Transaction {
	return m.command.Transaction()
}

func TestInsertDo(t *testing.T) {
	m := newMockInsert(t)
	m.pipe = func() bool {
		return false
	}
	m.command.args = []string{"test/test2/test3ss/password"}
	m.command.confirm = false
	m.input = func() ([]byte, error) {
		return nil, errors.New("failure")
	}
	m.command.buf = bytes.Buffer{}
	if err := app.Insert(m); err == nil || err.Error() != "invalid input: failure" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.confirm = false
	m.command.args = []string{"test/test2/test3/password"}
	m.pipe = func() bool {
		return true
	}
	if err := app.Insert(m); err == nil || err.Error() != "invalid input: failure" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.confirm = false
	m.command.args = []string{"test/test2/test3/Password"}
	m.pipe = func() bool {
		return true
	}
	if err := app.Insert(m); err == nil || err.Error() != "'Password' is not an allowed field name" {
		t.Errorf("invalid error: %v", err)
	}
	m.input = func() ([]byte, error) {
		return []byte("TEST"), nil
	}
	m.command.confirm = true
	m.command.args = []string{"a/b/password"}
	if err := app.Insert(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() != "" {
		t.Error("invalid insert")
	}
	m.pipe = func() bool {
		return false
	}
	m.command.buf = bytes.Buffer{}
	if err := app.Insert(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" {
		t.Error("invalid insert")
	}
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1/password"}
	if err := app.Insert(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" {
		t.Error("invalid insert")
	}
	if m.prompt != "password" || !m.isPass {
		t.Error("invalid field prompt")
	}
	m.command.confirm = false
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1/password"}
	if err := app.Insert(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() != "" {
		t.Error("invalid insert")
	}
	m.interactive = false
	m.command.confirm = true
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1/password"}
	if err := app.Insert(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" || !m.interactive {
		t.Error("invalid insert")
	}
	m.interactive = false
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1/notes"}
	if err := app.Insert(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" || m.interactive {
		t.Errorf("invalid insert %s %v", m.command.buf.String(), m.interactive)
	}
	if m.prompt != "notes" || !m.isPass {
		t.Error("invalid field prompt")
	}
	m.interactive = false
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1/url"}
	if err := app.Insert(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() != "" || !m.interactive {
		t.Errorf("invalid insert %s %v", m.command.buf.String(), m.interactive)
	}
	if m.prompt != "url" || m.isPass {
		t.Error("invalid field prompt")
	}
}

func TestInsertTOTP(t *testing.T) {
	defer store.Clear()
	m := newMockInsert(t)
	m.pipe = func() bool {
		return false
	}
	m.input = func() ([]byte, error) {
		return []byte("t"), nil
	}
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1/otp"}
	if err := app.Insert(m); err == nil || err.Error() != "Decoding of secret as base32 failed." {
		t.Errorf("invalid error: %v", err)
	}
	store.SetBool("LOCKBOX_TOTP_CHECK_ON_INSERT", false)
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1/otp"}
	if err := app.Insert(m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
