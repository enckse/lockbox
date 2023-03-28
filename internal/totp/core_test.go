package totp_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/totp"
)

type (
	mockOptions struct {
		buf bytes.Buffer
		tx  *backend.Transaction
	}
)

func newMock(t *testing.T) *mockOptions {
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test3", "totp"), "5ae472abqdekjqykoyxk7hvc2leklq5n")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "totp"), "5ae472abqdekjqykoyxk7hvc2leklq5n")
	return &mockOptions{
		buf: bytes.Buffer{},
		tx:  fullSetup(t, true),
	}
}

func fullSetup(t *testing.T, keep bool) *backend.Transaction {
	if !keep {
		os.Remove("test.kdbx")
	}
	os.Setenv("LOCKBOX_READONLY", "no")
	os.Setenv("LOCKBOX_STORE", "test.kdbx")
	os.Setenv("LOCKBOX_KEY", "test")
	os.Setenv("LOCKBOX_KEYFILE", "")
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_TOTP", "totp")
	os.Setenv("LOCKBOX_HOOKDIR", "")
	os.Setenv("LOCKBOX_SET_MODTIME", "")
	os.Setenv("LOCKBOX_TOTP_MAX", "1")
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

func setup(t *testing.T) *backend.Transaction {
	return fullSetup(t, false)
}

func TestNewArgumentsErrors(t *testing.T) {
	if _, err := totp.NewArguments(nil, ""); err == nil || err.Error() != "not enough arguments for totp" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := totp.NewArguments([]string{"test"}, ""); err == nil || err.Error() != "invalid token type, not set?" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := totp.NewArguments([]string{"test"}, "a"); err == nil || err.Error() != "unknown totp mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := totp.NewArguments([]string{"ls", "test"}, "a"); err == nil || err.Error() != "list takes no arguments" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := totp.NewArguments([]string{"show"}, "a"); err == nil || err.Error() != "missing entry" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestNewArguments(t *testing.T) {
	args, _ := totp.NewArguments([]string{"ls"}, "test")
	if args.Mode != totp.ListMode || args.Entry != "" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"show", "test"}, "test")
	if args.Mode != totp.ShowMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"clip", "test"}, "test")
	if args.Mode != totp.ClipMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"minimal", "test"}, "test")
	if args.Mode != totp.MinimalMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"once", "test"}, "test")
	if args.Mode != totp.OnceMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"insert", "test2"}, "test")
	if args.Mode != totp.InsertMode || args.Entry != "test2/test" {
		t.Errorf("invalid args: %s", args.Entry)
	}
	args, _ = totp.NewArguments([]string{"insert", "test2/test"}, "test")
	if args.Mode != totp.InsertMode || args.Entry != "test2/test" {
		t.Errorf("invalid args: %s", args.Entry)
	}
}

func TestDoErrors(t *testing.T) {
	args := totp.Arguments{}
	if err := args.Do(totp.Options{}); err == nil || err.Error() != "unknown totp mode" {
		t.Errorf("invalid error: %v", err)
	}
	args.Mode = totp.ShowMode
	if err := args.Do(totp.Options{}); err == nil || err.Error() != "invalid option functions" {
		t.Errorf("invalid error: %v", err)
	}
	opts := totp.Options{}
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

func TestList(t *testing.T) {
	setup(t)
	args, _ := totp.NewArguments([]string{"ls"}, "totp")
	opts := testOptions()
	m := newMock(t)
	opts.App = m
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "test/test2\ntest/test3\n" {
		t.Errorf("invalid list: %s", m.buf.String())
	}
}

func TestNonListError(t *testing.T) {
	setup(t)
	args, _ := totp.NewArguments([]string{"clip", "test"}, "totp")
	opts := testOptions()
	m := newMock(t)
	opts.App = m
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
	setup(t)
	args, _ := totp.NewArguments([]string{"minimal", "test/test3"}, "totp")
	opts := testOptions()
	m := newMock(t)
	opts.App = m
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(m.buf.String()) != 7 {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}

func TestNonInteractive(t *testing.T) {
	setup(t)
	args, _ := totp.NewArguments([]string{"show", "test/test3"}, "totp")
	opts := testOptions()
	m := newMock(t)
	opts.App = m
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
	setup(t)
	args, _ := totp.NewArguments([]string{"once", "test/test3"}, "totp")
	opts := testOptions()
	m := newMock(t)
	opts.App = m
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(strings.Split(m.buf.String(), "\n")) != 5 {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}

func TestShow(t *testing.T) {
	setup(t)
	args, _ := totp.NewArguments([]string{"show", "test/test3"}, "totp")
	m := newMock(t)
	opts := testOptions()
	opts.App = m
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(strings.Split(m.buf.String(), "\n")) < 6 || !strings.Contains(m.buf.String(), "exiting (timeout)") {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}

func testOptions() totp.Options {
	opts := totp.Options{}
	opts.Clear = func() {
	}
	opts.IsNoTOTP = func() (bool, error) {
		return false, nil
	}
	opts.IsInteractive = func() (bool, error) {
		return true, nil
	}
	return opts
}
