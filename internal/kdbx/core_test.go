package kdbx_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/kdbx"
)

func TestAllowedSort(t *testing.T) {
	set := fmt.Sprintf("%v", kdbx.AllowedFields)
	have := kdbx.AllowedFields
	slices.SortFunc(have, func(x, y string) int {
		return strings.Compare(strings.ToLower(x), strings.ToLower(y))
	})
	if fmt.Sprintf("%v", have) != set {
		t.Error("allowed fields has incorrect sort")
	}
}

func TestLoad(t *testing.T) {
	if _, err := kdbx.Load("  "); err.Error() != "no store set" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := kdbx.Load("garbage"); err.Error() != "should use a .kdbx extension" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := kdbx.Load("garbage.kdbx"); err.Error() != "invalid file, does not exist" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestIsDirectory(t *testing.T) {
	if kdbx.IsDirectory("") {
		t.Error("invalid directory detection")
	}
	if !kdbx.IsDirectory("/") {
		t.Error("invalid directory detection")
	}
	if kdbx.IsDirectory("/a") {
		t.Error("invalid directory detection")
	}
}

func TestBase(t *testing.T) {
	b := kdbx.Base("")
	if b != "" {
		t.Error("invalid base")
	}
	b = kdbx.Base("aaa")
	if b != "aaa" {
		t.Error("invalid base")
	}
	b = kdbx.Base("aaa/")
	if b != "" {
		t.Error("invalid base")
	}
	b = kdbx.Base("aaa/a")
	if b != "a" {
		t.Error("invalid base")
	}
}

func TestDirectory(t *testing.T) {
	b := kdbx.Directory("")
	if b != "" {
		t.Error("invalid directory")
	}
	b = kdbx.Directory("/")
	if b != "" {
		t.Error("invalid directory")
	}
	b = kdbx.Directory("/a")
	if b != "" {
		t.Error("invalid directory")
	}
	b = kdbx.Directory("a")
	if b != "" {
		t.Error("invalid directory")
	}
	b = kdbx.Directory("b/a")
	if b != "b" {
		t.Error("invalid directory")
	}
}

func TestIsLeafAttr(t *testing.T) {
	if kdbx.IsLeafAttribute("axyz", "z") {
		t.Error("invalid result")
	}
	if !kdbx.IsLeafAttribute("axy/z", "z") {
		t.Error("invalid result")
	}
}

func TestNewPath(t *testing.T) {
	p := kdbx.NewPath("abc", "xyz")
	if p != kdbx.NewPath("abc", "xyz") {
		t.Error("invalid new path")
	}
}

func TestNewSuffix(t *testing.T) {
	if kdbx.NewSuffix("test") != "/test" {
		t.Error("invalid suffix")
	}
}

func generateTestSeq(hasError, extra bool) kdbx.QuerySeq2 {
	return func(yield func(kdbx.Entity, error) bool) {
		if !yield(kdbx.Entity{}, nil) {
			return
		}
		if !yield(kdbx.Entity{}, nil) {
			return
		}
		if hasError {
			if !yield(kdbx.Entity{}, errors.New("test collect error")) {
				return
			}
		}
		if !yield(kdbx.Entity{}, nil) {
			return
		}
		if extra {
			if !yield(kdbx.Entity{}, nil) {
				return
			}
		}
	}
}

func TestQuerySeq2Collect(t *testing.T) {
	seq := generateTestSeq(true, true)
	if _, err := seq.Collect(); err == nil || err.Error() != "test collect error" {
		t.Errorf("invalid error: %v", err)
	}
	seq = generateTestSeq(false, false)
	c, err := seq.Collect()
	if err != nil || len(c) != 3 {
		t.Errorf("invalid collect: %v %v %d", c, err, len(c))
	}
	seq = generateTestSeq(false, true)
	c, err = seq.Collect()
	if err != nil || len(c) != 4 {
		t.Errorf("invalid collect: %v %v %d", c, err, len(c))
	}
}

func TestEntityValue(t *testing.T) {
	e := kdbx.Entity{}
	if _, ok := e.Value("key"); ok {
		t.Error("values are nil")
	}
	e.Values = make(map[string]string)
	if _, ok := e.Value("key"); ok {
		t.Error("values are not set")
	}
	e.Values["key2"] = "1"
	if _, ok := e.Value("key"); ok {
		t.Error("values are not matching")
	}
	if val, ok := e.Value("key2"); !ok || val != "1" {
		t.Error("values are not set")
	}
}
