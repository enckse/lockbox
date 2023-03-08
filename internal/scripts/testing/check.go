// package main runs the tests
package main

import (
	"bytes"
	"flag"
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
	testKey = "plaintextkey"
)

func runCommand(args []string, data []string) {
	p := exec.Command(os.Getenv("LB_BUILD"), args...)
	var buf bytes.Buffer
	for _, d := range data {
		if _, err := buf.WriteString(fmt.Sprintf("%s\n", d)); err != nil {
			exit.Dief("failed to write stdin: %v", err)
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
	runCommand([]string{"rm", k}, yes)
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
	if err := execute(); err != nil {
		exit.Die(err)
	}
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

func cleanup(dataFile, workDir string) error {
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

func execute() error {
	keyFile := flag.Bool("keyfile", false, "enable keyfile")
	dataPath := flag.String("data", "", "data area")
	runMode := flag.Bool("run", true, "execute tests")
	clipRetry := flag.Uint("clipboard-retry", 3, "number of tries for the clipboard test")
	clipWait := flag.Uint("clip-wait", 1000, "number of milliseconds to wait for the clipboard test")
	flag.Parse()
	path := *dataPath
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if !*runMode {
		return cleanup(path, cwd)
	}
	useKeyFile := ""
	if *keyFile {
		useKeyFile = filepath.Join(path, "test.key")
		if err := os.WriteFile(useKeyFile, []byte("thisisatest"), 0o644); err != nil {
			return err
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
	os.Setenv("LOCKBOX_HOOKDIR", filepath.Join(cwd, "hooks"))
	insert("keys/k2/t2/one", []string{"test"})
	fmt.Println()
	ls()
	rm("keys/k2/t1/*")
	fmt.Println()
	ls()
	rm("keys/k2/*")
	fmt.Println()
	ls()
	fmt.Println()
	reKeyStore := fmt.Sprintf("%s.rekey.kdbx", store)
	reKey := "rekey"
	os.Setenv("LOCKBOX_STORE_NEW", reKeyStore)
	os.Setenv("LOCKBOX_KEY_NEW", reKey)
	os.Setenv("LOCKBOX_KEYMODE_NEW", "plaintext")
	os.Setenv("LOCKBOX_KEYFILE_NEW", "")
	runCommand([]string{"rekey"}, yes)
	os.Setenv("LOCKBOX_STORE", reKeyStore)
	os.Setenv("LOCKBOX_KEYFILE", "")
	os.Setenv("LOCKBOX_KEY", reKey)
	fmt.Println()
	ls()
	testClipboard(path, *clipRetry, *clipWait)
	return nil
}

func testClipboard(dataPath string, tries uint, wait uint) {
	clipCopyFile := filepath.Join(dataPath, "clipboard")
	clipPasteFile := clipCopyFile + ".paste"
	clipFiles := []string{clipCopyFile, clipPasteFile}
	os.Setenv("LOCKBOX_CLIP_COPY", fmt.Sprintf("touch %s", clipCopyFile))
	os.Setenv("LOCKBOX_CLIP_PASTE", fmt.Sprintf("touch %s", clipPasteFile))
	os.Setenv("LOCKBOX_CLIP_MAX", "5")
	runCommand([]string{"clip", "keys/k/one2"}, nil)
	clipDur := time.Duration(wait) * time.Millisecond
	for {
		if tries == 0 {
			fmt.Println("missing clipboard files")
			break
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
}