package app_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
	"git.sr.ht/~enckse/lockbox/internal/kdbx"
)

type (
	mockOptions struct {
		tx  *kdbx.Transaction
		buf bytes.Buffer
	}
)

func newMock(t *testing.T) (*mockOptions, app.TOTPOptions) {
	fullTOTPSetup(t, true).Insert(kdbx.NewPath("test", "test3", "totp"), map[string]string{"password": "pass", "otp": "5ae472abqdekjqykoyxk7hvc2leklq5n"})
	fullTOTPSetup(t, true).Insert(kdbx.NewPath("test", "test2", "totp"), map[string]string{"password": "pass", "otp": "5ae472abqdekjqykoyxk7hvc2leklq5n"})
	m := &mockOptions{
		buf: bytes.Buffer{},
		tx:  fullTOTPSetup(t, true),
	}
	opts := app.NewDefaultTOTPOptions(m)
	opts.Clear = func() {
	}
	return m, opts
}

func fullTOTPSetup(t *testing.T, keep bool) *kdbx.Transaction {
	store.Clear()
	file := testFile()
	if !keep {
		os.Remove(file)
	}
	store.SetString("LOCKBOX_STORE", file)
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetInt64("LOCKBOX_TOTP_TIMEOUT", 1)
	tr, err := kdbx.NewTransaction()
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

func (m *mockOptions) Transaction() *kdbx.Transaction {
	return m.tx
}

func (m *mockOptions) Writer() io.Writer {
	return &m.buf
}

func setupTOTP(t *testing.T) *kdbx.Transaction {
	return fullTOTPSetup(t, false)
}

func TestNewTOTPArgumentsErrors(t *testing.T) {
	if _, err := app.NewTOTPArguments(nil); err == nil || err.Error() != "not enough arguments for totp" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.NewTOTPArguments([]string{"test"}); err == nil || err.Error() != "unknown totp mode" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.NewTOTPArguments([]string{"ls", "test", "xxx"}); err == nil || err.Error() != "list takes only a filter (if any)" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.NewTOTPArguments([]string{"show"}); err == nil || err.Error() != "invalid arguments" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestNewTOTPArguments(t *testing.T) {
	args, _ := app.NewTOTPArguments([]string{"ls"})
	if args.Mode != "ls" || args.Entry != "" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"ls", "xyz"})
	if args.Mode != "ls" || args.Entry != "xyz" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"show", "test"})
	if args.Mode != "show" || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"clip", "test"})
	if args.Mode != "clip" || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"minimal", "test"})
	if args.Mode != "minimal" || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"once", "test"})
	if args.Mode != "once" || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"url", "test"})
	if args.Mode != "url" || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = app.NewTOTPArguments([]string{"seed", "test"})
	if args.Mode != "seed" || args.Entry != "test" {
		t.Error("invalid args")
	}
}

func TestDoErrors(t *testing.T) {
	args := app.TOTPArguments{}
	if err := args.Do(app.TOTPOptions{}); err == nil || err.Error() != "unknown totp mode" {
		t.Errorf("invalid error: %v", err)
	}
	args.Mode = "show"
	if err := args.Do(app.TOTPOptions{}); err == nil || err.Error() != "invalid option functions" {
		t.Errorf("invalid error: %v", err)
	}
	opts := app.TOTPOptions{}
	opts.Clear = func() {
	}
	if err := args.Do(opts); err == nil || err.Error() != "'' is not a TOTP entry" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestTOTPList(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"ls"})
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "test/test2/totp/otp\ntest/test3/totp/otp\n" {
		t.Errorf("invalid list: %s", m.buf.String())
	}
}

func TestNonListError(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"show", "test/test3"})
	_, opts := newMock(t)
	if err := args.Do(opts); err == nil || err.Error() != "'test/test3' is not a TOTP entry" {
		t.Errorf("invalid error: %v", err)
	}
	args, _ = app.NewTOTPArguments([]string{"clip", "test/test3/otp"})
	_, opts = newMock(t)
	if err := args.Do(opts); err == nil || err.Error() != "entry does not exist" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestSeed(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"seed", "test/test3/totp/otp"})
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "5ae472abqdekjqykoyxk7hvc2leklq5n\n" {
		t.Errorf("invalid seed: %s", m.buf.String())
	}
}

func TestURL(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"url", "test/test3/totp/otp"})
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !strings.Contains(m.buf.String(), "url") {
		t.Errorf("invalid url dump: %s", m.buf.String())
	}
}

func TestMinimal(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"minimal", "test/test3/totp/otp"})
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(m.buf.String()) != 7 {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}

func TestOnce(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"once", "test/test3/totp/otp"})
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
	args, _ := app.NewTOTPArguments([]string{"show", "test/test3/totp/otp"})
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(strings.Split(m.buf.String(), "\n")) < 6 || !strings.Contains(m.buf.String(), "exiting (timeout)") {
		t.Errorf("invalid short: %s", m.buf.String())
	}
}

func TestParseWindows(t *testing.T) {
	if _, err := app.ParseTimeWindow(); err.Error() != "invalid colorization rules for totp, none found" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow(" ", "2"); err.Error() != "invalid colorization rule found: 2" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow(" 1:200"); err.Error() != "invalid time found for colorization rule: 1:200" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow(" 1:-1"); err.Error() != "invalid time found for colorization rule: 1:-1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow(" 200:1"); err.Error() != "invalid time found for colorization rule: 200:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow(" -1:1"); err.Error() != "invalid time found for colorization rule: -1:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow(" 2:1"); err.Error() != "invalid time found for colorization rule: 2:1" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow("xxx:1"); err.Error() != "strconv.Atoi: parsing \"xxx\": invalid syntax" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow(" 1:xxx"); err.Error() != "strconv.Atoi: parsing \"xxx\": invalid syntax" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseTimeWindow("1:2", " 11:22"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestTOTPListFilter(t *testing.T) {
	setupTOTP(t)
	args, _ := app.NewTOTPArguments([]string{"ls", "test"})
	m, opts := newMock(t)
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "test/test2/totp/otp\ntest/test3/totp/otp\n" {
		t.Errorf("invalid list: %s", m.buf.String())
	}
	m.buf.Reset()
	args, _ = app.NewTOTPArguments([]string{"ls", "[zzzz]"})
	if err := args.Do(opts); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.buf.String() != "" {
		t.Errorf("invalid list: %s", m.buf.String())
	}
}
