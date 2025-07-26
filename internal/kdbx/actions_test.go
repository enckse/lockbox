package kdbx_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/enckse/lockbox/internal/config/store"
	"github.com/enckse/lockbox/internal/kdbx"
	"github.com/enckse/lockbox/internal/platform"
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

func fullSetup(t *testing.T, keep bool) *kdbx.Transaction {
	file := testFile("test.kdbx")
	if !keep {
		os.Remove(file)
	}
	store.SetBool("LOCKBOX_READONLY", false)
	store.SetString("LOCKBOX_STORE", file)
	store.SetArray("LOCKBOX_CREDENTIALS_PASSWORD", []string{"test"})
	store.SetString("LOCKBOX_CREDENTIALS_PASSWORD_MODE", "plaintext")
	tr, err := kdbx.NewTransaction()
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
	os.WriteFile(keyFile, []byte("test"), 0o644)
	tr, err := kdbx.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if err := tr.Insert(kdbx.NewPath("a", "b"), map[string]string{"password": "t"}); err != nil {
		t.Errorf("no error: %v", err)
	}
}

func setup(t *testing.T) *kdbx.Transaction {
	return fullSetup(t, false)
}

func TestNoWriteOnRO(t *testing.T) {
	setup(t)
	store.SetBool("LOCKBOX_READONLY", true)
	tr, _ := kdbx.NewTransaction()
	if err := tr.Insert("a/a/a", map[string]string{"password": "xyz"}); err.Error() != "unable to alter database in readonly mode" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestBadAction(t *testing.T) {
	tr := &kdbx.Transaction{}
	if err := tr.Insert("a/a/a", map[string]string{"notes": "xyz"}); err.Error() != "invalid transaction" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestMove(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(kdbx.NewPath("test", "test2", "test1"), map[string]string{"passworD": "pass"})
	fullSetup(t, true).Insert(kdbx.NewPath("test", "test2", "test3"), map[string]string{"NoTES": "pass", "password": "xxx"})
	if err := fullSetup(t, true).Move(); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := fullSetup(t, true).Move(kdbx.MoveRequest{nil, ""}); err == nil || err.Error() != "source entity is not set" {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Move(kdbx.MoveRequest{&kdbx.Entity{Path: kdbx.NewPath("test", "test2", "test3"), Values: map[string]string{"Notes": "abc"}}, kdbx.NewPath("test1", "test2", "test3")}); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(kdbx.NewPath("test1", "test2", "test3"), kdbx.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if val, ok := q.Value("notes"); !ok || val != "abc" {
		t.Errorf("invalid retrieval")
	}
	if err := fullSetup(t, true).Move(kdbx.MoveRequest{&kdbx.Entity{Path: kdbx.NewPath("test", "test2", "test1"), Values: map[string]string{"password": "test"}}, kdbx.NewPath("test1", "test2", "test3")}); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err = fullSetup(t, true).Get(kdbx.NewPath("test1", "test2", "test3"), kdbx.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if val, ok := q.Value("password"); !ok || val != "test" {
		t.Errorf("invalid retrieval")
	}
}

func TestInserts(t *testing.T) {
	if err := setup(t).Insert("", nil); err.Error() != "empty path not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("a", map[string]string{"randomfield": "1"}); err.Error() != "unknown entity field: randomfield" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests", map[string]string{"notes": "1"}); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests//l", map[string]string{"notes": "test"}); err.Error() != "unwilling to operate on path with empty segment" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests/", map[string]string{"password": "test"}); err.Error() != "path can NOT end with separator" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("/tests", map[string]string{"password": "test"}); err.Error() != "path can NOT be rooted" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("test", map[string]string{"otp": "test", "url": "xyz"}); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("a", nil); err.Error() != "empty secrets not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("a", make(map[string]string)); err.Error() != "empty secrets not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert(kdbx.NewPath("test", "offset", "value"), map[string]string{"password": "pass"}); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert(kdbx.NewPath("test", "offset", "value"), map[string]string{"NoTes": "pass2"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := fullSetup(t, true).Insert(kdbx.NewPath("test", "offset", "value2"), map[string]string{"NOTES": "pass\npass", "uRL": "123", "password": "xxx", "otP": "zzz"}); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(kdbx.NewPath("test", "offset", "value"), kdbx.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if val, ok := q.Value("notes"); !ok || val != "pass2" {
		t.Errorf("invalid retrieval")
	}
	q, err = fullSetup(t, true).Get(kdbx.NewPath("test", "offset", "value2"), kdbx.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if val, ok := q.Value("notes"); !ok || val != "pass\npass" {
		t.Errorf("invalid retrieval: %s", val)
	}
	if val, ok := q.Value("password"); !ok || val != "xxx" {
		t.Errorf("invalid retrieval: %s", val)
	}
	if val, ok := q.Value("otp"); !ok || val != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=zzz" {
		t.Errorf("invalid retrieval: %s", val)
	}
	if val, ok := q.Value("url"); !ok || val != "123" {
		t.Errorf("invalid retrieval: %s", val)
	}
	if err := fullSetup(t, true).Insert(kdbx.NewPath("test", "offset"), map[string]string{"otp": "5ae472sabqdekjqykoyxk7hvc2leklq5n"}); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert(kdbx.NewPath("test", "offset"), map[string]string{"OTP": "ljaf\n5ae472abqdekjqykoyxk7hvc2leklq5n"}); err == nil || err.Error() != "otp can NOT be multi-line" {
		t.Errorf("wrong error: %v", err)
	}
	if err := fullSetup(t, true).Insert(kdbx.NewPath("test", "offset"), map[string]string{"urL": "ljaf\n5ae472abqdekjqykoyxk7hvc2leklq5n"}); err == nil || err.Error() != "url can NOT be multi-line" {
		t.Errorf("wrong error: %v", err)
	}
	if err := fullSetup(t, true).Insert(kdbx.NewPath("test", "offset"), map[string]string{"password": "ljaf\n5ae472abqdekjqykoyxk7hvc2leklq5n"}); err == nil || err.Error() != "password can NOT be multi-line" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRemoves(t *testing.T) {
	if err := setup(t).Remove(nil); err.Error() != "entity is empty/invalid" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Remove(&kdbx.Entity{}); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	tx := kdbx.Entity{Path: kdbx.NewPath("test1", "test2", "test3")}
	if err := setup(t).Remove(&tx); err.Error() != "failed to remove entity" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{"test1", "test2"} {
		fullSetup(t, true).Insert(kdbx.NewPath(i, i, i), map[string]string{"PASSWORD": "pass"})
	}
	tx = kdbx.Entity{Path: kdbx.NewPath("test1", "test1", "test1")}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, kdbx.NewPath("test2", "test2", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = kdbx.Entity{Path: kdbx.NewPath("test2", "test2", "test2")}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{kdbx.NewPath("test", "test", "test1"), kdbx.NewPath("test", "test", "test2"), kdbx.NewPath("test", "test", "test3"), kdbx.NewPath("test", "test1", "test2"), kdbx.NewPath("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, map[string]string{"password": "pass"})
	}
	tx = kdbx.Entity{Path: "test/test/test3"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, kdbx.NewPath("test", "test", "test2"), kdbx.NewPath("test", "test", "test1"), kdbx.NewPath("test", "test1", "test2"), kdbx.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = kdbx.Entity{Path: "test/test/test1"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, kdbx.NewPath("test", "test", "test2"), kdbx.NewPath("test", "test1", "test2"), kdbx.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = kdbx.Entity{Path: "test/test1/test5"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, kdbx.NewPath("test", "test", "test2"), kdbx.NewPath("test", "test1", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = kdbx.Entity{Path: "test/test1/test2"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, kdbx.NewPath("test", "test", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	tx = kdbx.Entity{Path: "test/test/test2"}
	if err := fullSetup(t, true).Remove(&tx); err != nil {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRemoveAlls(t *testing.T) {
	if err := setup(t).RemoveAll(nil); err.Error() != "no entities given" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).RemoveAll([]kdbx.Entity{}); err.Error() != "no entities given" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{kdbx.NewPath("test", "test", "test1"), kdbx.NewPath("test", "test", "test2"), kdbx.NewPath("test", "test", "test3"), kdbx.NewPath("test", "test1", "test2"), kdbx.NewPath("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, map[string]string{"PaSsWoRd": "pass"})
	}
	if err := fullSetup(t, true).RemoveAll([]kdbx.Entity{{Path: "test/test/test3"}, {Path: "test/test/test1"}}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, kdbx.NewPath("test", "test", "test2"), kdbx.NewPath("test", "test1", "test2"), kdbx.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
}

func check(t *testing.T, checks ...string) error {
	tr := fullSetup(t, true)
	for _, c := range checks {
		q, err := tr.Get(c, kdbx.BlankValue)
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
	tr, err := kdbx.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	invalid := !key && !keyFile
	err = tr.Insert(kdbx.NewPath("a", "b"), map[string]string{"password": "t"})
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
	tr, err := kdbx.NewTransaction()
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
