// Package test has some integration tests for the binary lb variant
package test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/enckse/pgl/os/exit"
	"github.com/enckse/pgl/os/paths"
)

var yes = []string{"y"}

const (
	testKey   = "plaintextkey"
	clipRetry = 3
	clipWait  = 1000
)

type (
	testRunner struct {
		file *os.File
		exe  string
	}
)

func (r testRunner) runCommand(args []string, data []string) {
	p := exec.Command(r.exe, args...)
	var buf bytes.Buffer
	for _, d := range data {
		if _, err := buf.WriteString(fmt.Sprintf("%s\n", d)); err != nil {
			exit.Dief("failed to write stdin: %v", err)
		}
	}
	p.Stdout = r.file
	p.Stderr = r.file
	p.Stdin = &buf
	if err := p.Run(); err != nil {
		fmt.Fprintf(r.file, "%v\n", err)
	}
}

func (r testRunner) ls() {
	r.runCommand([]string{"ls"}, nil)
}

func (r testRunner) rm(k string) {
	r.runCommand([]string{"rm", k}, yes)
}

func (r testRunner) show(k string) {
	r.runCommand([]string{"show", k}, nil)
}

func (r testRunner) insert(k string, d []string) {
	r.runCommand([]string{"insert", k}, d)
}

func (r testRunner) totpList() {
	r.runCommand([]string{"totp", "-list"}, nil)
}

func (r testRunner) ln() {
	r.file.Write([]byte("\n"))
}

func replace(input string, re *regexp.Regexp, to string) string {
	matches := re.FindAllStringSubmatch(input, -1)
	res := input
	for _, match := range matches {
		for _, m := range match {
			res = strings.ReplaceAll(res, m, to)
		}
	}
	return res
}

// Cleanup will cleanup the data log outputs
func Cleanup(dataFile string) error {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}
	totp, err := regexp.Compile("^[0-9][0-9][0-9][0-9][0-9][0-9]$")
	if err != nil {
		return err
	}
	date := fmt.Sprintf("modtime: %s", time.Now().Format("2006-01-02"))
	var results []string
	for _, l := range strings.Split(string(data), "\n") {
		payload := l
		payload = replace(payload, totp, "XXXXXX")
		if strings.Contains(payload, date) {
			prefix := ""
			if strings.HasPrefix(payload, "  ") {
				prefix = "  "
			}
			payload = fmt.Sprintf("%s%s", prefix, "modtime: XXXX-XX-XX")
		}
		results = append(results, payload)
	}
	return os.WriteFile(dataFile, []byte(strings.Join(results, "\n")), 0o644)
}

// Execute will run a test
func Execute(keyFile bool, exe, dataPath, logFile string) error {
	f, err := os.Create(logFile)
	if err != nil {
		return err
	}
	defer f.Close()
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	useKeyFile := ""
	if keyFile {
		useKeyFile = filepath.Join(dataPath, "test.key")
		if err := os.WriteFile(useKeyFile, []byte("thisisatest"), 0o644); err != nil {
			return err
		}
	}
	store := filepath.Join(dataPath, fmt.Sprintf("%s.kdbx", time.Now().Format("20060102150405")))
	os.Setenv("LOCKBOX_HOOKDIR", "")
	os.Setenv("LOCKBOX_STORE", store)
	os.Setenv("LOCKBOX_KEY", testKey)
	os.Setenv("LOCKBOX_TOTP", "totp")
	os.Setenv("LOCKBOX_INTERACTIVE", "no")
	os.Setenv("LOCKBOX_READONLY", "no")
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_KEYFILE", useKeyFile)
	runner := testRunner{file: f, exe: exe}
	runner.insert("keys/k/one2", []string{"test2"})
	for _, k := range []string{"keys/k/one", "key/a/one", "keys/k/one", "keys/k/one/", "/keys/k/one", "keys/aa/b//s///e"} {
		runner.insert(k, []string{"test"})
	}
	runner.insert("keys2/k/three", []string{"test3", "test4"})
	runner.ls()
	runner.rm("keys/k/one")
	runner.ln()
	runner.ls()
	runner.runCommand([]string{"find", "e"}, nil)
	runner.show("keys/k/one2")
	runner.show("keys2/k/three")
	runner.runCommand([]string{"stats", "keys2/k/three"}, nil)
	for _, k := range []string{"test/k", "test/k/totp"} {
		runner.runCommand([]string{"insert", "-totp", k}, []string{"5ae472abqdekjqykoyxk7hvc2leklq5n"})
	}
	runner.totpList()
	runner.runCommand([]string{"totp", "test/k"}, nil)
	runner.runCommand([]string{"hash", store}, nil)
	runner.rm("keys2/k/three")
	runner.ln()
	runner.rm("test/k/totp")
	runner.ln()
	runner.rm("test/k/one")
	runner.ln()
	runner.ln()
	runner.runCommand([]string{"mv", "key/a/one", "keyx/d/e"}, nil)
	runner.ls()
	runner.rm("keyx/d/e")
	runner.ln()
	runner.ls()
	runner.insert("keys/k2/one2", []string{"test2"})
	runner.insert("keys/k2/one", []string{"test"})
	runner.insert("keys/k2/t1/one2", []string{"test2"})
	runner.insert("keys/k2/t1/one", []string{"test"})
	runner.insert("keys/k2/t2/one2", []string{"test2"})
	os.Setenv("LOCKBOX_HOOKDIR", filepath.Join(cwd, "hooks"))
	runner.insert("keys/k2/t2/one", []string{"test"})
	runner.ln()
	runner.ls()
	runner.rm("keys/k2/t1/*")
	runner.ln()
	runner.ls()
	runner.rm("keys/k2/*")
	runner.ln()
	runner.ls()
	runner.ln()
	reKeyStore := fmt.Sprintf("%s.rekey.kdbx", store)
	reKey := "rekey"
	os.Setenv("LOCKBOX_STORE_NEW", reKeyStore)
	os.Setenv("LOCKBOX_KEY_NEW", reKey)
	os.Setenv("LOCKBOX_KEYMODE_NEW", "plaintext")
	os.Setenv("LOCKBOX_KEYFILE_NEW", "")
	runner.runCommand([]string{"rekey"}, yes)
	os.Setenv("LOCKBOX_STORE", reKeyStore)
	os.Setenv("LOCKBOX_KEYFILE", "")
	os.Setenv("LOCKBOX_KEY", reKey)
	runner.ln()
	runner.ls()
	return runner.clipboard(dataPath)
}

func (r testRunner) clipboard(dataPath string) error {
	clipCopyFile := filepath.Join(dataPath, "clipboard")
	clipPasteFile := clipCopyFile + ".paste"
	clipFiles := []string{clipCopyFile, clipPasteFile}
	os.Setenv("LOCKBOX_CLIP_COPY", fmt.Sprintf("touch %s", clipCopyFile))
	os.Setenv("LOCKBOX_CLIP_PASTE", fmt.Sprintf("touch %s", clipPasteFile))
	os.Setenv("LOCKBOX_CLIP_MAX", "5")
	r.runCommand([]string{"clip", "keys/k/one2"}, nil)
	clipDur := time.Duration(clipWait) * time.Millisecond
	tries := clipRetry
	for {
		if tries == 0 {
			return errors.New("missing clipboard files")
		}
		foundClipCount := 0
		for _, f := range clipFiles {
			if paths.Exist(f) {
				foundClipCount++
			}
		}
		if foundClipCount == len(clipFiles) {
			break
		}
		time.Sleep(clipDur)
		tries--
	}
	return nil
}
