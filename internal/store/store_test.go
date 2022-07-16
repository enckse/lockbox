package store

import (
	"os"
	"path/filepath"
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
		if err := os.WriteFile(filepath.Join(testStore, path+Extension), []byte(""), 0644); err != nil {
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
}
