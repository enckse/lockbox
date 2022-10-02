package backend_test

import (
	"fmt"
	"os"
	"path/filepath"
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
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	return tr
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

func TestBadAction(t *testing.T) {
	tr := &backend.Transaction{}
	if err := tr.Insert("a/a/a", "a"); err.Error() != "invalid transaction" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestMove(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(filepath.Join("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(filepath.Join("test", "test2", "test3"), "pass")
	if err := fullSetup(t, true).Move(backend.QueryEntity{Path: filepath.Join("test", "test2", "test3"), Value: "abc"}, filepath.Join("test1", "test2", "test3")); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(filepath.Join("test1", "test2", "test3"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "abc" {
		t.Errorf("invalid retrieval")
	}
	if err := fullSetup(t, true).Move(backend.QueryEntity{Path: filepath.Join("test", "test2", "test1"), Value: "test"}, filepath.Join("test1", "test2", "test3")); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err = fullSetup(t, true).Get(filepath.Join("test1", "test2", "test3"), backend.SecretValue)
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
	if err := setup(t).Insert(filepath.Join("test", "offset"), "test"); err.Error() != "invalid component path" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("test", "test"); err.Error() != "invalid component path" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("a", ""); err.Error() != "empty secret not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert(filepath.Join("test", "offset", "value"), "pass"); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert(filepath.Join("test", "offset", "value"), "pass2"); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := fullSetup(t, true).Insert(filepath.Join("test", "offset", "value2"), "pass\npass"); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(filepath.Join("test", "offset", "value"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "pass2" {
		t.Errorf("invalid retrieval")
	}
	q, err = fullSetup(t, true).Get(filepath.Join("test", "offset", "value2"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "pass\npass" {
		t.Errorf("invalid retrieval")
	}
	if err := fullSetup(t, true).Insert(filepath.Join("test", "offset", "totp"), "5ae472abqdekjqykoyxk7hvc2leklq5n"); err != nil {
		t.Errorf("no error: %v", err)
	}
}

func TestRemoves(t *testing.T) {
	if err := setup(t).Remove(nil); err.Error() != "entity is empty/invalid" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Remove(&backend.QueryEntity{}); err.Error() != "invalid component path" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Remove(&backend.QueryEntity{Path: filepath.Join("test1", "test2", "test3")}); err.Error() != "failed to remove entity" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{"test1", "test2"} {
		fullSetup(t, true).Insert(filepath.Join(i, i, i), "pass")
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: filepath.Join("test1", "test1", "test1")}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, filepath.Join("test2", "test2", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: filepath.Join("test2", "test2", "test2")}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{filepath.Join("test", "test", "test1"), filepath.Join("test", "test", "test2"), filepath.Join("test", "test", "test3"), filepath.Join("test", "test1", "test2"), filepath.Join("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, "pass")
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test3"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, filepath.Join("test", "test", "test2"), filepath.Join("test", "test", "test1"), filepath.Join("test", "test1", "test2"), filepath.Join("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test1"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, filepath.Join("test", "test", "test2"), filepath.Join("test", "test1", "test2"), filepath.Join("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test1/test5"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, filepath.Join("test", "test", "test2"), filepath.Join("test", "test1", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test1/test2"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, filepath.Join("test", "test", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test2"}); err != nil {
		t.Errorf("wrong error: %v", err)
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
