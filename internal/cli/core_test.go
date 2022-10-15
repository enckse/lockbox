package cli_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/cli"
)

func TestUsage(t *testing.T) {
	u, _ := cli.Usage()
	if len(u) != 20 {
		t.Errorf("invalid usage, out of date? %d", len(u))
	}
}

func TestCompletionsBash(t *testing.T) {
	os.Setenv("LOCKBOX_NOTOTP", "yes")
	os.Setenv("LOCKBOX_READONLY", "yes")
	os.Setenv("LOCKBOX_NOCLIP", "yes")
	defaults, _ := cli.BashCompletions(true)
	roNoTOTPClip, _ := cli.BashCompletions(false)
	if roNoTOTPClip[0] == defaults[0] {
		t.Error("should not match defaults")
	}
	os.Setenv("LOCKBOX_NOTOTP", "")
	roNoClip, _ := cli.BashCompletions(false)
	if roNoClip[0] == defaults[0] || roNoClip[0] == roNoTOTPClip[0] {
		t.Error("should not equal defaults nor no totp/clip")
	}
	os.Setenv("LOCKBOX_READONLY", "")
	os.Setenv("LOCKBOX_NOCLIP", "yes")
	noClip, _ := cli.BashCompletions(false)
	if roNoClip[0] == noClip[0] || noClip[0] == defaults[0] || noClip[0] == roNoTOTPClip[0] {
		t.Error("readonly/noclip != noclip (nor defaults, nor ro/no totp/clip)")
	}
	os.Setenv("LOCKBOX_READONLY", "yes")
	os.Setenv("LOCKBOX_NOCLIP", "")
	ro, _ := cli.BashCompletions(false)
	if roNoClip[0] == ro[0] || noClip[0] == ro[0] || ro[0] == defaults[0] || ro[0] == roNoTOTPClip[0] {
		t.Error("readonly/noclip != ro (nor ro == noclip, nor ro == defaults)")
	}
	os.Setenv("LOCKBOX_READONLY", "")
	os.Setenv("LOCKBOX_NOCLIP", "")
	os.Setenv("LOCKBOX_NOTOTP", "")
	isDefaultsToo, _ := cli.BashCompletions(false)
	if isDefaultsToo[0] != defaults[0] {
		t.Error("defaults should match env defaults")
	}
	for _, confirm := range [][]string{defaults, roNoClip, noClip, ro, isDefaultsToo} {
		if len(confirm) != 1 {
			t.Error("completions returned an invalid array")
		}
	}
}
