package backend_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/backend"
)

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
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	return tr
}

func TestKeyFile(t *testing.T) {
	os.Remove("file.key")
	os.Remove("keyfile_test.kdbx")
	os.Setenv("LOCKBOX_READONLY", "no")
	os.Setenv("LOCKBOX_STORE", "keyfile_test.kdbx")
	os.Setenv("LOCKBOX_KEY", "test")
	os.Setenv("LOCKBOX_KEYFILE", "file.key.kdbx")
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_TOTP", "totp")
	os.Setenv("LOCKBOX_HOOKDIR", "")
	os.Setenv("LOCKBOX_SET_MODTIME", "")
	os.WriteFile("file.key.kdbx", []byte("test"), 0644)
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if err := tr.Insert(backend.NewPath("a", "b"), "t"); err != nil {
		t.Errorf("no error: %v", err)
	}
}

func setup(t *testing.T) *backend.Transaction {
	return fullSetup(t, false)
}

func TestNoWriteOnRO(t *testing.T) {
	setup(t)
	os.Setenv("LOCKBOX_READONLY", "yes")
	tr, _ := backend.NewTransaction()
	if err := tr.Insert("a/a/a", "a"); err.Error() != "unable to alter database in readonly mode" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestBadTOTP(t *testing.T) {
	tr := setup(t)
	os.Setenv("LOCKBOX_TOTP", "Title")
	if err := tr.Insert("a/a/a", "a"); err.Error() != "invalid totp field, uses restricted name" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestBadAction(t *testing.T) {
	tr := &backend.Transaction{}
	if err := tr.Insert("a/a/a", "a"); err.Error() != "invalid transaction" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestMove(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	if err := fullSetup(t, true).Move(backend.QueryEntity{Path: backend.NewPath("test", "test2", "test3"), Value: "abc"}, backend.NewPath("test1", "test2", "test3")); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(backend.NewPath("test1", "test2", "test3"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "abc" {
		t.Errorf("invalid retrieval")
	}
	if err := fullSetup(t, true).Move(backend.QueryEntity{Path: backend.NewPath("test", "test2", "test1"), Value: "test"}, backend.NewPath("test1", "test2", "test3")); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err = fullSetup(t, true).Get(backend.NewPath("test1", "test2", "test3"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "test" {
		t.Errorf("invalid retrieval")
	}
}

func TestInserts(t *testing.T) {
	if err := setup(t).Insert("", ""); err.Error() != "empty path not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests", "test"); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests//l", "test"); err.Error() != "unwilling to operate on path with empty segment" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests/", "test"); err.Error() != "path can NOT end with separator" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("/tests", "test"); err.Error() != "path can NOT be rooted" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("test", "test"); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("a", ""); err.Error() != "empty secret not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert(backend.NewPath("test", "offset", "value"), "pass"); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert(backend.NewPath("test", "offset", "value"), "pass2"); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := fullSetup(t, true).Insert(backend.NewPath("test", "offset", "value2"), "pass\npass"); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(backend.NewPath("test", "offset", "value"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "pass2" {
		t.Errorf("invalid retrieval")
	}
	q, err = fullSetup(t, true).Get(backend.NewPath("test", "offset", "value2"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "pass\npass" {
		t.Errorf("invalid retrieval")
	}
	if err := fullSetup(t, true).Insert(backend.NewPath("test", "offset", "totp"), "5ae472abqdekjqykoyxk7hvc2leklq5n"); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert(backend.NewPath("test", "offset", "totp"), "ljaf\n5ae472abqdekjqykoyxk7hvc2leklq5n"); err.Error() != "totp tokens can NOT be multi-line" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRemoves(t *testing.T) {
	if err := setup(t).Remove(nil); err.Error() != "entity is empty/invalid" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Remove(&backend.QueryEntity{}); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Remove(&backend.QueryEntity{Path: backend.NewPath("test1", "test2", "test3")}); err.Error() != "failed to remove entity" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{"test1", "test2"} {
		fullSetup(t, true).Insert(backend.NewPath(i, i, i), "pass")
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: backend.NewPath("test1", "test1", "test1")}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test2", "test2", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: backend.NewPath("test2", "test2", "test2")}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test3"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, "pass")
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test3"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test1"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test1/test5"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test1", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test1/test2"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test2"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRemoveAlls(t *testing.T) {
	if err := setup(t).RemoveAll(nil); err.Error() != "no entities given" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).RemoveAll([]backend.QueryEntity{}); err.Error() != "no entities given" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test3"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, "pass")
	}
	if err := fullSetup(t, true).RemoveAll([]backend.QueryEntity{{Path: "test/test/test3"}, {Path: "test/test/test1"}}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
}

func check(t *testing.T, checks ...string) error {
	tr := fullSetup(t, true)
	for _, c := range checks {
		q, err := tr.Get(c, backend.BlankValue)
		if err != nil {
			return err
		}
		if q == nil {
			return fmt.Errorf("failed to find entity: %s", c)
		}
	}
	return nil
}

func TestHooks(t *testing.T) {
	os.Setenv("LOCKBOX_HOOKDIR", "")
	h, err := backend.NewHook("a", backend.InsertAction)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := h.Run(backend.HookPre); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := backend.NewHook("", backend.InsertAction); err.Error() != "empty path is not allowed for hooks" {
		t.Errorf("wrong error: %v", err)
	}
	os.Setenv("LOCKBOX_HOOKDIR", "is_garbage")
	if _, err := backend.NewHook("b", backend.InsertAction); err.Error() != "hook directory does NOT exist" {
		t.Errorf("wrong error: %v", err)
	}
	testPath := "hooks.kdbx"
	os.RemoveAll(testPath)
	if err := os.MkdirAll(testPath, 0755); err != nil {
		t.Errorf("failed, mkdir: %v", err)
	}
	os.Setenv("LOCKBOX_HOOKDIR", testPath)
	h, err = backend.NewHook("a", backend.InsertAction)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := h.Run(backend.HookPre); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	sub := filepath.Join(testPath, "subdir")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Errorf("failed, mkdir sub: %v", err)
	}
	if _, err := backend.NewHook("b", backend.InsertAction); err.Error() != "found subdirectory in hookdir" {
		t.Errorf("wrong error: %v", err)
	}
	if err := os.RemoveAll(sub); err != nil {
		t.Errorf("failed rmdir: %v", err)
	}
	script := filepath.Join(testPath, "testscript")
	if err := os.WriteFile(script, []byte{}, 0644); err != nil {
		t.Errorf("unable to write script: %v", err)
	}
	h, err = backend.NewHook("a", backend.InsertAction)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := h.Run(backend.HookPre); strings.Contains("fork/exec", err.Error()) {
		t.Errorf("wrong error: %v", err)
	}
}
