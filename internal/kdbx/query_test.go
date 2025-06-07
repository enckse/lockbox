package kdbx_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/kdbx"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

func compareEntity(actual *kdbx.Entity, expect kdbx.Entity) bool {
	if err := compareToEntity(actual, expect); err != nil {
		return false
	}
	return true
}

func compareToEntity(actual *kdbx.Entity, expect kdbx.Entity) error {
	if actual == nil || actual.Values == nil {
		return errors.New("invalid actual")
	}
	if actual.Path == "" || actual.Path != expect.Path {
		return errors.New("invalid actual, no path")
	}
	for k, v := range actual.Values {
		isMod := k == "modtime"
		if isMod {
			if len(v) < 20 {
				return fmt.Errorf("%s invalid mod time", k)
			}
		}
		e, ok := expect.Value(k)
		if !ok {
			if !isMod {
				return fmt.Errorf("%s is missing from expected", k)
			}
		}
		if e != v {
			if isMod {
				if e == "" {
					continue
				}
			}
			return fmt.Errorf("mismatch %s: (%s != %s)", k, e, v)
		}
	}
	return nil
}

func setupInserts(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert("test/test/abc", map[string]string{"password": "tedst", "notes": "xxx"})
	fullSetup(t, true).Insert("test/test/abcx", map[string]string{"password": "tedst"})
	fullSetup(t, true).Insert("test/test/ab11c", map[string]string{"password": "tedst", "notes": "tdest\ntest"})
	fullSetup(t, true).Insert("test/test/abc1ak", map[string]string{"password": "atest", "notes": "atest"})
}

func TestMatchPath(t *testing.T) {
	store.Clear()
	setupInserts(t)
	q, err := fullSetup(t, true).MatchPath("test/test/abc")
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(q) != 1 {
		t.Error("invalid entity result")
	}
	if q[0].Path != "test/test/abc" {
		t.Error("invalid query result")
	}
	for _, k := range []string{"notes", "password"} {
		if val, ok := q[0].Value(k); !ok || val != "" {
			t.Errorf("invalid result value: %s", k)
		}
	}
	q, err = fullSetup(t, true).MatchPath("test/test/abcxxx")
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(q) != 0 {
		t.Error("invalid entity result")
	}
	q, err = fullSetup(t, true).MatchPath("test/test/*")
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(q) != 4 {
		t.Error("invalid entity result")
	}
	if _, err := fullSetup(t, true).MatchPath("test/test//*"); err.Error() != "invalid match criteria, too many path separators" {
		t.Errorf("wrong error: %v", err)
	}
	q, err = fullSetup(t, true).MatchPath("test/test*")
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(q) != 0 {
		t.Error("invalid entity result")
	}
}

func TestGet(t *testing.T) {
	setupInserts(t)
	q, err := fullSetup(t, true).Get("test/test/abc", kdbx.BlankValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Path != "test/test/abc" {
		t.Error("invalid query result")
	}
	for _, k := range []string{"notes", "password"} {
		if val, ok := q.Value(k); !ok || val != "" {
			t.Errorf("invalid result value: %s", k)
		}
	}
	q, err = fullSetup(t, true).Get("a/b/aaaa", kdbx.BlankValue)
	if err != nil || q != nil {
		t.Error("invalid result, should be empty")
	}
	if _, err := fullSetup(t, true).Get("aaaa", kdbx.BlankValue); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestValueModes(t *testing.T) {
	store.Clear()
	setupInserts(t)
	q, err := fullSetup(t, true).Get("test/test/abc", kdbx.BlankValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	for _, k := range []string{"notes", "password"} {
		if val, ok := q.Value(k); !ok || val != "" {
			t.Errorf("invalid result value: %s", k)
		}
	}
	q, err = fullSetup(t, true).Get("test/test/abc", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/abc",
		Values: map[string]string{
			"notes":    "9057ff1aa9509b2a0af624d687461d2bbeb07e2f37d953b1ce4a9dc921a7f19c45dc35d7c5363b373792add57d0d7dc41596e1c585d6ef7844cdf8ae87af443f",
			"password": "44276ba24db13df5568aa6db81e0190ab9d35d2168dce43dca61e628f5c666b1d8b091f1dda59c2359c86e7d393d59723a421d58496d279031e7f858c11d893e",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
	store.SetInt64("LOCKBOX_JSON_HASH_LENGTH", 10)
	q, err = fullSetup(t, true).Get("test/test/abc", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/abc",
		Values: map[string]string{
			"notes":    "9057ff1aa9",
			"password": "44276ba24d",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
	q, err = fullSetup(t, true).Get("test/test/ab11c", kdbx.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/ab11c",
		Values: map[string]string{
			"notes":    "tdest\ntest",
			"password": "tedst",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
	store.SetString("LOCKBOX_JSON_MODE", "plAINtExt")
	q, err = fullSetup(t, true).Get("test/test/abc", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/abc",
		Values: map[string]string{
			"notes":    "xxx",
			"password": "tedst",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
	store.SetString("LOCKBOX_JSON_MODE", "emPTY")
	q, err = fullSetup(t, true).Get("test/test/abc", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/abc",
		Values: map[string]string{
			"notes":    "",
			"password": "",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
}

func testCollect(t *testing.T, count int, seq kdbx.QuerySeq2) []kdbx.Entity {
	collected, err := seq.Collect()
	if err != nil {
		t.Errorf("invalid collect error: %v", err)
	}
	if len(collected) != count {
		t.Errorf("unexpected entity count: %d", count)
	}
	return collected
}

func TestQueryCallback(t *testing.T) {
	setupInserts(t)
	if _, err := fullSetup(t, true).QueryCallback(kdbx.QueryOptions{}); err.Error() != "no query mode specified" {
		t.Errorf("wrong error: %v", err)
	}
	seq, err := fullSetup(t, true).QueryCallback(kdbx.QueryOptions{Mode: kdbx.ListMode})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res := testCollect(t, 4, seq)
	if res[0].Path != "test/test/ab11c" || res[1].Path != "test/test/abc" || res[2].Path != "test/test/abc1ak" || res[3].Path != "test/test/abcx" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(kdbx.QueryOptions{Mode: kdbx.FindMode, Criteria: "1"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res = testCollect(t, 2, seq)
	if res[0].Path != "test/test/ab11c" || res[1].Path != "test/test/abc1ak" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(kdbx.QueryOptions{Mode: kdbx.ExactMode, Criteria: "test/test/abc"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res = testCollect(t, 1, seq)
	if res[0].Path != "test/test/abc" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(kdbx.QueryOptions{Mode: kdbx.ExactMode, Criteria: "test/test/abc", PathFilter: "abc"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res = testCollect(t, 1, seq)
	if res[0].Path != "test/test/abc" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(kdbx.QueryOptions{Mode: kdbx.ExactMode, Criteria: "test/test/abc", PathFilter: "abz"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	testCollect(t, 0, seq)
	seq, err = fullSetup(t, true).QueryCallback(kdbx.QueryOptions{Mode: kdbx.ExactMode, Criteria: "abczzz"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	testCollect(t, 0, seq)
}

func TestSetModTime(t *testing.T) {
	store.Clear()
	testDateTime := "2022-12-30T12:34:56-05:00"
	tr := fullSetup(t, false)
	store.SetString("LOCKBOX_DEFAULTS_MODTIME", testDateTime)
	tr.Insert("test/xyz", map[string]string{"password": "test"})
	q, err := fullSetup(t, true).Get("test/xyz", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/xyz",
		Values: map[string]string{
			"password": "ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff",
			"modtime":  testDateTime,
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}

	store.Clear()
	tr = fullSetup(t, false)
	tr.Insert("test/xyz", map[string]string{"password": "test"})
	q, err = fullSetup(t, true).Get("test/xyz", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}

	if val, ok := q.Value("modtime"); !ok || len(val) < 20 || val == testDateTime {
		t.Errorf("invalid mod: %s", val)
	}

	tr = fullSetup(t, false)
	store.SetString("LOCKBOX_DEFAULTS_MODTIME", "garbage")
	err = tr.Insert("test/xyz", map[string]string{"password": "test"})
	if err == nil || !strings.Contains(err.Error(), "parsing time") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestAttributeModes(t *testing.T) {
	store.Clear()
	setupInserts(t)
	fullSetup(t, true).Insert("test/test/totp", map[string]string{"otp": "atest"})
	q, err := fullSetup(t, true).Get("test/test/totp", kdbx.BlankValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/totp",
		Values: map[string]string{
			"otp": "",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
	q, err = fullSetup(t, true).Get("test/test/totp", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/totp",
		Values: map[string]string{
			"otp": "7f8fd0e1a714f63da75206748d0ea1dd601fc8f92498bc87c9579b403c3004a0eefdd7ead976f7dbd6e5143c9aa7a569e24322d870ec7745a4605a154557458e",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
	store.SetInt64("LOCKBOX_JSON_HASH_LENGTH", 10)
	q, err = fullSetup(t, true).Get("test/test/totp", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/totp",
		Values: map[string]string{
			"otp": "7f8fd0e1a7",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
	store.SetString("LOCKBOX_JSON_MODE", "PlAINtExt")
	q, err = fullSetup(t, true).Get("test/test/totp", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/totp",
		Values: map[string]string{
			"otp": "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=atest",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
	store.SetString("LOCKBOX_JSON_MODE", "emPty")
	q, err = fullSetup(t, true).Get("test/test/totp", kdbx.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if !compareEntity(q, kdbx.Entity{
		Path: "test/test/totp",
		Values: map[string]string{
			"otp": "",
		},
	}) {
		t.Errorf("invalid entity: %v", q)
	}
}
