package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal"
)

var (
	version = "development"
)

func getEntry(store string, args []string, idx int) string {
	if len(args) != idx+1 {
		internal.Die("invalid entry given", internal.NewLockboxError("specific entry required"))
	}
	return filepath.Join(store, args[idx]) + internal.Extension
}

func hooks() {
	hookDir := os.Getenv("LOCKBOX_HOOKDIR")
	if !internal.PathExists(hookDir) {
		return
	}
	dirs, err := os.ReadDir(hookDir)
	if err != nil {
		internal.Die("unable to read hookdir", err)
	}
	for _, d := range dirs {
		if !d.IsDir() {
			name := d.Name()
			cmd := exec.Command(filepath.Join(hookDir, name))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				internal.Die(fmt.Sprintf("hook failed: %s", name), err)
			}
			continue
		}
		internal.Die("invalid hook", internal.NewLockboxError("hook is not file and/or has wrong mode"))
	}
}

func main() {
	args := os.Args
	if len(args) < 2 {
		internal.Die("missing arguments", internal.NewLockboxError("requires subcommand"))
	}
	command := args[1]
	store := internal.GetStore()
	switch command {
	case "ls", "list", "find":
		isFind := command == "find"
		searchTerm := ""
		if isFind {
			if len(args) < 3 {
				internal.Die("find requires an argument to search for", internal.NewLockboxError("search term required"))
			}
			searchTerm = args[2]
		}
		files, err := internal.Find(store, true)
		if err != nil {
			internal.Die("unable to list files", err)
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
			internal.Die("insert missing required arguments", internal.NewLockboxError("entry required"))
		case 3:
		case 4:
			multi = args[2] == "-m"
			if !multi {
				internal.Die("multi-line insert must be after 'insert'", internal.NewLockboxError("invalid command"))
			}
			idx = 3
		default:
			internal.Die("too many arguments", internal.NewLockboxError("insert can only perform one operation"))
		}
		isPipe := internal.IsInputFromPipe()
		entry := getEntry(store, args, idx)
		if internal.PathExists(entry) {
			if !isPipe {
				if !confirm("overwrite existing") {
					return
				}
			}
		} else {
			dir := filepath.Dir(entry)
			if !internal.PathExists(dir) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					internal.Die("failed to create directory structure", err)
				}
			}
		}
		var password string
		if !multi && !isPipe {
			input, err := internal.ConfirmInput()
			if err != nil {
				internal.Die("password input failed", err)
			}
			password = input
		} else {
			input, err := internal.Stdin(false)
			if err != nil {
				internal.Die("failed to read stdin", err)
			}
			password = input
		}
		if password == "" {
			internal.Die("empty password provided", internal.NewLockboxError("password can NOT be empty"))
		}
		l, err := internal.NewLockbox("", "", entry)
		if err != nil {
			internal.Die("unable to make lockbox model instance", err)
		}
		if err := l.Encrypt([]byte(password)); err != nil {
			internal.Die("failed to save password", err)
		}
		fmt.Println("")
		hooks()
	case "rm":
		entry := getEntry(store, args, 2)
		if !internal.PathExists(entry) {
			internal.Die("does not exists", internal.NewLockboxError("can not delete unknown entry"))
		}
		if confirm("remove entry") {
			os.Remove(entry)
			hooks()
		}
	case "show", "-c", "clip":
		inEntry := getEntry(store, args, 2)
		isShow := command == "show"
		entries := []string{inEntry}
		if strings.Contains(inEntry, "*") {
			matches, err := filepath.Glob(inEntry)
			if err != nil {
				internal.Die("bad glob", err)
			}
			entries = matches
		}
		isGlob := len(entries) > 1
		if isGlob {
			if !isShow {
				internal.Die("cannot glob to clipboard", internal.NewLockboxError("bad glob request"))
			}
			sort.Strings(entries)
		}
		startColor, endColor, err := internal.GetColor(internal.ColorRed)
		if err != nil {
			internal.Die("unable to get color for terminal", err)
		}
		for _, entry := range entries {
			if !internal.PathExists(entry) {
				internal.Die("invalid entry", internal.NewLockboxError("entry not found"))
			}
			l, err := internal.NewLockbox("", "", entry)
			if err != nil {
				internal.Die("unable to make lockbox model instance", err)
			}
			decrypt, err := l.Decrypt()
			if err != nil {
				internal.Die("unable to decrypt", err)
			}
			value := strings.TrimSpace(string(decrypt))
			if isShow {
				if isGlob {
					fileName := strings.ReplaceAll(entry, store, "")
					if fileName[0] == '/' {
						fileName = fileName[1:]
					}
					fileName = strings.ReplaceAll(fileName, internal.Extension, "")
					fmt.Printf("%s%s:%s\n", startColor, fileName, endColor)
				}
				fmt.Println(value)
				continue
			}
			internal.CopyToClipboard(value)
		}
	case "clear":
		idx := 0
		val, err := internal.Stdin(false)
		if err != nil {
			internal.Die("unable to read value to clear", err)
		}
		_, paste, err := internal.GetClipboardCommand()
		if err != nil {
			internal.Die("unable to get paste command", err)
		}
		var args []string
		if len(paste) > 1 {
			args = paste[1:]
		}
		val = strings.TrimSpace(val)
		for idx < internal.MaxClipTime {
			idx++
			time.Sleep(1 * time.Second)
			out, err := exec.Command(paste[0], args...).Output()
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
		exe, err := os.Executable()
		if err != nil {
			internal.Die("unable to get exe", err)
		}
		tryCommand := fmt.Sprintf(filepath.Join(exe, "lb-%s"), command)
		c := exec.Command(tryCommand, args[2:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			internal.Die("bad command", err)
		}
	}
}

func confirm(prompt string) bool {
	fmt.Printf("%s? (y/N) ", prompt)
	resp, err := internal.Stdin(true)
	if err != nil {
		internal.Die("failed to get response", err)
	}
	return resp == "Y" || resp == "y"
}
