package backend_test

import (
	"encoding/json"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/backend"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

func setupInserts(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert("test/test/abc", "tedst")
	fullSetup(t, true).Insert("test/test/abcx", "tedst")
	fullSetup(t, true).Insert("test/test/ab11c", "tdest\ntest")
	fullSetup(t, true).Insert("test/test/abc1ak", "atest")
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
	if q[0].Path != "test/test/abc" || q[0].Value != "" {
		t.Error("invalid query result")
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
	q, err := fullSetup(t, true).Get("test/test/abc", backend.BlankValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Path != "test/test/abc" || q.Value != "" {
		t.Error("invalid query result")
	}
	q, err = fullSetup(t, true).Get("a/b/aaaa", backend.BlankValue)
	if err != nil || q != nil {
		t.Error("invalid result, should be empty")
	}
	if _, err := fullSetup(t, true).Get("aaaa", backend.BlankValue); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestValueModes(t *testing.T) {
	store.Clear()
	setupInserts(t)
	q, err := fullSetup(t, true).Get("test/test/abc", backend.BlankValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	q, err = fullSetup(t, true).Get("test/test/abc", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m := backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if m.Data != "44276ba24db13df5568aa6db81e0190ab9d35d2168dce43dca61e628f5c666b1d8b091f1dda59c2359c86e7d393d59723a421d58496d279031e7f858c11d893e" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	if len(m.ModTime) < 20 {
		t.Errorf("invalid date/time")
	}
	store.SetInt64("LOCKBOX_JSON_HASH_LENGTH", 10)
	q, err = fullSetup(t, true).Get("test/test/abc", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m = backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if m.Data != "44276ba24d" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	q, err = fullSetup(t, true).Get("test/test/ab11c", backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "tdest\ntest" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	q, err = fullSetup(t, true).Get("test/test/abc", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m = backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(m.ModTime) < 20 || m.Data == "" {
		t.Errorf("invalid json: %v", m)
	}
	store.SetString("LOCKBOX_JSON_MODE", "plAINtExt")
	q, err = fullSetup(t, true).Get("test/test/abc", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m = backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(m.ModTime) < 20 || m.Data != "tedst" {
		t.Errorf("invalid json: %v", m)
	}
	store.SetString("LOCKBOX_JSON_MODE", "emPTY")
	q, err = fullSetup(t, true).Get("test/test/abc", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m = backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(m.ModTime) < 20 || m.Data != "" {
		t.Errorf("invalid json: %v", m)
	}
}

func testCollect(t *testing.T, count int, seq backend.QuerySeq2) []backend.Entity {
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
	if _, err := fullSetup(t, true).QueryCallback(backend.QueryOptions{}); err.Error() != "no query mode specified" {
		t.Errorf("wrong error: %v", err)
	}
	seq, err := fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.ListMode})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res := testCollect(t, 4, seq)
	if res[0].Path != "test/test/ab11c" || res[1].Path != "test/test/abc" || res[2].Path != "test/test/abc1ak" || res[3].Path != "test/test/abcx" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.FindMode, Criteria: "1"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res = testCollect(t, 2, seq)
	if res[0].Path != "test/test/ab11c" || res[1].Path != "test/test/abc1ak" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.SuffixMode, Criteria: "c"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res = testCollect(t, 2, seq)
	if res[0].Path != "test/test/ab11c" || res[1].Path != "test/test/abc" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.ExactMode, Criteria: "test/test/abc"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res = testCollect(t, 1, seq)
	if res[0].Path != "test/test/abc" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.ExactMode, Criteria: "test/test/abc", PathFilter: "abc"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	res = testCollect(t, 1, seq)
	if res[0].Path != "test/test/abc" {
		t.Errorf("invalid results: %v", res)
	}
	seq, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.ExactMode, Criteria: "test/test/abc", PathFilter: "abz"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	testCollect(t, 0, seq)
	seq, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.ExactMode, Criteria: "abczzz"})
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
	tr.Insert("test/xyz", "test")
	q, err := fullSetup(t, true).Get("test/xyz", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m := backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if m.ModTime != testDateTime {
		t.Errorf("invalid date/time")
	}

	store.Clear()
	tr = fullSetup(t, false)
	tr.Insert("test/xyz", "test")
	q, err = fullSetup(t, true).Get("test/xyz", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m = backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if m.ModTime == testDateTime {
		t.Errorf("invalid date/time")
	}

	tr = fullSetup(t, false)
	store.SetString("LOCKBOX_DEFAULTS_MODTIME", "garbage")
	err = tr.Insert("test/xyz", "test")
	if err == nil || !strings.Contains(err.Error(), "parsing time") {
		t.Errorf("invalid error: %v", err)
	}
}

func TestAttributeModes(t *testing.T) {
	store.Clear()
	setupInserts(t)
	fullSetup(t, true).Insert("test/test/totp", "atest")
	q, err := fullSetup(t, true).Get("test/test/totp", backend.BlankValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	q, err = fullSetup(t, true).Get("test/test/totp", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m := backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if m.Data != "d39e1526f86e69f9ff1f443f5a575b5ce516b66746773997fa11bd4aefb5facd0a6fa79d0b970cad3af342ff5a63f4df1a05ef110573631d84f174b7b1d17c63" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	if m.Attributes == nil || len(m.Attributes) != 1 {
		t.Errorf("invalid result value: %v", m.Attributes)
	}
	val, ok := m.Attributes["otp"]
	if !ok || val != "7f8fd0e1a714f63da75206748d0ea1dd601fc8f92498bc87c9579b403c3004a0eefdd7ead976f7dbd6e5143c9aa7a569e24322d870ec7745a4605a154557458e" {
		t.Errorf("invalid result value: %v", m.Attributes)
	}
	if len(m.ModTime) < 20 {
		t.Errorf("invalid date/time")
	}
	store.SetInt64("LOCKBOX_JSON_HASH_LENGTH", 10)
	q, err = fullSetup(t, true).Get("test/test/totp", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m = backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if m.Data != "d39e1526f8" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	if m.Attributes == nil || len(m.Attributes) != 1 {
		t.Errorf("invalid result value: %v", m.Attributes)
	}
	val, ok = m.Attributes["otp"]
	if !ok || val != "7f8fd0e1a7" {
		t.Errorf("invalid result value: %v", m.Attributes)
	}
	store.SetString("LOCKBOX_JSON_MODE", "PlAINtExt")
	q, err = fullSetup(t, true).Get("test/test/totp", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m = backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(m.ModTime) < 20 || m.Data != "atest" {
		t.Errorf("invalid json: %v", m)
	}
	if m.Attributes == nil || len(m.Attributes) != 1 {
		t.Errorf("invalid result value: %v", m.Attributes)
	}
	val, ok = m.Attributes["otp"]
	if !ok || !strings.Contains(val, "otpauth://") {
		t.Errorf("invalid result value: %v", m.Attributes)
	}
	store.SetString("LOCKBOX_JSON_MODE", "emPty")
	q, err = fullSetup(t, true).Get("test/test/totp", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m = backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(m.ModTime) < 20 || m.Data != "" {
		t.Errorf("invalid json: %v", m)
	}
	if m.Attributes == nil || len(m.Attributes) != 1 {
		t.Errorf("invalid result value: %v", m.Attributes)
	}
	val, ok = m.Attributes["otp"]
	if !ok || val != "" {
		t.Errorf("invalid result value: %v", m.Attributes)
	}
}
