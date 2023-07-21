package cli_test

import (
	"os"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/cli"
)

func TestUsage(t *testing.T) {
	u, _ := cli.Usage(false)
	if len(u) != 24 {
		t.Errorf("invalid usage, out of date? %d", len(u))
	}
	u, _ = cli.Usage(true)
	if len(u) != 94 {
		t.Errorf("invalid verbose usage, out of date? %d", len(u))
	}
	for _, usage := range u {
		for _, l := range strings.Split(usage, "\n") {
			if len(l) > 79 {
				t.Errorf("usage line > 79 (%d), line: %s", len(l), l)
			}
		}
	}
}

func TestGenerateCompletions(t *testing.T) {
	testCompletions(t, true)
	testCompletions(t, false)
}

func testCompletions(t *testing.T, bash bool) {
	os.Setenv("LOCKBOX_NOTOTP", "yes")
	os.Setenv("LOCKBOX_READONLY", "yes")
	os.Setenv("LOCKBOX_NOCLIP", "yes")
	defaults, _ := cli.GenerateCompletions(bash, true)
	roNoTOTPClip, _ := cli.GenerateCompletions(bash, false)
	if roNoTOTPClip[0] == defaults[0] {
		t.Error("should not match defaults")
	}
	os.Setenv("LOCKBOX_NOTOTP", "")
	roNoClip, _ := cli.GenerateCompletions(bash, false)
	if roNoClip[0] == defaults[0] || roNoClip[0] == roNoTOTPClip[0] {
		t.Error("should not equal defaults nor no totp/clip")
	}
	os.Setenv("LOCKBOX_READONLY", "")
	os.Setenv("LOCKBOX_NOCLIP", "yes")
	noClip, _ := cli.GenerateCompletions(bash, false)
	if roNoClip[0] == noClip[0] || noClip[0] == defaults[0] || noClip[0] == roNoTOTPClip[0] {
		t.Error("readonly/noclip != noclip (nor defaults, nor ro/no totp/clip)")
	}
	os.Setenv("LOCKBOX_READONLY", "yes")
	os.Setenv("LOCKBOX_NOCLIP", "")
	ro, _ := cli.GenerateCompletions(bash, false)
	if roNoClip[0] == ro[0] || noClip[0] == ro[0] || ro[0] == defaults[0] || ro[0] == roNoTOTPClip[0] {
		t.Error("readonly/noclip != ro (nor ro == noclip, nor ro == defaults)")
	}
	os.Setenv("LOCKBOX_READONLY", "")
	os.Setenv("LOCKBOX_NOCLIP", "")
	os.Setenv("LOCKBOX_NOTOTP", "")
	isDefaultsToo, _ := cli.GenerateCompletions(bash, false)
	if isDefaultsToo[0] != defaults[0] {
		t.Error("defaults should match env defaults")
	}
	for _, confirm := range [][]string{defaults, roNoClip, noClip, ro, isDefaultsToo} {
		if len(confirm) != 1 {
			t.Error("completions returned an invalid array")
		}
	}
}
