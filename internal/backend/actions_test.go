package backend_test

import (
	"os"
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
	if err := tr.Insert("a", "a", nil, false); err.Error() != "invalid transaction" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestInserts(t *testing.T) {
	if err := setup(t).Insert("", "", nil, false); err.Error() != "empty path not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("a", "", nil, false); err.Error() != "empty secret not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("value", "pass", nil, false); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert("value", "pass", nil, false); err.Error() != "trying to insert over existing entity" {
		t.Errorf("wrong error: %v", err)
	}
	if err := fullSetup(t, true).Insert("value", "pass2", &backend.QueryEntity{Path: "value"}, false); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert("value2", "pass", nil, true); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get("value", backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "pass2" {
		t.Errorf("invalid retrieval")
	}
	q, err = fullSetup(t, true).Get("value2", backend.SecretValue)
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
	if err := setup(t).Remove(&backend.QueryEntity{}); err.Error() != "unable to select entity for deletion" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{"test1", "test2", "test3", "test4", "test5"} {
		fullSetup(t, true).Insert(i, "pass", nil, false)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test3"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	check(t, []string{"test1", "test2", "test4", "test5"})
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test1"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	check(t, []string{"test2", "test4", "test5"})
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test5"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	check(t, []string{"test2", "test4"})
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test4"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	check(t, []string{"test2"})
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test2"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
}

func check(t *testing.T, checks []string) {
	tr := fullSetup(t, true)
	for _, c := range checks {
		q, err := tr.Get(c, backend.BlankValue)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if q == nil {
			t.Errorf("failed to find entity: %s", c)
		}
	}
}
