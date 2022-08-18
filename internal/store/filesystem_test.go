package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/misc"
)

func TestListErrors(t *testing.T) {
	_, err := FileSystem{path: "aaa"}.List(ViewOptions{})
	if err == nil || err.Error() != "store does not exist" {
		t.Errorf("invalid store error: %v", err)
	}
}

func TestList(t *testing.T) {
	testStore := "bin"
	if misc.PathExists(testStore) {
		if err := os.RemoveAll(testStore); err != nil {
			t.Errorf("invalid error on remove: %v", err)
		}
	}
	if err := os.MkdirAll(filepath.Join(testStore, "sub"), 0755); err != nil {
		t.Errorf("unable to makedir: %v", err)
	}
	for _, path := range []string{"test", "test2", "aaa", "sub/aaaaajk", "sub/12lkjafav"} {
		if err := os.WriteFile(filepath.Join(testStore, path+extension), []byte(""), 0644); err != nil {
			t.Errorf("failed to write %s: %v", path, err)
		}
	}
	s := FileSystem{path: testStore}
	res, err := s.List(ViewOptions{})
	if err != nil {
		t.Errorf("unable to list: %v", err)
	}
	if len(res) != 5 {
		t.Error("mismatched results")
	}
	res, err = s.List(ViewOptions{Display: true})
	if err != nil {
		t.Errorf("unable to list: %v", err)
	}
	if len(res) != 5 {
		t.Error("mismatched results")
	}
	if res[0] != "aaa" || res[1] != "sub/12lkjafav" || res[2] != "sub/aaaaajk" || res[3] != "test" || res[4] != "test2" {
		t.Errorf("not sorted: %v", res)
	}
	idx := 0
	res, err = s.List(ViewOptions{Filter: func(path string) string {
		if strings.Contains(path, "test") {
			idx++
			return fmt.Sprintf("%d", idx)
		}
		return ""
	}})
	if err != nil {
		t.Errorf("unable to list: %v", err)
	}
	if len(res) != 2 || res[0] != "1" || res[1] != "2" {
		t.Error("mismatch filter results")
	}
	res, err = s.List(ViewOptions{ErrorOnEmpty: false, Filter: func(path string) string {
		return ""
	}})
	if err != nil {
		t.Errorf("should be non-error: %v", err)
	}
	if len(res) != 0 {
		t.Error("should be empty list")
	}
	_, err = s.List(ViewOptions{ErrorOnEmpty: true, Filter: func(path string) string {
		return ""
	}})
	if err == nil || err.Error() != "no results found" {
		t.Errorf("should be non-error: %v", err)
	}
}

func TestFileSystemFile(t *testing.T) {
	f := FileSystem{path: "abc"}
	p := f.NewPath("test")
	if p != "abc/test.lb" {
		t.Error("invalid join result")
	}
}

func TestCleanPath(t *testing.T) {
	f := FileSystem{path: "abc"}
	c := f.CleanPath("xyz")
	if c != "xyz" {
		t.Error("invalid clean")
	}
	c = f.CleanPath("abc/xyz")
	if c != "xyz" {
		t.Error("invalid clean")
	}
	c = f.CleanPath("xyz.lb.lb")
	if c != "xyz.lb" {
		t.Error("invalid clean")
	}
}

func TestNewFile(t *testing.T) {
	f := FileSystem{path: "abc"}.NewFile("xyz")
	if f != "xyz.lb" {
		t.Error("invalid file")
	}
	f = FileSystem{path: "abc"}.NewFile("xyz.lb")
	if f != "xyz.lb" {
		t.Error("invalid file, had suffix")
	}
}
