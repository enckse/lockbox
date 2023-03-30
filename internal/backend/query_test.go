package backend_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/backend"
)

func setupInserts(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert("test/test/abc", "tedst")
	fullSetup(t, true).Insert("test/test/abcx", "tedst")
	fullSetup(t, true).Insert("test/test/ab11c", "tdest\ntest")
	fullSetup(t, true).Insert("test/test/abc1ak", "atest")
}

func TestMatchPath(t *testing.T) {
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
	setupInserts(t)
	q, err := fullSetup(t, true).Get("test/test/abc", backend.BlankValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	q, err = fullSetup(t, true).Get("test/test/abc", backend.HashedValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	hash := strings.TrimSpace(q.Value)
	parts := strings.Split(hash, "\n")
	if len(parts) != 2 {
		t.Errorf("invalid hash output: %v", parts)
	}
	if parts[1] != "hash: 44276ba24db13df5568aa6db81e0190ab9d35d2168dce43dca61e628f5c666b1d8b091f1dda59c2359c86e7d393d59723a421d58496d279031e7f858c11d893e" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	dt := parts[0]
	if !strings.HasPrefix(dt, "modtime: ") || len(dt) <= 20 {
		t.Errorf("invalid date/time: %s", dt)
	}
	q, err = fullSetup(t, true).Get("test/test/ab11c", backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "tdest\ntest" {
		t.Errorf("invalid result value: %s", q.Value)
	}
	q, err = fullSetup(t, true).Get("test/test/abc", backend.StatsValue)
	if err != nil || !strings.HasPrefix(q.Value, "modtime: ") || len(strings.Split(q.Value, "\n")) != 1 {
		t.Errorf("invalid stats: %s", q.Value)
	}
	q, err = fullSetup(t, true).Get("test/test/abc", backend.JSONValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	m := backend.JSON{}
	if err := json.Unmarshal([]byte(q.Value), &m); err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(m.ModTime) != 25 || m.Path != "test/test/abc" {
		t.Errorf("invalid json: %v", m)
	}
}

func TestQueryCallback(t *testing.T) {
	setupInserts(t)
	if _, err := fullSetup(t, true).QueryCallback(backend.QueryOptions{}); err.Error() != "no query mode specified" {
		t.Errorf("wrong error: %v", err)
	}
	res, err := fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.ListMode})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(res) != 4 {
		t.Error("invalid results: not enough")
	}
	if res[0].Path != "test/test/ab11c" || res[1].Path != "test/test/abc" || res[2].Path != "test/test/abc1ak" || res[3].Path != "test/test/abcx" {
		t.Errorf("invalid results: %v", res)
	}
	res, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.FindMode, Criteria: "1"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(res) != 2 {
		t.Error("invalid results: not enough")
	}
	if res[0].Path != "test/test/ab11c" || res[1].Path != "test/test/abc1ak" {
		t.Errorf("invalid results: %v", res)
	}
	res, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.SuffixMode, Criteria: "c"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(res) != 2 {
		t.Error("invalid results: not enough")
	}
	if res[0].Path != "test/test/ab11c" || res[1].Path != "test/test/abc" {
		t.Errorf("invalid results: %v", res)
	}
	res, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.ExactMode, Criteria: "test/test/abc"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(res) != 1 {
		t.Error("invalid results: not enough")
	}
	if res[0].Path != "test/test/abc" {
		t.Errorf("invalid results: %v", res)
	}
	res, err = fullSetup(t, true).QueryCallback(backend.QueryOptions{Mode: backend.ExactMode, Criteria: "abczzz"})
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if len(res) != 0 {
		t.Error("invalid results: should be empty")
	}
}

func TestEntityDir(t *testing.T) {
	q := backend.QueryEntity{Path: backend.NewPath("abc", "xyz")}
	if q.Directory() != "abc" {
		t.Error("invalid query directory")
	}
	q = backend.QueryEntity{Path: backend.NewPath("abc", "xyz", "111")}
	if q.Directory() != "abc/xyz" {
		t.Error("invalid query directory")
	}
	q = backend.QueryEntity{Path: ""}
	if q.Directory() != "" {
		t.Error("invalid query directory")
	}
	q = backend.QueryEntity{Path: backend.NewPath("abc")}
	if q.Directory() != "" {
		t.Error("invalid query directory")
	}
}

func TestNewPath(t *testing.T) {
	p := backend.NewPath("abc", "xyz")
	if p != backend.NewPath("abc", "xyz") {
		t.Error("invalid new path")
	}
}

func TestSetModTime(t *testing.T) {
	testDateTime := "2022-12-30T12:34:56-05:00"
	tr := fullSetup(t, false)
	os.Setenv("LOCKBOX_SET_MODTIME", testDateTime)
	tr.Insert("test/xyz", "test")
	q, err := fullSetup(t, true).Get("test/xyz", backend.HashedValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	hash := strings.TrimSpace(q.Value)
	parts := strings.Split(hash, "\n")
	if len(parts) != 2 {
		t.Errorf("invalid hash output: %v", parts)
	}
	dt := parts[0]
	if dt != fmt.Sprintf("modtime: %s", testDateTime) {
		t.Errorf("invalid date/time: %s", dt)
	}

	tr = fullSetup(t, false)
	os.Setenv("LOCKBOX_SET_MODTIME", "")
	tr.Insert("test/xyz", "test")
	q, err = fullSetup(t, true).Get("test/xyz", backend.HashedValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	hash = strings.TrimSpace(q.Value)
	parts = strings.Split(hash, "\n")
	if len(parts) != 2 {
		t.Errorf("invalid hash output: %v", parts)
	}
	dt = parts[0]
	if dt == fmt.Sprintf("modtime: %s", testDateTime) {
		t.Errorf("invalid date/time: %s", dt)
	}

	tr = fullSetup(t, false)
	os.Setenv("LOCKBOX_SET_MODTIME", "garbage")
	err = tr.Insert("test/xyz", "test")
	if err == nil || !strings.Contains(err.Error(), "parsing time") {
		t.Errorf("invalid error: %v", err)
	}
}
