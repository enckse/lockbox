package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/clipboard"
	"github.com/enckse/lockbox/internal/misc"
	"github.com/enckse/lockbox/internal/store"
)

var (
	version = "development"
	libExec = ""
)

func getEntry(store string, args []string, idx int) string {
	if len(args) != idx+1 {
		misc.Die("invalid entry given", errors.New("specific entry required"))
	}
	return filepath.Join(store, args[idx]) + store.Extension
}

func getExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		misc.Die("unable to get exe", err)
	}
	return exe
}

func main() {
	args := os.Args
	if len(args) < 2 {
		misc.Die("missing arguments", errors.New("requires subcommand"))
	}
	command := args[1]
	store := internal.GetStore()
	switch command {
	case "ls", "list", "find":
		isFind := command == "find"
		searchTerm := ""
		if isFind {
			if len(args) < 3 {
				misc.Die("find requires an argument to search for", errors.New("search term required"))
			}
			searchTerm = args[2]
		}
		files, err := internal.List(store, true)
		if err != nil {
			misc.Die("unable to list files", err)
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
		options := cli.Arguments{}
		idx := 2
		switch len(args) {
		case 2:
			misc.Die("insert missing required arguments", errors.New("entry required"))
		case 3:
		case 4:
			options = cli.ParseArgs(args[2])
			if !options.Multi {
				misc.Die("multi-line insert must be after 'insert'", errors.New("invalid command"))
			}
			idx = 3
		default:
			misc.Die("too many arguments", errors.New("insert can only perform one operation"))
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
					misc.Die("failed to create directory structure", err)
				}
			}
		}
		var password string
		if !options.Multi && !isPipe {
			input, err := internal.ConfirmInputsMatch("password")
			if err != nil {
				misc.Die("password input failed", err)
			}
			password = input
		} else {
			input, err := internal.Stdin(false)
			if err != nil {
				misc.Die("failed to read stdin", err)
			}
			password = input
		}
		if password == "" {
			misc.Die("empty password provided", errors.New("password can NOT be empty"))
		}
		l, err := internal.NewLockbox(internal.LockboxOptions{File: entry})
		if err != nil {
			misc.Die("unable to make lockbox model instance", err)
		}
		if err := l.Encrypt([]byte(password)); err != nil {
			misc.Die("failed to save password", err)
		}
		fmt.Println("")
		internal.Hooks(internal.InsertHook, internal.PostHookStep)
	case "rm":
		entry := getEntry(store, args, 2)
		if !internal.PathExists(entry) {
			misc.Die("does not exists", errors.New("can not delete unknown entry"))
		}
		if confirm("remove entry") {
			os.Remove(entry)
			internal.Hooks(internal.RemoveHook, internal.PostHookStep)
		}
	case "show", "-c", "clip", "dump":
		isDump := command == "dump"
		startEntry := 2
		options := cli.Arguments{}
		if isDump {
			if len(args) > 2 {
				options = cli.ParseArgs(args[2])
				if options.Yes {
					startEntry = 3
				}
			}
		}
		inEntry := getEntry(store, args, startEntry)
		isShow := command == "show" || isDump
		entries := []string{inEntry}
		if strings.Contains(inEntry, "*") {
			if inEntry == getEntry(store, []string{"***"}, 0) {
				all, err := internal.List(store, false)
				if err != nil {
					misc.Die("unable to get all files", err)
				}
				entries = all
			} else {
				matches, err := filepath.Glob(inEntry)
				if err != nil {
					misc.Die("bad glob", err)
				}
				entries = matches
			}
		}
		isGlob := len(entries) > 1
		if isGlob {
			if !isShow {
				misc.Die("cannot glob to clipboard", errors.New("bad glob request"))
			}
			sort.Strings(entries)
		}
		startColor, endColor, err := internal.GetColor(internal.ColorRed)
		if err != nil {
			misc.Die("unable to get color for terminal", err)
		}
		dumpData := []Dump{}
		clip := clipboard.Commands{}
		if !isShow {
			clip, err = clipboard.NewCommands()
			if err != nil {
				misc.Die("unable to get clipboard", err)
			}
		}
		for _, entry := range entries {
			if !internal.PathExists(entry) {
				misc.Die("invalid entry", errors.New("entry not found"))
			}
			l, err := internal.NewLockbox(internal.LockboxOptions{File: entry})
			if err != nil {
				misc.Die("unable to make lockbox model instance", err)
			}
			decrypt, err := l.Decrypt()
			if err != nil {
				misc.Die("unable to decrypt", err)
			}
			value := strings.TrimSpace(string(decrypt))
			dump := Dump{}
			if isShow {
				if isGlob {
					fileName := strings.ReplaceAll(entry, store, "")
					if fileName[0] == '/' {
						fileName = fileName[1:]
					}
					fileName = strings.ReplaceAll(fileName, internal.Extension, "")
					if isDump {
						dump.Path = fileName
					} else {
						fmt.Printf("%s%s:%s\n", startColor, fileName, endColor)
					}
				}
				if isDump {
					dump.Value = value
					dumpData = append(dumpData, dump)
				} else {
					fmt.Println(value)
				}
				continue
			}
			clip.CopyToClipboard(value, getExecutable())
		}
		if isDump {
			if !options.Yes {
				if !confirm("dump data to stdout as plaintext") {
					return
				}
			}
			fmt.Println("[")
			for idx, d := range dumpData {
				if idx > 0 {
					fmt.Println(",")
				}
				b, err := json.MarshalIndent(d, "", "  ")
				if err != nil {
					misc.Die("failed to marshal dump item", err)
				}
				fmt.Println(string(b))
			}
			fmt.Println("]")
		}
	case "clear":
		idx := 0
		val, err := internal.Stdin(false)
		if err != nil {
			misc.Die("unable to read value to clear", err)
		}
		clip, err := clipboard.NewCommands()
		if err != nil {
			misc.Die("unable to get paste command", err)
		}
		var args []string
		if len(clip.Paste) > 1 {
			args = clip.Paste[1:]
		}
		val = strings.TrimSpace(val)
		for idx < clipboard.MaxTime {
			idx++
			time.Sleep(1 * time.Second)
			out, err := exec.Command(clip.Paste[0], args...).Output()
			if err != nil {
				continue
			}
			fmt.Println(string(out))
			fmt.Println(val)
			if strings.TrimSpace(string(out)) != val {
				return
			}
		}
		clip.CopyToClipboard("", getExecutable())
	default:
		lib := os.Getenv("LOCKBOX_LIBEXEC")
		if lib == "" {
			lib = libExec
		}
		tryCommand := fmt.Sprintf(filepath.Join(lib, "lb-%s"), command)
		c := exec.Command(tryCommand, args[2:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			misc.Die("bad command", err)
		}
	}
}

func confirm(prompt string) bool {
	yesNo, err := internal.ConfirmYesNoPrompt(prompt)
	if err != nil {
		misc.Die("failed to get response", err)
	}
	return yesNo
}
