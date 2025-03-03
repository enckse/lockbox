package backend_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/backend"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
	"git.sr.ht/~enckse/lockbox/internal/platform"
)

const (
	testDir = "testdata"
)

func testFile(name string) string {
	file := filepath.Join(testDir, name)
	if !platform.PathExists(testDir) {
		os.Mkdir(testDir, 0o755)
	}
	return file
}

func fullSetup(t *testing.T, keep bool) *backend.Transaction {
	file := testFile("test.kdbx")
	if !keep {
		os.Remove(file)
	}
	store.SetBool("LOCKBOX_READONLY", false)
	store.SetString("LOCKBOX_STORE", file)
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetString("LOCKBOX_TOTP_ENTRY", "totp")
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	return tr
}

func TestKeyFile(t *testing.T) {
	store.Clear()
	defer store.Clear()
	file := testFile("keyfile_test.kdbx")
	keyFile := testFile("file.key")
	os.Remove(file)
	store.SetString("LOCKBOX_STORE", file)
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	store.SetString("LOCKBOX_CREDENTIALS_KEY_FILE", keyFile)
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetString("LOCKBOX_TOTP_ENTRY", "totp")
	os.WriteFile(keyFile, []byte("test"), 0o644)
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
	store.SetBool("LOCKBOX_READONLY", true)
	tr, _ := backend.NewTransaction()
	if err := tr.Insert("a/a/a", "a"); err.Error() != "unable to alter database in readonly mode" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestBadTOTP(t *testing.T) {
	tr := setup(t)
	store.SetString("LOCKBOX_TOTP_ENTRY", "Title")
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
	if err := fullSetup(t, true).Move(nil, ""); err == nil || err.Error() != "source entity is not set" {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Move(&backend.Entity{Path: backend.NewPath("test", "test2", "test3"), Value: "abc"}, backend.NewPath("test1", "test2", "test3")); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(backend.NewPath("test1", "test2", "test3"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "abc" {
		t.Errorf("invalid retrieval")
	}
	if err := fullSetup(t, true).Move(&backend.Entity{Path: backend.NewPath("test", "test2", "test1"), Value: "test"}, backend.NewPath("test1", "test2", "test3")); err != nil {
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
	if err := setup(t).Remove(&backend.Entity{}); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	tx := backend.Entity{Path: backend.NewPath("test1", "test2", "test3")}
	if err := setup(t).Remove(&tx); err.Error() != "failed to remove entity" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{"test1", "test2"} {
		fullSetup(t, true).Insert(backend.NewPath(i, i, i), "pass")
	}
	tx = backend.Entity{Path: backend.NewPath("test1", "test1", "test1")}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test2", "test2", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = backend.Entity{Path: backend.NewPath("test2", "test2", "test2")}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test3"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, "pass")
	}
	tx = backend.Entity{Path: "test/test/test3"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = backend.Entity{Path: "test/test/test1"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = backend.Entity{Path: "test/test1/test5"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test1", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = backend.Entity{Path: "test/test1/test2"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = backend.Entity{Path: "test/test/test2"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRemoveAlls(t *testing.T) {
	if err := setup(t).RemoveAll(nil); err.Error() != "no entities given" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).RemoveAll([]backend.Entity{}); err.Error() != "no entities given" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test3"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, "pass")
	}
	if err := fullSetup(t, true).RemoveAll([]backend.Entity{{Path: "test/test/test3"}, {Path: "test/test/test1"}}); err != nil {
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

func TestKeyAndOrKeyFile(t *testing.T) {
	keyAndOrKeyFile(t, true, true)
	keyAndOrKeyFile(t, false, true)
	keyAndOrKeyFile(t, true, false)
	keyAndOrKeyFile(t, false, false)
}

func keyAndOrKeyFile(t *testing.T, key, keyFile bool) {
	store.Clear()
	file := testFile("keyorkeyfile.kdbx")
	os.Remove(file)
	store.SetString("LOCKBOX_STORE", file)
	if key {
		store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
		store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	} else {
		store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "none")
	}
	if keyFile {
		key := testFile("keyfileor.key")
		store.SetString("LOCKBOX_CREDENTIALS_KEY_FILE", key)
		os.WriteFile(key, []byte("test"), 0o644)
	}
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	invalid := !key && !keyFile
	err = tr.Insert(backend.NewPath("a", "b"), "t")
	if invalid {
		if err == nil || err.Error() != "key and/or keyfile must be set" {
			t.Errorf("invalid error: %v", err)
		}
	} else {
		if err != nil {
			t.Errorf("no error allowed: %v", err)
		}
	}
}

func TestReKey(t *testing.T) {
	store.Clear()
	f := "rekey_test.kdbx"
	file := testFile(f)
	defer os.Remove(filepath.Join(testDir, f))
	store.SetString("LOCKBOX_STORE", file)
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	store.SetString("LOCKBOX_TOTP_ENTRY", "totp")
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if err := tr.ReKey("", ""); err == nil || err.Error() != "key and/or keyfile must be set" {
		t.Errorf("no error: %v", err)
	}
	if err := tr.ReKey("abc", ""); err != nil {
		t.Errorf("no error: %v", err)
	}
}
