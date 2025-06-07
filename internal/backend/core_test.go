package backend_test

import (
	"errors"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/backend"
)

func TestLoad(t *testing.T) {
	if _, err := backend.Load("  "); err.Error() != "no store set" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := backend.Load("garbage"); err.Error() != "should use a .kdbx extension" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := backend.Load("garbage.kdbx"); err.Error() != "invalid file, does not exist" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestIsDirectory(t *testing.T) {
	if backend.IsDirectory("") {
		t.Error("invalid directory detection")
	}
	if !backend.IsDirectory("/") {
		t.Error("invalid directory detection")
	}
	if backend.IsDirectory("/a") {
		t.Error("invalid directory detection")
	}
}

func TestBase(t *testing.T) {
	b := backend.Base("")
	if b != "" {
		t.Error("invalid base")
	}
	b = backend.Base("aaa")
	if b != "aaa" {
		t.Error("invalid base")
	}
	b = backend.Base("aaa/")
	if b != "" {
		t.Error("invalid base")
	}
	b = backend.Base("aaa/a")
	if b != "a" {
		t.Error("invalid base")
	}
}

func TestDirectory(t *testing.T) {
	b := backend.Directory("")
	if b != "" {
		t.Error("invalid directory")
	}
	b = backend.Directory("/")
	if b != "" {
		t.Error("invalid directory")
	}
	b = backend.Directory("/a")
	if b != "" {
		t.Error("invalid directory")
	}
	b = backend.Directory("a")
	if b != "" {
		t.Error("invalid directory")
	}
	b = backend.Directory("b/a")
	if b != "b" {
		t.Error("invalid directory")
	}
}

func TestIsLeafAttr(t *testing.T) {
	if backend.IsLeafAttribute("axyz", "z") {
		t.Error("invalid result")
	}
	if !backend.IsLeafAttribute("axy/z", "z") {
		t.Error("invalid result")
	}
}

func TestNewPath(t *testing.T) {
	p := backend.NewPath("abc", "xyz")
	if p != backend.NewPath("abc", "xyz") {
		t.Error("invalid new path")
	}
}

func TestNewSuffix(t *testing.T) {
	if backend.NewSuffix("test") != "/test" {
		t.Error("invalid suffix")
	}
}

func generateTestSeq(hasError, extra bool) backend.QuerySeq2 {
	return func(yield func(backend.Entity, error) bool) {
		if !yield(backend.Entity{}, nil) {
			return
		}
		if !yield(backend.Entity{}, nil) {
			return
		}
		if hasError {
			if !yield(backend.Entity{}, errors.New("test collect error")) {
				return
			}
		}
		if !yield(backend.Entity{}, nil) {
			return
		}
		if extra {
			if !yield(backend.Entity{}, nil) {
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
	e := backend.Entity{}
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
