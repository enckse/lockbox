// Package inputs handles user inputs/UI elements.
package inputs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	colorWindowDelimiter = ","
	colorWindowSpan      = ":"
)

var (
	// TOTPDefaultColorWindow is the default coloring rules for totp
	TOTPDefaultColorWindow = []ColorWindow{{Start: 0, End: 5}, {Start: 30, End: 35}}
	// TOTPDefaultBetween is the default color window as a string
	TOTPDefaultBetween = toString(TOTPDefaultColorWindow)
)

type (
	// ColorWindow for handling terminal colors based on timing
	ColorWindow struct {
		Start int
		End   int
	}
)

func toString(windows []ColorWindow) string {
	var results []string
	for _, w := range windows {
		results = append(results, fmt.Sprintf("%d%s%d", w.Start, colorWindowSpan, w.End))
	}
	return strings.Join(results, colorWindowDelimiter)
}

// ParseColorWindow will handle parsing a window of colors for TOTP operations
func ParseColorWindow(windowString string) ([]ColorWindow, error) {
	var rules []ColorWindow
	for _, item := range strings.Split(windowString, colorWindowDelimiter) {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		parts := strings.Split(line, colorWindowSpan)
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
