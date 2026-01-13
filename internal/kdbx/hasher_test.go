package kdbx_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/config/store"
	"github.com/enckse/lockbox/internal/kdbx"
)

func TestTransform(t *testing.T) {
	hasher, _ := kdbx.NewHasher(kdbx.BlankValue)
	if hasher.Transform("xyz") != "" {
		t.Error("should be empty")
	}
	hasher, _ = kdbx.NewHasher(kdbx.SecretValue)
	if hasher.Transform("xyz") != "xyz" {
		t.Error("should be empty")
	}
	hasher, _ = kdbx.NewHasher(kdbx.JSONValue)
	if hasher.Transform("xyz") != "xyz" {
		t.Error("should be empty")
	}
	store.SetString("LOCKBOX_JSON_MODE", "plaintext")
	hasher, _ = kdbx.NewHasher(kdbx.JSONValue)
	if hasher.Transform("xyz") != "xyz" {
		t.Error("should be empty")
	}
	store.SetString("LOCKBOX_JSON_MODE", "hash")
	hasher, _ = kdbx.NewHasher(kdbx.JSONValue)
	if hasher.Transform("xyz") != "xyz" {
		t.Error("should be empty")
	}
}

func TestAdd(t *testing.T) {
	hasher, _ := kdbx.NewHasher(kdbx.BlankValue)
	if hasher.Add("x", "y") {
		t.Error("add should be false")
	}
	hasher, _ = kdbx.NewHasher(kdbx.SecretValue)
	if hasher.Add("x", "y") {
		t.Error("add should be false")
	}
	store.SetString("LOCKBOX_JSON_MODE", "plaintext")
	hasher, _ = kdbx.NewHasher(kdbx.JSONValue)
	if hasher.Add("x", "y") {
		t.Error("add should be false")
	}
	store.SetString("LOCKBOX_JSON_MODE", "hash")
	hasher, _ = kdbx.NewHasher(kdbx.JSONValue)
	if !hasher.Add("x", "y") {
		t.Error("add should be true")
	}
	if !hasher.Add("", "") {
		t.Error("add should be true")
	}
	if !hasher.Add("x", "") {
		t.Error("add should be true")
	}
	if !hasher.Add("", "y") {
		t.Error("add should be true")
	}
}

func TestCalculate(t *testing.T) {
	hasher, _ := kdbx.NewHasher(kdbx.BlankValue)
	hasher.Add("x", "y")
	if v, ok := hasher.Calculate("x"); ok || v != "" {
		t.Error("calculate is a noop")
	}
	hasher, _ = kdbx.NewHasher(kdbx.SecretValue)
	hasher.Add("x", "y")
	if v, ok := hasher.Calculate("x"); ok || v != "" {
		t.Error("calculate is a noop")
	}
	store.SetString("LOCKBOX_JSON_MODE", "plaintext")
	hasher, _ = kdbx.NewHasher(kdbx.JSONValue)
	hasher.Add("x", "y")
	if v, ok := hasher.Calculate("x"); ok || v != "" {
		t.Error("calculate is a noop")
	}
	store.SetString("LOCKBOX_JSON_MODE", "hash")
	hasher, _ = kdbx.NewHasher(kdbx.JSONValue)
	if v, ok := hasher.Calculate(""); !ok || v != "" {
		t.Error("result is ok, but empty")
	}
	hasher.Add("", "")
	if v, ok := hasher.Calculate(""); !ok || v != "" {
		t.Error("result is ok, but empty (nothing really added)")
	}
	if v, ok := hasher.Calculate("d"); !ok || v != "" {
		t.Error("result is ok, but empty (nothing really added, even with key)")
	}
	hasher.Add("x", "y")
	if v, ok := hasher.Calculate(""); !ok || v != "[00 00 00 00 00 1x]" {
		t.Errorf("results invalid for calculate: %s", v)
	}
	hasher.Add("z", "1")
	if v, ok := hasher.Calculate("d"); !ok || v != "[00 00 00 4d 1x 4z]" {
		t.Errorf("results invalid for calculate: %s", v)
	}
}

func TestReset(t *testing.T) {
	store.SetString("LOCKBOX_JSON_MODE", "hash")
	hasher, _ := kdbx.NewHasher(kdbx.JSONValue)
	hasher.Add("x", "y")
	if v, ok := hasher.Calculate(""); !ok || v != "[00 00 00 00 00 1x]" {
		t.Errorf("results invalid for calculate: %s", v)
	}
	hasher.Add("1", "y")
	if v, ok := hasher.Calculate(""); !ok || v != "[00 00 00 00 11 1x]" {
		t.Errorf("results invalid for calculate: %s", v)
	}
	hasher.Reset()
	hasher.Add("1", "z")
	if v, ok := hasher.Calculate(""); !ok || v != "[00 00 00 00 00 51]" {
		t.Errorf("results invalid for calculate: %s", v)
	}
}
