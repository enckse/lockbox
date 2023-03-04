package app_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/backend"
)

func TestInsertArgs(t *testing.T) {
	obj := app.InsertOptions{}
	if _, err := app.ParseInsertArgs(obj, []string{}); err == nil || err.Error() != "insert requires an entry" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := app.ParseInsertArgs(obj, []string{"test", "test", "test"}); err == nil || err.Error() != "too many arguments" {
		t.Errorf("invalid error: %v", err)
	}
	r, err := app.ParseInsertArgs(obj, []string{"test"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if r.Multi || r.Entry != "test" {
		t.Error("invalid parse")
	}
	if _, err := app.ParseInsertArgs(obj, []string{"-t", "b"}); err == nil || err.Error() != "unknown argument" {
		t.Errorf("invalid error: %v", err)
	}
	r, err = app.ParseInsertArgs(obj, []string{"-multi", "test3"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !r.Multi || r.Entry != "test3" {
		t.Error("invalid parse")
	}
}

func TestInsertDo(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	args := app.InsertArgs{}
	var buf bytes.Buffer
	args.Opts.IsPipe = func() bool {
		return false
	}
	args.Entry = "test/test2"
	tx := fullSetup(t, true)
	args.Opts.Confirm = func(string) bool {
		return true
	}
	args.Opts.Input = func(bool, bool) ([]byte, error) {
		return nil, errors.New("failure")
	}
	if err := args.Do(&buf, tx); err == nil || err.Error() != "invalid input (failure)" {
		t.Errorf("invalid error: %v", err)
	}
	args.Opts.Confirm = func(string) bool {
		return false
	}
	args.Opts.IsPipe = func() bool {
		return true
	}
	if err := args.Do(&buf, tx); err == nil || err.Error() != "invalid input (failure)" {
		t.Errorf("invalid error: %v", err)
	}
	args.Opts.Input = func(bool, bool) ([]byte, error) {
		return []byte("TEST"), nil
	}
	args.Opts.Confirm = func(string) bool {
		return true
	}
	args.Entry = "a/b/c"
	if err := args.Do(&buf, tx); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() != "" {
		t.Error("invalid insert")
	}
	args.Opts.IsPipe = func() bool {
		return false
	}
	buf = bytes.Buffer{}
	if err := args.Do(&buf, tx); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("invalid insert")
	}
	buf = bytes.Buffer{}
	args.Entry = "test/test2/test1"
	if err := args.Do(&buf, tx); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() == "" {
		t.Error("invalid insert")
	}
	args.Opts.Confirm = func(string) bool {
		return false
	}
	buf = bytes.Buffer{}
	args.Entry = "test/test2/test1"
	if err := args.Do(&buf, tx); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if buf.String() != "" {
		t.Error("invalid insert")
	}
}
