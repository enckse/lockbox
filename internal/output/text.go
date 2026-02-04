package output

import (
	"bytes"
	"fmt"
	"strings"
)

// TextWrap performs simple block text word wrapping
func TextWrap(indent uint, in string) string {
	var sections []string
	var cur []string
	for line := range strings.SplitSeq(strings.TrimSpace(in), "\n") {
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
	var indenting string
	var cnt uint
	for cnt < indent {
		indenting = fmt.Sprintf("%s ", indenting)
		cnt++
	}
	indenture := int(80 - indent)
	for _, s := range sections {
		for line := range strings.SplitSeq(wrap(s, indenture), "\n") {
			fmt.Fprintf(&out, "%s%s\n", indenting, line)
		}
		fmt.Fprint(&out, "\n")
	}
	return out.String()
}

func wrap(in string, maxLength int) string {
	var lines []string
	var cur []string
	for p := range strings.SplitSeq(in, " ") {
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
