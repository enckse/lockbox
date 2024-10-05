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

	"github.com/seanenck/lockbox/internal/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GeneratePassword generates a password
func GeneratePassword(cmd CommandOptions) error {
	ok, err := config.EnvNoPasswordGen.Get()
	if err != nil {
		return err
	}
	if ok {
		return errors.New("password generation is disabled")
	}
	length, err := config.EnvPasswordGenCount.Get()
	if err != nil {
		return err
	}
	if length < 1 {
		return fmt.Errorf("word count must be >= 1")
	}
	tmplString := config.EnvPasswordGenTemplate.Get()
	tmplString = strings.ReplaceAll(tmplString, config.TemplateVariable, "$")
	wordList, err := config.EnvPasswordGenWordList.Get()
	if err != nil {
		return err
	}
	if len(wordList) == 0 {
		return errors.New("word list command must set")
	}
	exe := wordList[0]
	var args []string
	if len(wordList) > 1 {
		args = wordList[1:]
	}
	capitalize, err := config.EnvPasswordGenTitle.Get()
	if err != nil {
		return err
	}
	wordResults, err := exec.Command(exe, args...).Output()
	if err != nil {
		return err
	}
	lang, err := language.Parse(config.EnvLanguage.Get())
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
	caser := cases.Title(lang)
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
		if capitalize {
			use = caser.String(use)
		}
		choices = append(choices, use)
	}
	found := len(choices)
	if found < length {
		return errors.New("choices <= word count requested")
	}
	if found > 1 {
		l := found - 1
		for i := 0; i <= l; i++ {
			n := rand.Intn(l)
			x := choices[i]
			choices[i] = choices[n]
			choices[n] = x
		}
	}
	type word struct {
		Text   string
		Length int
	}
	var selected []word
	cnt := 0
	totalLength := 0
	for cnt < length {
		w := word{choices[cnt], totalLength}
		selected = append(selected, w)
		totalLength += len(w.Text)
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
