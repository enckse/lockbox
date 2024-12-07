// Package util has to assist with some time windowing
package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// TimeWindowDelimiter indicates how windows are split in env/config keys
	TimeWindowDelimiter = " "
	// TimeWindowSpan indicates the delineation between start -> end (start:end)
	TimeWindowSpan = ":"
)

// TimeWindow for handling terminal colors based on timing
type TimeWindow struct {
	Start int
	End   int
}

// ParseTimeWindow will handle parsing a window of colors for TOTP operations
func ParseTimeWindow(windowString string) ([]TimeWindow, error) {
	var rules []TimeWindow
	for _, item := range strings.Split(windowString, TimeWindowDelimiter) {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		parts := strings.Split(line, TimeWindowSpan)
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
		rules = append(rules, TimeWindow{Start: s, End: e})
	}
	if len(rules) == 0 {
		return nil, errors.New("invalid colorization rules for totp, none found")
	}
	return rules, nil
}
