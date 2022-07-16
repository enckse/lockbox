package internal

import (
	"errors"
)

type (
	// Color are terminal colors for dumb terminal coloring.
	Color int
)

const (
	termBeginRed = "\033[1;31m"
	termEndRed   = "\033[0m"
	// ColorRed will get red terminal coloring.
	ColorRed = iota
)

// GetColor will retrieve start/end terminal coloration indicators.
func GetColor(color Color) (string, string, error) {
	if color != ColorRed {
		return "", "", errors.New("bad color")
	}
	interactive, err := IsInteractive()
	if err != nil {
		return "", "", err
	}
	colors := interactive
	if colors {
		isColored, err := isYesNoEnv(false, "LOCKBOX_NOCOLOR")
		if err != nil {
			return "", "", err
		}
		colors = !isColored
	}
	if colors {
		return termBeginRed, termEndRed, nil
	}
	return "", "", nil
}
