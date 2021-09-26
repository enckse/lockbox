package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"voidedtech.com/lockbox/internal"
	"voidedtech.com/stock"
)

const (
	transformModeSed  = "sed"
	transformModePick = "pick"
	transformModeNone = "none"
)

func makeChoice() bool {
	return rand.Intn(2)%2 == 0
}

func main() {
	defaultTransform := transformModePick
	sedPattern := strings.TrimSpace(os.Getenv("PWGEN_SED"))
	if len(sedPattern) > 0 {
		defaultTransform = transformModeSed
	}
	rand.Seed(time.Now().UnixNano())
	length := flag.Int("length", 64, "length of the password to generate")
	extras := flag.Bool("special", false, "include special characters")
	rawTokens := flag.String("transform", defaultTransform, "pick how to transform words")
	flag.Parse()
	src := strings.TrimSpace(os.Getenv("PWGEN_SOURCE"))
	special := strings.TrimSpace(os.Getenv("PWGEN_SPECIAL"))
	transform := *rawTokens
	var paths []string
	parts := strings.Split(src, ":")
	for _, p := range parts {
		if stock.PathExists(p) {
			info, err := os.Stat(p)
			if err != nil {
				stock.Die("unable to stat", err)
			}
			if info.IsDir() {
				files, err := os.ReadDir(p)
				if err != nil {
					stock.Die("failed to read directory", err)
				}
				var results []string
				for _, f := range files {
					results = append(results, f.Name())
				}
				if len(results) > 0 {
					paths = append(paths, results...)
				}
			}
		}
	}
	if len(paths) == 0 {
		stock.Die("no paths found for generation", internal.NewLockboxError("unable to read paths"))
	}
	result := ""
	l := *length
	pathOptions := len(paths)
	specials := []rune{}
	if *extras {
		specials = []rune(special)
	}
	specialChars := len(specials)
	for len(result) < l {
		if specialChars > 0 && makeChoice() {
			subChar := rand.Intn(specialChars)
			result += string(specials[subChar])
		}
		sub := rand.Intn(pathOptions)
		name := paths[sub]
		switch transform {
		case transformModePick:
			newValue := ""
			for _, c := range name {
				if makeChoice() {
					newValue = strings.ToUpper(string(c))
				} else {
					newValue = string(c)
				}
			}
			name = newValue
		case transformModeSed:
			if len(sedPattern) == 0 {
				stock.Die("unable to use sed transform without pattern", internal.NewLockboxError("set PWGEN_SED"))
			}
			cmd := exec.Command("sed", "-e", sedPattern)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				stock.Die("unable to attach stdin to sed", err)
			}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Start(); err != nil {
				stock.Die("failed to run sed", err)
			}
			if _, err := io.WriteString(stdin, name); err != nil {
				stdin.Close()
				stock.Die("write to stdin failed for sed", err)
			}
			stdin.Close()
			if err := cmd.Wait(); err != nil {
				stock.Die("sed failed", err)
			}
			errors := strings.TrimSpace(stderr.String())
			if len(errors) > 0 {
				stock.Die("sed stderr failure", internal.NewLockboxError(errors))
			}
			name = strings.TrimSpace(stdout.String())
		case transformModeNone:
			break
		default:
			stock.Die("unknown transform mode", internal.NewLockboxError(transform))
		}
		result += name
	}
	fmt.Println(result[0:l])
}