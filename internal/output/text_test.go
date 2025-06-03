package output_test

import (
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/output"
)

func TestWrap(t *testing.T) {
	w := output.TextWrap(0, "")
	if w != "" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = output.TextWrap(0, "abc\n\nabc\nxyz\n")
	if w != "abc\n\nabc xyz\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = output.TextWrap(0, "abc\n\nabc\nxyz\n\nx")
	if w != "abc\n\nabc xyz\n\nx\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
	w = output.TextWrap(5, "abc\n\nabc\nxyz\n\nx")
	if w != "     abc\n\n     abc xyz\n\n     x\n\n" {
		t.Errorf("invalid wrap: %s", w)
	}
}
