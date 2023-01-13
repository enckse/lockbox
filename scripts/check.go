// package main runs the tests
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	testKey = "plaintextkey"
)

func die(message string, err error) {
	fmt.Fprintf(os.Stderr, "%s (%v)", message, err)
	os.Exit(1)
}

func runCommand(args []string, data []string) {
	p := exec.Command(os.Getenv("LB_BUILD"), args...)
	var buf bytes.Buffer
	for _, d := range data {
		if _, err := buf.WriteString(fmt.Sprintf("%s\n", d)); err != nil {
			die("failed to write stdin", err)
		}
	}
	p.Stdout = os.Stdout
	p.Stderr = os.Stderr
	p.Stdin = &buf
	if err := p.Run(); err != nil {
		fmt.Println(err)
	}
}

func ls() {
	runCommand([]string{"ls"}, nil)
}

func rm(k string) {
	runCommand([]string{"rm", k}, []string{"y"})
}

func show(k string) {
	runCommand([]string{"show", k}, nil)
}

func insert(k string, d []string) {
	runCommand([]string{"insert", k}, d)
}

func totpList() {
	runCommand([]string{"totp", "-list"}, nil)
}

func main() {
	keyFile := flag.Bool("keyfile", false, "enable keyfile")
	flag.Parse()
	path := os.Getenv("TEST_DATA")
	useKeyFile := ""
	if *keyFile {
		useKeyFile = filepath.Join(path, "test.key")
		if err := os.WriteFile(useKeyFile, []byte("thisisatest"), 0o644); err != nil {
			die("unable to write keyfile", err)
		}
	}
	store := filepath.Join(path, fmt.Sprintf("%s.kdbx", time.Now().Format("20060102150405")))
	os.Setenv("LOCKBOX_HOOKDIR", "")
	os.Setenv("LOCKBOX_STORE", store)
	os.Setenv("LOCKBOX_KEY", testKey)
	os.Setenv("LOCKBOX_TOTP", "totp")
	os.Setenv("LOCKBOX_INTERACTIVE", "no")
	os.Setenv("LOCKBOX_READONLY", "no")
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_KEYFILE", useKeyFile)
	insert("keys/k/one2", []string{"test2"})
	for _, k := range []string{"keys/k/one", "key/a/one", "keys/k/one", "keys/k/one/", "/keys/k/one", "keys/aa/b//s///e"} {
		insert(k, []string{"test"})
	}
	insert("keys2/k/three", []string{"test3", "test4"})
	ls()
	rm("keys/k/one")
	fmt.Println()
	ls()
	runCommand([]string{"find", "e"}, nil)
	show("keys/k/one2")
	show("keys2/k/three")
	runCommand([]string{"stats", "keys2/k/three"}, nil)
	for _, k := range []string{"test/k", "test/k/totp"} {
		runCommand([]string{"insert", "-totp", k}, []string{"5ae472abqdekjqykoyxk7hvc2leklq5n"})
	}
	totpList()
	insert("test/k/totp", []string{"5ae472abqdekjqykoyxk7hvc2leklq5n"})
	totpList()
	runCommand([]string{"totp", "test/k"}, nil)
	runCommand([]string{"hash", store}, nil)
	rm("keys2/k/three")
	fmt.Println()
	rm("test/k/totp")
	fmt.Println()
	rm("test/k/one")
	fmt.Println()
	fmt.Println()
	runCommand([]string{"mv", "key/a/one", "keyx/d/e"}, nil)
	ls()
	rm("keyx/d/e")
	fmt.Println()
	ls()
	insert("keys/k2/one2", []string{"test2"})
	insert("keys/k2/one", []string{"test"})
	insert("keys/k2/t1/one2", []string{"test2"})
	insert("keys/k2/t1/one", []string{"test"})
	insert("keys/k2/t2/one2", []string{"test2"})
	os.Setenv("LOCKBOX_HOOKDIR", filepath.Join(os.Getenv("SCRIPTS"), "hooks"))
	insert("keys/k2/t2/one", []string{"test"})
	fmt.Println()
	ls()
	rm("keys/k2/t1/*")
	fmt.Println()
	ls()
	rm("keys/k2/*")
	fmt.Println()
	ls()
}
