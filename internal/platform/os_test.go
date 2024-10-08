package platform_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/platform"
)

func TestPathExist(t *testing.T) {
	testDir := filepath.Join("testdata", "exists")
	os.RemoveAll(testDir)
	if platform.PathExists(testDir) {
		t.Error("test dir SHOULD NOT exist")
	}
	os.MkdirAll(testDir, 0o755)
	if !platform.PathExists(testDir) {
		t.Error("test dir SHOULD exist")
	}
}

func TestLoadEnvConfigs(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()
	files := []string{filepath.Join("testdata", "xyz"), filepath.Join("testdata", "abc")}
	if err := platform.LoadEnvConfigs(files...); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	cfg := files[1]
	os.WriteFile(cfg, []byte(`
TEST_X=1
TEST_Y=2
TEST_Z="1
TEST11=1"
TEST_3="abc $HOME $X $TEST_X"`), 0o644)
	if err := platform.LoadEnvConfigs(files...); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	env := fmt.Sprintf("%v", os.Environ())
	verify := func(expects []string) {
		for _, e := range expects {
			if !strings.Contains(env, e) {
				t.Errorf("invalid env: %s (missing '%s')", env, e)
			}
		}
	}
	verify([]string{"TEST_X=1", "TEST_Y=2", "TEST_3=abc   1", "TEST_Z=\"1", "TEST11=1\""})
	os.Setenv("HOME", "a123")
	if err := platform.LoadEnvConfigs(files...); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	env = fmt.Sprintf("%v", os.Environ())
	verify([]string{"TEST_X=1", "TEST_Y=2", "TEST_3=abc a123  1", "TEST_Z=\"1", "TEST11=1\""})
	os.Setenv("XYZ", "xyz")
	os.Setenv("HOME", "$TEST4")
	os.Setenv("TEST4", "$XYZ")
	if err := platform.LoadEnvConfigs(files...); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	env = fmt.Sprintf("%v", os.Environ())
	verify([]string{"TEST_X=1", "TEST_Y=2", "TEST_3=abc xyz  1", "TEST_Z=\"1", "TEST11=1\""})
	final := filepath.Join("testdata", "zzz")
	files = append(files, final)
	os.Setenv("TEST4", "a")
	os.WriteFile(final, []byte(`
TEST_X=1
TEST_Y=2
TEST_Z="1
TEST11=2"
TEST_3="abc $HOME $X $TEST_X"`), 0o644)
	if err := platform.LoadEnvConfigs(files...); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	env = fmt.Sprintf("%v", os.Environ())
	verify([]string{"TEST_X=1", "TEST_Y=2", "TEST_3=abc a  1", "TEST_Z=\"1", "TEST11=1\""})
}
