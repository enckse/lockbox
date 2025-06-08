//go:build integration
// +build integration

package main_test

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"git.sr.ht/~enckse/lockbox/internal/platform"
)

const (
	bothProfile    = "both"
	passProfile    = "password"
	keyFileProfile = "keyfile"
	testPass       = "testingpass"
	testKeyData    = "testkey"
	reKeyPass      = "rekeying"
	reKeyKeyData   = "rekeyfile"
	clipWait       = 1
	clipTries      = 6
)

var (
	target = filepath.Join("..", "..", "target")
	binary = filepath.Join(target, "lb")
	//go:embed tests/*
	testingFiles embed.FS
)

type (
	conf   map[string]string
	runner struct {
		log     string
		testDir string
		config  string
		store   string
	}
)

func newRunner(profile string) (runner, error) {
	t := filepath.Join("testdata", profile)
	l := filepath.Join(t, "actual.log")
	wd, err := os.Getwd()
	if err != nil {
		return runner{}, err
	}
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", filepath.Join(wd, target), os.PathListSeparator, os.Getenv("PATH")))
	if err := os.RemoveAll(t); err != nil {
		return runner{}, err
	}
	if err := os.MkdirAll(t, 0o755); err != nil {
		return runner{}, err
	}
	return runner{l, t, filepath.Join(t, "config.toml"), filepath.Join(t, "pass.kdbx")}, nil
}

func TestPass(t *testing.T) {
	run(t, passProfile)
}

func TestKeyFile(t *testing.T) {
	run(t, keyFileProfile)
}

func TestBoth(t *testing.T) {
	run(t, bothProfile)
}

func run(t *testing.T, profile string) {
	if err := test(profile); err != nil {
		t.Errorf("%s failed: %v", profile, err)
	}
}

func setConfig(config string) {
	os.Setenv("LOCKBOX_CONFIG_TOML", config)
}

func (r runner) writeConfig(c conf) {
	var set []string
	for k, v := range c {
		set = append(set, fmt.Sprintf("%s = %s", k, v))
	}
	sort.Strings(set)
	os.WriteFile(r.config, []byte(strings.Join(set, "\n")), 0o644)
}

func (r runner) run(pipeIn, command string) error {
	return r.raw(pipeIn, command, r.log, "&1")
}

func (r runner) raw(pipeIn, command, stdout, stderr string) error {
	text := fmt.Sprintf("%s %s %s >> %s 2>%s", pipeIn, binary, command, stdout, stderr)
	cmd := exec.Command("/bin/sh", "-c", text)
	return cmd.Run()
}

func (r runner) logAppend(command string) error {
	return exec.Command("/bin/sh", "-c", fmt.Sprintf("%s >> %s", command, r.log)).Run()
}

func (r runner) newConf() conf {
	c := make(conf)
	c["store"] = c.quoteString(r.store)
	return c
}

func (c conf) makePass(pass string) string {
	return fmt.Sprintf("[\"%s\"]", pass)
}

func (c conf) quoteString(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

func unpackDir(dir, under string, mode os.FileMode) error {
	dirs, err := testingFiles.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			return fmt.Errorf("unexpected embedded dir: %s", name)
		}
		data, err := testingFiles.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(under, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(under, name), data, mode); err != nil {
			return err
		}
	}
	return nil
}

func test(profile string) error {
	r, err := newRunner(profile)
	if err != nil {
		return err
	}
	if err := unpackDir("tests", r.testDir, 0o644); err != nil {
		return err
	}

	setConfig(filepath.Join(r.testDir, "invalid"))
	if err := r.raw("", "help", "/dev/null", "/dev/null"); err != nil {
		return err
	}
	setConfig(r.config)
	c := r.newConf()
	c["interactive"] = "false"
	keyFile := filepath.Join(r.testDir, "key.file")
	hasPass := profile == passProfile || profile == bothProfile
	hasFile := profile == keyFileProfile || profile == bothProfile
	if hasPass {
		c["credentials.password"] = c.makePass(testPass)
		c["credentials.password_mode"] = c.quoteString("plaintext")
	}
	if hasFile {
		os.WriteFile(keyFile, []byte(testKeyData), 0o644)
		c["credentials.key_file"] = c.quoteString(keyFile)
		if !hasPass {
			c["credentials.password_mode"] = c.quoteString("none")
		}
	}
	r.writeConfig(c)
	r.run("echo testing |", "insert test1/key1/password")
	if hasPass {
		delete(c, "credentials.password")
		c["interactive"] = "true"
		c["credentials.password_mode"] = c.quoteString("ask")
		r.writeConfig(c)
	} else {
		r.logAppend("printf \"password: \"")
	}
	r.raw(fmt.Sprintf("echo %s |", testPass), "ls", r.log, "/dev/null")
	c = r.newConf()
	c["interactive"] = "false"
	if hasPass {
		c["credentials.password_mode"] = c.quoteString("plaintext")
		c["credentials.password"] = c.makePass(testPass)
	}
	if hasFile {
		c["credentials.key_file"] = c.quoteString(keyFile)
		if !hasPass {
			c["credentials.password_mode"] = c.quoteString("none")
		}
	}
	r.writeConfig(c)
	for _, k := range []string{"test2/key1/password", "test2/key1/notes", "test3", "test3/invalid/", "test3/invalid/still"} {
		r.run("echo testing2 |", fmt.Sprintf("insert %s", k))
	}
	for _, k := range []string{"insert test4/multiline/notes", "insert test5/multiline/notes", "insert test5/multiline/otp", "insert test5/multiline/password"} {
		r.run(`printf "testing3\ntesting4\n" |`, k)
	}
	for _, k := range []string{"insert test6/multiline/password", "insert test6/multiline/notes", "insert test7/deeper/rooted/notes", "insert test7/deeper/rooted/otp", "insert test7/deeper/root/url", "insert test8/unset/password", "insert test8/unset/notes", "insert test9/key1/sub1/password", "insert test9/key1/sub2/password", "insert test9/key2/sub1/password"} {
		r.run(`printf "testing5" |`, k)
		r.run("", fmt.Sprintf("show %s", strings.ReplaceAll(k, "insert ", "")))
	}
	r.run("", "ls")
	r.run("", "groups")
	r.run("echo y |", "rm test2/key1")
	r.logAppend("echo")
	r.run("", "ls")
	r.run("", "ls multiline")
	r.run("", "ls url")
	r.run("", "json")
	r.run("", "json 'multiline'")
	r.logAppend("echo")
	r.run("echo 5ae472abqdekjqykoyxk7hvc2leklq5n |", "insert test6/multiline/otp")
	r.run("", "totp ls")
	r.run("", "totp ls rooted")
	r.run("", "totp show test6/multiline/otp")
	r.run("", "totp once test6/multiline/otp")
	r.run("", "totp minimal test6/multiline/otp")
	r.run("", "totp url test6/multiline/otp")
	r.run("", "totp seed test6/multiline/otp")
	r.run("", fmt.Sprintf("conv \"%s\"", r.store))
	r.run("echo y |", "rm test7/deeper")
	r.run("echo y |", "rm test7/deeper/ro")
	r.run("echo y |", "rm test1/key1/password")
	r.run("echo y |", "rm test1/key1")
	r.logAppend("echo")
	r.run("echo y |", "rm test7/deeper/*")
	r.logAppend("echo")
	r.run("", "ls")
	r.run("", "groups")
	r.run("", "groups test9")
	r.run("echo y |", "unset test8/unset/password")
	r.logAppend("echo")
	r.run("", "ls")
	r.run("", "groups")
	r.run("echo y |", "unset test8/unset/notes")
	r.logAppend("echo")
	r.run("", "ls")
	r.run("", "groups")
	r.run("", "mv test9/key1/* test9/")
	r.run("", "mv test9/key2/sub1 test9/sub3")
	r.run("", "ls")
	r.run("", "groups")
	r.run("echo y |", "rm test9/*")
	r.logAppend("echo")

	// test rekeying
	reKeyArgs := []string{}
	reKeyFile := filepath.Join(r.testDir, "rekey.file")
	if hasFile {
		os.WriteFile(reKeyFile, []byte(reKeyKeyData), 0o644)
		reKeyArgs = append(reKeyArgs, fmt.Sprintf("-keyfile %s", reKeyFile))
		if !hasPass {
			reKeyArgs = append(reKeyArgs, "-nokey")
		}
	}
	r.run(fmt.Sprintf("echo %s |", reKeyPass), fmt.Sprintf("rekey %s", strings.Join(reKeyArgs, " ")))
	if hasPass {
		c["credentials.password"] = c.makePass(reKeyPass)
	}
	if hasFile {
		c["credentials.key_file"] = c.quoteString(reKeyFile)
	}
	r.writeConfig(c)
	r.logAppend("echo")
	r.run("", "ls")
	r.run("", "show test6/multiline/password")

	// test json modes
	c["json.mode"] = c.quoteString("plaintext")
	r.writeConfig(c)
	r.run("", "json test6")
	c["json.mode"] = c.quoteString("empty")
	r.writeConfig(c)
	r.run("", "json test6")
	c["json.mode"] = c.quoteString("hash")
	c["json.hash_length"] = "3"
	r.writeConfig(c)
	r.run("", "json test6")

	// clipboard
	copyFile := filepath.Join(r.testDir, "clip.copy")
	pasteFile := filepath.Join(r.testDir, "clip.paste")
	c["clip.copy_command"] = fmt.Sprintf("[\"touch\", \"%s\"]", copyFile)
	c["clip.paste_command"] = fmt.Sprintf("[\"touch\", \"%s\"]", pasteFile)
	c["clip.timeout"] = "3"
	r.writeConfig(c)
	r.run("", "clip test6/multiline/password")
	clipPassed := false
	tries := 0
	for tries < clipTries {
		if platform.PathExists(copyFile) && platform.PathExists(pasteFile) {
			clipPassed = true
			break
		}
		time.Sleep(500 * clipWait * time.Millisecond)
		tries++
	}
	if !clipPassed {
		return errors.New("clipboard test failed unexpectedly")
	}

	// invalid configuration
	invalid := r.newConf()
	for k, v := range c {
		invalid[k] = v
	}
	if hasPass {
		invalid["credentials.password"] = c.makePass("garbage")
	}
	if hasFile {
		invalidFile := filepath.Join(r.testDir, "bad.file")
		os.WriteFile(invalidFile, []byte{}, 0o644)
		invalid["credentials.key_file"] = c.quoteString(invalidFile)
	}
	r.writeConfig(invalid)
	r.run("", "ls")
	r.writeConfig(c)
	setConfig(filepath.Join(r.testDir, "invalid"))
	r.run("", "ls")
	setConfig(r.config)
	r.run("", "ls")

	// what is env
	r.run("", fmt.Sprintf("vars | sed 's#/%s#/datadir#g' | grep -v CREDENTIALS | sort", profile))

	// cleanup and diff results
	tmpFile := fmt.Sprintf("%s.tmp", r.log)
	for _, item := range []string{"'s/\"modtime\": \"[0-9].*$/\"modtime\": \"XXXX-XX-XX\",/g'", "'s/^[0-9][0-9][0-9][0-9][0-9][0-9]$/XXXXXX/g'"} {
		exec.Command("/bin/sh", "-c", fmt.Sprintf("sed %s %s > %s", item, r.log, tmpFile)).Run()
		exec.Command("mv", tmpFile, r.log).Run()
	}
	diff := exec.Command("diff", "-u", filepath.Join(r.testDir, "expected.log"), r.log)
	diff.Stdout = os.Stdout
	diff.Stderr = os.Stderr
	return diff.Run()
}
