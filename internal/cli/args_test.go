package cli

import (
	"testing"
)

func TestClipArg(t *testing.T) {
	for _, check := range []string{"-c", "-clip"} {
		options := ParseArgs(check)
		if !options.Clip {
			t.Error("clip should be set")
		}
	}
}

func TestMultiArg(t *testing.T) {
	for _, check := range []string{"-m", "-multi"} {
		options := ParseArgs(check)
		if !options.Multi {
			t.Error("multi should be set")
		}
	}
}

func TestListArg(t *testing.T) {
	for _, check := range []string{"-list", "-ls"} {
		options := ParseArgs(check)
		if !options.List {
			t.Error("list should be set")
		}
	}
}

func TestOnce(t *testing.T) {
	if options := ParseArgs("-once"); !options.Once {
		t.Error("once should be set")
	}
}

func TestShort(t *testing.T) {
	if options := ParseArgs("-short"); !options.Short {
		t.Error("short should be set")
	}
}

func TestYes(t *testing.T) {
	if options := ParseArgs("-yes"); !options.Yes {
		t.Error("yes should be set")
	}
}
