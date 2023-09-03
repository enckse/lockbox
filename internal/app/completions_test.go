package app_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestGenerateCompletions(t *testing.T) {
	testCompletions(t, true)
	testCompletions(t, false)
}

func generate(keys []string, bash bool, t *testing.T) (string, string) {
	os.Setenv("LOCKBOX_NOTOTP", "")
	os.Setenv("LOCKBOX_READONLY", "")
	os.Setenv("LOCKBOX_NOCLIP", "")
	os.Setenv("LOCKBOX_KEYMODE", "")
	key := "bash"
	if !bash {
		key = "zsh"
	}
	for _, k := range keys {
		use := "yes"
		if k == "KEYMODE" {
			use = "ask"
		}
		os.Setenv(fmt.Sprintf("LOCKBOX_%s", k), use)
		key = fmt.Sprintf("%s-%s", key, strings.ToLower(k))
	}
	v, err := app.GenerateCompletions(bash, false, "lb")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(v) != 1 {
		t.Errorf("invalid result")
	}
	return key, v[0]
}

func generateTest(keys []string, bash bool, t *testing.T) map[string]string {
	r := make(map[string]string)
	if len(keys) == 0 {
		return r
	}
	k, v := generate(keys, bash, t)
	r[k] = v
	for _, cur := range keys {
		var subset []string
		for _, key := range keys {
			if key == cur {
				continue
			}
			subset = append(subset, key)
		}

		for k, v := range generateTest(subset, bash, t) {
			r[k] = v
		}
	}
	return r
}

func testCompletions(t *testing.T, bash bool) {
	m := make(map[string]string)
	defaults, _ := app.GenerateCompletions(bash, true, "lb")
	m["defaults"] = defaults[0]
	for k, v := range generateTest([]string{"NOTOTP", "READONLY", "NOCLIP", "KEYMODE"}, true, t) {
		m[k] = v
	}
	os.Setenv("LOCKBOX_KEYMODE", "")
	os.Setenv("LOCKBOX_READONLY", "")
	os.Setenv("LOCKBOX_NOCLIP", "")
	os.Setenv("LOCKBOX_NOTOTP", "")
	defaultsToo, _ := app.GenerateCompletions(bash, false, "lb")
	if defaultsToo[0] != defaults[0] || len(defaultsToo) != 1 || len(defaults) != 1 {
		t.Error("defaults should match env defaults/invalid defaults detected")
	}
	for k, v := range m {
		fmt.Println(k)
		for kOther, vOther := range m {
			if kOther == k {
				continue
			}
			if vOther == v {
				t.Errorf("found overlapping completion: %s == %s", k, kOther)
			}
		}
	}
}
