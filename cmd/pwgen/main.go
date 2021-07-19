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
	allowed := strings.TrimSpace(os.Getenv("PWGEN_ALLOWED"))
	special := strings.TrimSpace(os.Getenv("PWGEN_SPECIAL"))
	transform := *rawTokens
	if len(allowed) == 0 {
		internal.Die("no allowed characters found", fmt.Errorf("allowed characters required"))
	}
	var paths []string
	parts := strings.Split(src, ":")
	for _, p := range parts {
		if internal.PathExists(p) {
			info, err := os.Stat(p)
			if err != nil {
				internal.Die("unable to stat", err)
			}
			if info.IsDir() {
				files, err := os.ReadDir(p)
				if err != nil {
					internal.Die("failed to read directory", err)
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
		internal.Die("no paths found for generation", fmt.Errorf("unable to read paths"))
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
			result = result + string(specials[subChar])
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
				internal.Die("unable to use sed transform without pattern", fmt.Errorf("set PWGEN_SED"))
			}
			cmd := exec.Command("sed", "-e", sedPattern)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				internal.Die("unable to attach stdin to sed", err)
			}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Start(); err != nil {
				internal.Die("failed to run sed", err)
			}
			if _, err := io.WriteString(stdin, name); err != nil {
				stdin.Close()
				internal.Die("write to stdin failed for sed", err)
			}
			stdin.Close()
			if err := cmd.Wait(); err != nil {
				internal.Die("sed failed", err)
			}
			errors := strings.TrimSpace(stderr.String())
			if len(errors) > 0 {
				internal.Die("sed stderr failure", fmt.Errorf(errors))
			}
			name = strings.TrimSpace(stdout.String())
		case transformModeNone:
			break
		default:
			internal.Die("unknown transform mode", fmt.Errorf(transform))
		}
		result = result + name
	}
	fmt.Println(result[0:l])
}
