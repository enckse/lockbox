// Package app can generate passwords
package app

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os/exec"
	"slices"
	"strings"
	"text/template"

	"git.sr.ht/~enckse/lockbox/internal/config"
)

// GeneratePassword generates a password
func GeneratePassword(cmd CommandOptions) error {
	enabled := config.EnvPasswordGenEnabled.Get()
	if !enabled {
		return errors.New("password generation is disabled")
	}
	length, err := config.EnvPasswordGenWordCount.Get()
	if err != nil {
		return err
	}
	tmplString := config.EnvPasswordGenTemplate.Get()
	wordList := config.EnvPasswordGenWordList.Get()
	if len(wordList) == 0 {
		return errors.New("word list command must set")
	}
	exe := wordList[0]
	var args []string
	if len(wordList) > 1 {
		args = wordList[1:]
	}
	wordResults, err := exec.Command(exe, args...).Output()
	if err != nil {
		return err
	}
	chars := config.EnvPasswordGenChars.Get()
	hasChars := len(chars) > 0
	var allowedChars []rune
	if hasChars {
		for _, c := range chars {
			allowedChars = append(allowedChars, c)
		}
	}
	var choices []string
	for _, line := range strings.Split(string(wordResults), "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		use := line
		if hasChars {
			res := ""
			for _, c := range use {
				if slices.Contains(allowedChars, c) {
					res = fmt.Sprintf("%s%c", res, c)
				}
			}
			if res == "" {
				continue
			}
			use = res
		}
		choices = append(choices, use)
	}
	found := len(choices)
	if found == 0 {
		return errors.New("no sources given")
	}
	var selected []config.Word
	var cnt int64
	totalLength := 0
	for cnt < length {
		choice := choices[rand.Intn(found)]
		textLength := len(choice)
		selected = append(selected, config.Word{Text: choice, Position: config.Position{Start: totalLength, End: totalLength + textLength}})
		totalLength += textLength
		cnt++
	}
	tmpl, err := template.New("t").Parse(tmplString)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, selected); err != nil {
		return err
	}
	if _, err := buf.WriteString("\n"); err != nil {
		return err
	}
	if _, err := cmd.Writer().Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
