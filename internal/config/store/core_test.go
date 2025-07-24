package store_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/config/store"
)

func TestClear(t *testing.T) {
	store.Clear()
	store.SetString("abc", "abc")
	store.SetBool("xyz", true)
	store.SetArray("sss", []string{})
	store.SetInt64("aaa", 1)
	if len(store.List()) != 4 {
		t.Error("invalid list")
	}
	store.Clear()
	if len(store.List()) != 0 {
		t.Error("invalid list")
	}
}

func checkItem(keyValue store.KeyValue, key, value string) error {
	if keyValue.Key != key || fmt.Sprintf("%v", keyValue.Value) != value {
		return fmt.Errorf("invalid value: %v", keyValue)
	}
	return nil
}

func TestList(t *testing.T) {
	store.Clear()
	store.SetString("abc", "abc")
	store.SetBool("xyz", true)
	store.SetArray("sss", []string{})
	store.SetInt64("aaa", 1)
	l := store.List()
	if len(l) != 4 {
		t.Error("invalid list")
	}
	slices.SortFunc(l, func(x, y store.KeyValue) int {
		return strings.Compare(x.Key, y.Key)
	})
	if err := checkItem(l[0], "aaa", "1"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := checkItem(l[1], "abc", "abc"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := checkItem(l[2], "sss", "[]"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := checkItem(l[3], "xyz", "true"); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestGetSetBool(t *testing.T) {
	store.Clear()
	store.SetBool("xyz", true)
	val, ok := store.GetBool("xyz")
	if !val || !ok {
		t.Error("invalid get")
	}
	_, ok = store.GetBool("zzz")
	if ok {
		t.Error("invalid get")
	}
}

func TestGetSetString(t *testing.T) {
	store.Clear()
	store.SetString("xyz", "sss")
	val, ok := store.GetString("xyz")
	if val != "sss" || !ok {
		t.Error("invalid get")
	}
	_, ok = store.GetString("zzz")
	if ok {
		t.Error("invalid get")
	}
}

func TestGetSetArray(t *testing.T) {
	store.Clear()
	store.SetArray("xyz", []string{"xyz", "xxx"})
	val, ok := store.GetArray("xyz")
	if fmt.Sprintf("%v", val) != "[xyz xxx]" || !ok {
		t.Error("invalid get")
	}
	_, ok = store.GetArray("zzz")
	if ok {
		t.Error("invalid get")
	}
}

func TestGetSetInt64(t *testing.T) {
	store.Clear()
	store.SetInt64("xyz", 1)
	val, ok := store.GetInt64("xyz")
	if val != 1 || !ok {
		t.Error("invalid get")
	}
	_, ok = store.GetInt64("zzz")
	if ok {
		t.Error("invalid get")
	}
}
