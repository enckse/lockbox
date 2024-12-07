package core_test

import (
	"testing"

	"github.com/seanenck/lockbox/internal/core"
)

func TestWrap(t *testing.T) {
	w := core.TextWrap(0, "")
	if w != "" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = core.TextWrap(0, "abc\n\nabc\nxyz\n")
	if w != "abc\n\nabc xyz\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = core.TextWrap(0, "abc\n\nabc\nxyz\n\nx")
	if w != "abc\n\nabc xyz\n\nx\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = core.TextWrap(5, "abc\n\nabc\nxyz\n\nx")
	if w != "     abc\n\n     abc xyz\n\n     x\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
}
