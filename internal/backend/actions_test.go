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

func TestBadAction(t *testing.T) {
	tr := &backend.Transaction{}
	if err := tr.Insert("a/a/a", "a", false); err.Error() != "invalid transaction" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestInserts(t *testing.T) {
	if err := setup(t).Insert("", "", false); err.Error() != "empty path not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert(filepath.Join("test", "offset"), "test", false); err.Error() != "invalid component path" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("test", "test", false); err.Error() != "invalid component path" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("a", "", false); err.Error() != "empty secret not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert(filepath.Join("test", "offset", "value"), "pass", false); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert(filepath.Join("test", "offset", "value"), "pass2", false); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := fullSetup(t, true).Insert(filepath.Join("test", "offset", "value2"), "pass", true); err != nil {
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
	if q.Value != "pass" {
		t.Errorf("invalid retrieval")
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
		fullSetup(t, true).Insert(filepath.Join(i, i, i), "pass", false)
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
		fullSetup(t, true).Insert(i, "pass", false)
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
