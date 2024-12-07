package util_test

import (
	"testing"

	"github.com/seanenck/lockbox/internal/util"
)

type mockReadFile struct{}

func (m mockReadFile) ReadFile(path string) ([]byte, error) {
	return []byte(path), nil
}

func TestReadEmbed(t *testing.T) {
	read, err := util.ReadDirFile("xyz", "zzz", mockReadFile{})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if string(read) != "xyz/zzz" {
		t.Errorf("invalid read: %s", string(read))
	}
}
