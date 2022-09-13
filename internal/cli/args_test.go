package cli_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/cli"
)

func TestClipArg(t *testing.T) {
	options := cli.ParseArgs("-clip")
	if !options.Clip {
		t.Error("clip should be set")
	}
}

func TestMultiArg(t *testing.T) {
	options := cli.ParseArgs("-multi")
	if !options.Multi {
		t.Error("multi should be set")
	}
}

func TestListArg(t *testing.T) {
	options := cli.ParseArgs("-list")
	if !options.List {
		t.Error("list should be set")
	}
}

func TestOnce(t *testing.T) {
	if options := cli.ParseArgs("-once"); !options.Once {
		t.Error("once should be set")
	}
}

func TestShort(t *testing.T) {
	if options := cli.ParseArgs("-short"); !options.Short {
		t.Error("short should be set")
	}
}

func TestYes(t *testing.T) {
	if options := cli.ParseArgs("-yes"); !options.Yes {
		t.Error("yes should be set")
	}
}

func TestDefaults(t *testing.T) {
	args := cli.ParseArgs("this/is/a/test")
	if args.Clip || args.Once || args.Short || args.List || args.Multi || args.Yes {
		t.Error("defaults should all be false")
	}
}
