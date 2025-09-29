package config_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/config/store"
)

type mockReader struct {
	read  func(string) (io.Reader, error)
	check func(string) bool
}

func (m mockReader) Read(path string) (io.Reader, error) {
	if m.read == nil {
		return nil, nil
	}
	return m.read(path)
}

func (m mockReader) Check(path string) bool {
	if m.check == nil {
		return true
	}
	return m.check(path)
}

func TestLoadIncludes(t *testing.T) {
	store.Clear()
	defer os.Clearenv()
	t.Setenv("TEST", "xyz")
	data := `include = ["$TEST/abc"]`
	r := strings.NewReader(data)
	mock := mockReader{}
	mock.read = func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"$TEST/abc\"]"), nil
		}
		return nil, errors.New("invalid path")
	}
	if err := config.Load(r, mock); err == nil || err.Error() != "too many nested includes (11 > 10)" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = ["abc"]`
	r = strings.NewReader(data)
	mock.read = func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"aaa\"]"), nil
		}
		return nil, errors.New("invalid path")
	}
	if err := config.Load(r, mock); err == nil || err.Error() != "invalid path" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = 1`
	r = strings.NewReader(data)
	mock.read = func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"aaa\"]"), nil
		}
		return nil, errors.New("invalid path")
	}
	if err := config.Load(r, mock); err == nil || err.Error() != "value is not of array type: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = [1]`
	r = strings.NewReader(data)
	mock.read = func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [\"aaa\"]"), nil
		}
		return nil, errors.New("invalid path")
	}
	if err := config.Load(r, mock); err == nil || err.Error() != "value is not valid array value: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = ["$TEST/abc"]
store="xyz"
`
	r = strings.NewReader(data)
	mock.read = func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("store = 'abc'"), nil
		}
		return nil, errors.New("invalid path")
	}
	if err := config.Load(r, mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 1 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetString("LOCKBOX_STORE")
	if val != "abc" || !ok {
		t.Errorf("invalid object: %v", val)
	}
}

func TestArrayLoad(t *testing.T) {
	store.Clear()
	defer os.Clearenv()
	t.Setenv("TEST", "abc")
	data := `store="xyz"
[clip]
copy = ["'xyz/$TEST'", "s", 1]
`
	r := strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || err.Error() != "value is not valid array value: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
store="xyz"
[clip]
copy = ["'xyz/$TEST'", "s"]
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 2 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetString("LOCKBOX_STORE")
	if val != "xyz" || !ok {
		t.Errorf("invalid object: %v", val)
	}
	a, ok := store.GetArray("LOCKBOX_CLIP_COPY")
	if fmt.Sprintf("%v", a) != "['xyz/abc' s]" || !ok {
		t.Errorf("invalid object: %v", a)
	}
	data = `include = [{file = 'xxx', required = false}]
store="xyz"
[clip]
copy = [{file = "'cliptest/$TEST'"}, "s"]
`
	r = strings.NewReader(data)
	mock := mockReader{}
	mock.check = func(string) bool {
		return false
	}
	if err := config.Load(r, mock); err == nil || !strings.Contains(err.Error(), "value is not valid array value:") || !strings.Contains(err.Error(), "cliptest/") {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 2 {
		t.Errorf("invalid store")
	}
	data = `include = []
store="xyz"
[clip]
copy = ["'xyz/$TEST'", "s"]
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 2 {
		t.Errorf("invalid store")
	}
	val, ok = store.GetString("LOCKBOX_STORE")
	if val != "xyz" || !ok {
		t.Errorf("invalid object: %v", val)
	}
	a, ok = store.GetArray("LOCKBOX_CLIP_COPY")
	if fmt.Sprintf("%v", a) != "['xyz/abc' s]" || !ok {
		t.Errorf("invalid object: %v", val)
	}
}

func TestReadInt(t *testing.T) {
	store.Clear()
	data := `
[json]
hash_length = true
`
	r := strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || err.Error() != "non-int64 found where int64 expected: true" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[json]
hash_length = 1
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 1 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetInt64("LOCKBOX_JSON_HASH_LENGTH")
	if val != 1 || !ok {
		t.Errorf("invalid object: %v", val)
	}
	data = `include = []
[json]
hash_length = -1
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || err.Error() != "-1 is negative (not allowed here)" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReadBool(t *testing.T) {
	store.Clear()
	data := `
[feature]
clip = 1
`
	r := strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || err.Error() != "non-bool found where bool expected: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[feature]
clip = true
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 1 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetBool("LOCKBOX_FEATURE_CLIP")
	if !val || !ok {
		t.Errorf("invalid object: %v", val)
	}
	data = `include = []
[feature]
clip = false
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 1 {
		t.Errorf("invalid store")
	}
	val, ok = store.GetBool("LOCKBOX_FEATURE_CLIP")
	if val || !ok {
		t.Errorf("invalid object: %v", val)
	}
}

func TestBadValues(t *testing.T) {
	store.Clear()
	data := `
[totsp]
enabled = "false"
`
	r := strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || err.Error() != "unknown key: totsp_enabled (LOCKBOX_TOTSP_ENABLED)" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = []
[totp]
otp_format = -1
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || err.Error() != "non-string found where string expected: -1" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestExpands(t *testing.T) {
	store.Clear()
	t.Setenv("TEST", "1")
	data := `include = []
store = "$TEST"
clip.copy = ["$TEST", "$TEST"]
[totp]
otp_format = "$TEST"
`
	r := strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(store.List()) != 3 {
		t.Errorf("invalid store")
	}
	val, ok := store.GetString("LOCKBOX_TOTP_OTP_FORMAT")
	if val != "$TEST" || !ok {
		t.Errorf("invalid object: %v", val)
	}
	val, ok = store.GetString("LOCKBOX_STORE")
	if val != "1" || !ok {
		t.Errorf("invalid object: %v", val)
	}
	a, ok := store.GetArray("LOCKBOX_CLIP_COPY")
	if fmt.Sprintf("%v", a) != "[1 1]" || !ok {
		t.Errorf("invalid object: %v", a)
	}
}

func TestLoadIncludesControls(t *testing.T) {
	store.Clear()
	defer os.Clearenv()
	t.Setenv("TEST", "xyz")
	data := `include = ["$TEST/abc"]
store="xyz"
`
	r := strings.NewReader(data)
	mock := mockReader{}
	mock.read = func(p string) (io.Reader, error) {
		if p == "xyz/abc" {
			return strings.NewReader("include = [{file = '123', required = 1}]\nstore = 'abc'"), nil
		}
		return nil, errors.New("invalid path")
	}
	if err := config.Load(r, mock); err == nil || err.Error() != "non-bool found where bool expected: 1" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = [{file = "$TEST/abc", required = true}]
store="xyz"
`
	r = strings.NewReader(data)
	mock.check = func(string) bool {
		return false
	}
	if err := config.Load(r, mock); err == nil || err.Error() != "failed to load the included file: xyz/abc" {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = [{file = "$TEST/abc", required = false}]
store="xyz"
`
	r = strings.NewReader(data)
	if err := config.Load(r, mock); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = [{file = "$TEST/abc", required = false, other = 1}]
store="xyz"
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || !strings.Contains(err.Error(), "invalid map array, too many keys:") {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = [{fsle = "$TEST/abc"}]
store="xyz"
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || !strings.Contains(err.Error(), "'file' is required, missing:") {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = [{file = "$TEST/abc", require = 1}]
store="xyz"
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || !strings.Contains(err.Error(), "only 'required' key is allowed here:") {
		t.Errorf("invalid error: %v", err)
	}
	data = `include = [{file = "$TEST/abc", required = 1}]
store="xyz"
`
	r = strings.NewReader(data)
	if err := config.Load(r, mockReader{}); err == nil || err.Error() != "non-bool found where bool expected: 1" {
		t.Errorf("invalid error: %v", err)
	}
}
