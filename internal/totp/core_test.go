package totp_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/totp"
)

func TestNewArgumentsErrors(t *testing.T) {
	if _, err := totp.NewArguments(nil, ""); err == nil || err.Error() != "not enough arguments for totp" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := totp.NewArguments([]string{"test"}, ""); err == nil || err.Error() != "invalid token type, not set?" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := totp.NewArguments([]string{"test"}, "a"); err == nil || err.Error() != "unknown totp command" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := totp.NewArguments([]string{"ls", "test"}, "a"); err == nil || err.Error() != "list takes no arguments" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := totp.NewArguments([]string{"show"}, "a"); err == nil || err.Error() != "missing entry" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestNewArguments(t *testing.T) {
	args, _ := totp.NewArguments([]string{"ls"}, "test")
	if args.Mode != totp.ListMode || args.Entry != "" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"show", "test"}, "test")
	if args.Mode != totp.ShowMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"clip", "test"}, "test")
	if args.Mode != totp.ClipMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"short", "test"}, "test")
	if args.Mode != totp.ShortMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"once", "test"}, "test")
	if args.Mode != totp.OnceMode || args.Entry != "test" {
		t.Error("invalid args")
	}
	args, _ = totp.NewArguments([]string{"insert", "test2"}, "test")
	if args.Mode != totp.InsertMode || args.Entry != "test2/test" {
		t.Errorf("invalid args: %s", args.Entry)
	}
	args, _ = totp.NewArguments([]string{"insert", "test2/test"}, "test")
	if args.Mode != totp.InsertMode || args.Entry != "test2/test" {
		t.Errorf("invalid args: %s", args.Entry)
	}
}
