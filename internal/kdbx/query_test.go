package kdbx_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/config/store"
	"github.com/enckse/lockbox/internal/kdbx"
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
	q, err = fullSetup(t, true).MatchPath("test/test*")
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(q) != 0 {
		t.Error("invalid entity result")
	}
}

func TestGlob(t *testing.T) {
	_, err := kdbx.Glob("[", "x")
	if err == nil {
		t.Errorf("invalid error: %v", err)
	}
	ok, err := kdbx.Glob("a", "b")
	if ok || err != nil {
		t.Errorf("invalid result/error: %v", err)
	}
	ok, err = kdbx.Glob("a/*", "a/b")
	if !ok || err != nil {
		t.Errorf("invalid result/error: %v", err)
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
	if _, err := fullSetup(t, true).Get("aaaa", kdbx.BlankValue); err.Error() != "input paths must contain at LEAST 2 components (excluding field)" {
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
			"checksum": "0049b",
			"notes":    "164f7d1c788400c54db852f5f1ef4629e4d0020a87e935dfd643dc4f765dfd201ce43b2b2ec23ff8f5b966ed15715f79d276d4ededf05691197096bb4247d665",
			"password": "a3ea1c021135a8070c62a3a1080d9cd3385ebca45687636ba87c9abd1f5c2d68b17d68e72dc22461d0c8fc371573c568664e98fbfb832fcdda000318211b9538",
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
			"checksum": "0049b",
			"notes":    "164f7d1c78",
			"password": "a3ea1c0211",
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
			"password": "f4d691c1399b47b1a17d64da4e91f27ee739d8e49eee11d3ca5185940353325cfd5892cd375dd6a82f0b9f6e52d0365b4ddc2510106d134a1c3e9283becf72c9",
			"modtime":  testDateTime,
			"checksum": "000ef",
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
			"checksum": "0007e",
			"otp":      "cb9c99a3ba9f3370238a302adf9d3f4fa7cf4a2e01fe0225a7f69563b7c8160bd773471481d28d2f6654a6c88b41c54ca5c9930740554578b59832bd8ac2ee66",
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
			"checksum": "0007e",
			"otp":      "cb9c99a3ba",
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
