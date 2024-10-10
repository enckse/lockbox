package app_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
	"github.com/seanenck/lockbox/internal/backend"
)

type (
	mockOptions struct {
		buf bytes.Buffer
		tx  *backend.Transaction
	}
)

func newMock(t *testing.T) (*mockOptions, app.TOTPOptions) {
	fullTOTPSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullTOTPSetup(t, true).Insert(backend.NewPath("test", "test3", "totp"), "5ae472abqdekjqykoyxk7hvc2leklq5n")
	fullTOTPSetup(t, true).Insert(backend.NewPath("test", "test2", "totp"), "5ae472abqdekjqykoyxk7hvc2leklq5n")
	m := &mockOptions{
		buf: bytes.Buffer{},
		tx:  fullTOTPSetup(t, true),
	}
	opts := app.NewDefaultTOTPOptions(m)
	opts.Clear = func() {
	}
	opts.IsNoTOTP = func() (bool, error) {
		return false, nil
	}
	opts.IsInteractive = func() (bool, error) {
		return true, nil
	}
	return m, opts
}

func fullTOTPSetup(t *testing.T, keep bool) *backend.Transaction {
	file := testFile()
	if !keep {
		os.Remove(file)
	}
	t.Setenv("LOCKBOX_READONLY", "no")
	t.Setenv("LOCKBOX_STORE", file)
	t.Setenv("LOCKBOX_KEY", "test")
	t.Setenv("LOCKBOX_KEYFILE", "")
	t.Setenv("LOCKBOX_KEYMODE", "plaintext")
	t.Setenv("LOCKBOX_TOTP", "totp")
	t.Setenv("LOCKBOX_HOOKDIR", "")
	t.Setenv("LOCKBOX_SET_MODTIME", "")
	t.Setenv("LOCKBOX_TOTP_MAX", "1")
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	return tr
}

func (m *mockOptions) Confirm(string) bool {
	return true
}

func (m *mockOptions) Args() []string {
	return nil
}

func (m *mockOptions) Transaction() *backend.Transaction {
	return m.tx
}

func (m *mockOptions) Writer() io.Writer {
	return &m.buf
}

func setupTOTP(t *testing.T) *backend.Transaction {
	return fullTOTPSetup(t, false)
}

func TestNewTOTPArgumentsErrors(t *testing.T) {
	if _, err := app.NewTOTPArguments(nil, ""); err == nil || err.Error() != "not enough arguments for totp" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.NewTOTPArguments([]string{"test"}, ""); err == nil || err.Error() != "invalid token type, not set?" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.NewTOTPArguments([]string{"test"}, "a"); err == nil || err.Error() != "unknown totp mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.NewTOTPArguments([]string{"ls", "test"}, "a"); err == nil || err.Error() != "list takes no arguments" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.NewTOTPArguments([]string{"show"}, "a"); err == nil || err.Error() != "invalid arguments" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestNewTOTPArguments(t *testing.T) {
	args, _ := app.NewTOTPArguments([]string{"ls"}, "test")
	if args.Mode != app.ListTOTPMode || args.Entry != "" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"show", "test"}, "test")
	if args.Mode != app.ShowTOTPMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"clip", "test"}, "test")
	if args.Mode != app.ClipTOTPMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"minimal", "test"}, "test")
	if args.Mode != app.MinimalTOTPMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"once", "test"}, "test")
	if args.Mode != app.OnceTOTPMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"insert", "test2"}, "test")
	if args.Mode != app.InsertTOTPMode || args.Entry != "test2/test" {
		t.Errorf("invalid args: %s", args.Entry)
	}
	args, _ = app.NewTOTPArguments([]string{"insert", "test2/test"}, "test")
	if args.Mode != app.InsertTOTPMode || args.Entry != "test2/test" {
		t.Errorf("invalid args: %s", args.Entry)
	}
}

func TestDoErrors(t *testing.T) {
	args := app.TOTPArguments{}
	if err := args.Do(app.TOTPOptions{}); err == nil || err.Error() != "unknown totp mode" {
		t.Errorf("invalid error: %v", err)
	}
	args.Mode = app.ShowTOTPMode
	if err := args.Do(app.TOTPOptions{}); err == nil || err.Error() != "invalid option functions" {
		t.Errorf("invalid error: %v", err)
	}
	opts := app.TOTPOptions{}
	opts.Clear = func() {
	}
	if err := args.Do(opts); err == nil || err.Error() != "invalid option functions" {
		t.Errorf("invalid error: %v", err)
	}
	opts.IsNoTOTP = func() (bool, error) {
		return true, nil
	}
	if err := args.Do(opts); err == nil || err.Error() != "invalid option functions" {
		t.Errorf("invalid error: %v", err)
	}
	opts.IsInteractive = func() (bool, error) {
		return false, nil
	}
	if err := args.Do(opts); err == nil || err.Error() != "totp is disabled" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestTOTPList(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"ls"}, "totp")
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "test/test2\ntest/test3\n" {
		t.Errorf("invalid list: %s", m.buf.String())
	}
}

func TestNonListError(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"clip", "test"}, "totp")
	_, opts := newMock(t)
	opts.IsInteractive = func() (bool, error) {
		return false, nil
	}
	if err := args.Do(opts); err == nil || err.Error() != "clipboard not available in non-interactive mode" {
		t.Errorf("invalid error: %v", err)
	}
	opts.IsInteractive = func() (bool, error) {
		return true, nil
	}
	if err := args.Do(opts); err == nil || err.Error() != "object does not exist" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMinimal(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"minimal", "test/test3"}, "totp")
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(m.buf.String()) != 7 {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}

func TestNonInteractive(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"show", "test/test3"}, "totp")
	m, opts := newMock(t)
	opts.IsInteractive = func() (bool, error) {
		return false, nil
	}
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(m.buf.String()) != 7 {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}

func TestOnce(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"once", "test/test3"}, "totp")
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(strings.Split(m.buf.String(), "\n")) != 5 {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}

func TestShow(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"show", "test/test3"}, "totp")
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(strings.Split(m.buf.String(), "\n")) < 6 || !strings.Contains(m.buf.String(), "exiting (timeout)") {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}
