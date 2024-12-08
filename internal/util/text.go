package util

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type (
	// Position is the start/end of a word in a greater set
	Position struct {
		Start int
		End   int
	}
	// Word is the text and position in a greater position
	Word struct {
		Text     string
		Position Position
	}
)

// TextWrap performs simple block text word wrapping
func TextWrap(indent uint, in string) string {
	var sections []string
	var cur []string
	for _, line := range strings.Split(strings.TrimSpace(in), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(cur) > 0 {
				sections = append(sections, strings.Join(cur, " "))
				cur = []string{}
			}
			continue
		}
		cur = append(cur, line)
	}
	if len(cur) > 0 {
		sections = append(sections, strings.Join(cur, " "))
	}
	var out bytes.Buffer
	indenting := ""
	var cnt uint
	for cnt < indent {
		indenting = fmt.Sprintf("%s ", indenting)
		cnt++
	}
	indenture := int(80 - indent)
	for _, s := range sections {
		for _, line := range strings.Split(wrap(s, indenture), "\n") {
			fmt.Fprintf(&out, "%s%s\n", indenting, line)
		}
		fmt.Fprint(&out, "\n")
	}
	return out.String()
}

func wrap(in string, maxLength int) string {
	var lines []string
	var cur []string
	for _, p := range strings.Split(in, " ") {
		state := strings.Join(cur, " ")
		l := len(p)
		if len(state)+l >= maxLength {
			lines = append(lines, strings.Join(cur, " "))
			cur = []string{p}
		} else {
			cur = append(cur, p)
		}
	}
	if len(cur) > 0 {
		lines = append(lines, strings.Join(cur, " "))
	}
	return strings.Join(lines, "\n")
}

// TextPositionFields is the displayable set of templated fields
func TextPositionFields() string {
	return strings.Join(readNested(reflect.TypeOf(Word{}), ""), ", ")
}
