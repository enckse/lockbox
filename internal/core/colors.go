// Package core has to assist with some color components
package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// ColorWindowDelimiter indicates how windows are split in env/config keys
	ColorWindowDelimiter = " "
	// ColorWindowSpan indicates the delineation betwee start -> end (start:end)
	ColorWindowSpan = ":"
)

// ColorWindow for handling terminal colors based on timing
type ColorWindow struct {
	Start int
	End   int
}

// ParseColorWindow will handle parsing a window of colors for TOTP operations
func ParseColorWindow(windowString string) ([]ColorWindow, error) {
	var rules []ColorWindow
	for _, item := range strings.Split(windowString, ColorWindowDelimiter) {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ColorWindowSpan)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid colorization rule found: %s", line)
		}
		s, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		e, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		if s < 0 || e < 0 || e < s || s > 59 || e > 59 {
			return nil, fmt.Errorf("invalid time found for colorization rule: %s", line)
		}
		rules = append(rules, ColorWindow{Start: s, End: e})
	}
	if len(rules) == 0 {
		return nil, errors.New("invalid colorization rules for totp, none found")
	}
	return rules, nil
}
