package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"voidedtech.com/lockbox/internal"
	"voidedtech.com/stock"
)

var (
	version = "development"
)

func getEntry(store string, args []string, idx int) string {
	if len(args) != idx+1 {
		stock.Die("invalid entry given", internal.NewLockboxError("specific entry required"))
	}
	return filepath.Join(store, args[idx]) + internal.Extension
}

func termEcho(on bool) {
	// Common settings and variables for both stty calls.
	attrs := syscall.ProcAttr{
		Dir:   "",
		Env:   []string{},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		Sys:   nil}
	var ws syscall.WaitStatus
	cmd := "echo"
	if !on {
		cmd = "-echo"
	}

	// Enable/disable echoing.
	pid, err := syscall.ForkExec(
		"/bin/stty",
		[]string{"stty", cmd},
		&attrs)
	if err != nil {
		panic(err)
	}

	// Wait for the stty process to complete.
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		panic(err)
	}
}

func readInput() (string, error) {
	termEcho(false)
	defer func() {
		termEcho(true)
	}()
	fmt.Printf("please enter password: ")
	first, err := stdin(true)
	if err != nil {
		return "", err
	}
	fmt.Printf("\nplease re-enter password: ")
	second, err := stdin(true)
	if err != nil {
		return "", err
	}
	if first != second {
		return "", internal.NewLockboxError("passwords do NOT match")
	}
	return first, nil
}

func main() {
	args := os.Args
	if len(args) < 2 {
		stock.Die("missing arguments", internal.NewLockboxError("requires subcommand"))
	}
	command := args[1]
	store := internal.GetStore()
	switch command {
	case "ls", "list", "find":
		isFind := command == "find"
		searchTerm := ""
		if isFind {
			if len(args) < 3 {
				stock.Die("find requires an argument to search for", internal.NewLockboxError("search term required"))
			}
			searchTerm = args[2]
		}
		files, err := internal.Find(store, true)
		if err != nil {
			stock.Die("unable to list files", err)
		}
		for _, f := range files {
			if isFind {
				if !strings.Contains(f, searchTerm) {
					continue
				}
			}
			fmt.Println(f)
		}
	case "version":
		fmt.Printf("version: %s\n", version)
	case "insert":
		multi := false
		idx := 2
		switch len(args) {
		case 2:
			stock.Die("insert missing required arguments", internal.NewLockboxError("entry required"))
		case 3:
		case 4:
			multi = args[2] == "-m"
			if !multi {
				stock.Die("multi-line insert must be after 'insert'", internal.NewLockboxError("invalid command"))
			}
			idx = 3
		default:
			stock.Die("too many arguments", internal.NewLockboxError("insert can only perform one operation"))
		}
		isPipe := isInputFromPipe()
		entry := getEntry(store, args, idx)
		if stock.PathExists(entry) {
			if !isPipe {
				if !confirm("overwrite existing") {
					return
				}
			}
		} else {
			dir := filepath.Dir(entry)
			if !stock.PathExists(dir) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					stock.Die("failed to create directory structure", err)
				}
			}
		}
		var password string
		if !multi && !isPipe {
			input, err := readInput()
			if err != nil {
				stock.Die("password input failed", err)
			}
			password = input
		} else {
			input, err := stdin(false)
			if err != nil {
				stock.Die("failed to read stdin", err)
			}
			password = input
		}
		if password == "" {
			stock.Die("empty password provided", internal.NewLockboxError("password can NOT be empty"))
		}
		l, err := internal.NewLockbox("", "", entry)
		if err != nil {
			stock.Die("unable to make lockbox model instance", err)
		}
		if err := l.Encrypt([]byte(password)); err != nil {
			stock.Die("failed to save password", err)
		}
		fmt.Println("")
	case "rm":
		entry := getEntry(store, args, 2)
		if !stock.PathExists(entry) {
			stock.Die("does not exists", internal.NewLockboxError("can not delete unknown entry"))
		}
		if confirm("remove entry") {
			os.Remove(entry)
		}
	case "show", "-c", "clip":
		inEntry := getEntry(store, args, 2)
		isShow := command == "show"
		entries := []string{inEntry}
		if strings.Contains(inEntry, "*") {
			if !isShow {
				stock.Die("cannot glob to clipboard", internal.NewLockboxError("bad glob request"))
			}
			matches, err := filepath.Glob(inEntry)
			if err != nil {
				stock.Die("bad glob", err)
			}
			entries = matches
		}
		isGlob := len(entries) > 1
		for _, entry := range entries {
			if !stock.PathExists(entry) {
				stock.Die("invalid entry", internal.NewLockboxError("entry not found"))
			}
			l, err := internal.NewLockbox("", "", entry)
			if err != nil {
				stock.Die("unable to make lockbox model instance", err)
			}
			decrypt, err := l.Decrypt()
			if err != nil {
				stock.Die("unable to decrypt", err)
			}
			value := strings.TrimSpace(string(decrypt))
			if isShow {
				if isGlob {
					fileName := strings.ReplaceAll(entry, store, "")
					if fileName[0] == '/' {
						fileName = fileName[1:]
					}
					fileName = strings.ReplaceAll(fileName, internal.Extension, "")
					fmt.Printf("%s%s:%s\n", internal.TermBeginRed, fileName, internal.TermEndRed)
				}
				fmt.Println(value)
				continue
			}
			internal.CopyToClipboard(value)
		}
	case "clear":
		idx := 0
		val, err := stdin(false)
		if err != nil {
			stock.Die("unable to read value to clear", err)
		}
		val = strings.TrimSpace(val)
		for idx < internal.MaxClipTime {
			idx++
			time.Sleep(1 * time.Second)
			out, err := exec.Command("pbpaste").Output()
			if err != nil {
				continue
			}
			fmt.Println(string(out))
			fmt.Println(val)
			if strings.TrimSpace(string(out)) != val {
				return
			}
		}
		internal.CopyToClipboard("")
	default:
		tryCommand := fmt.Sprintf("lb-%s", command)
		c := exec.Command(tryCommand, args[2:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			stock.Die("bad command", err)
		}
	}
}

func stdin(one bool) (string, error) {
	b, err := stock.Stdin(one)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func confirm(prompt string) bool {
	fmt.Printf("%s? (y/N) ", prompt)
	resp, err := stdin(true)
	if err != nil {
		stock.Die("failed to get response", err)
	}
	return resp == "Y" || resp == "y"
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
