package cli_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/cli"
)

func TestUsage(t *testing.T) {
	u := cli.Usage()
	if len(u) != 17 {
		t.Errorf("invalid usage, out of date? %d", len(u))
	}
}
