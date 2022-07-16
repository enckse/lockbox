package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/colors"
	"github.com/enckse/lockbox/internal/dump"
	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/hooks"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/misc"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/store"
)

var (
	version = "development"
	libExec = ""
)

func getEntry(fs store.FileSystem, args []string, idx int) string {
	if len(args) != idx+1 {
		misc.Die("invalid entry given", errors.New("specific entry required"))
	}
	return fs.NewPath(args[idx])
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
	fs := store.NewFileSystemStore()
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
		viewOptions := store.ViewOptions{Display: true}
		if isFind {
			viewOptions.Filter = func(inPath string) string {
				if strings.Contains(inPath, searchTerm) {
					return inPath
				}
				return ""
			}
		}
		files, err := fs.List(store.ViewOptions{Display: true})
		if err != nil {
			misc.Die("unable to list files", err)
		}
		for _, f := range files {
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
		isPipe := inputs.IsInputFromPipe()
		entry := getEntry(fs, args, idx)
		if misc.PathExists(entry) {
			if !isPipe {
				if !confirm("overwrite existing") {
					return
				}
			}
		} else {
			dir := filepath.Dir(entry)
			if !misc.PathExists(dir) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					misc.Die("failed to create directory structure", err)
				}
			}
		}
		var password string
		if !options.Multi && !isPipe {
			input, err := inputs.ConfirmInputsMatch("password")
			if err != nil {
				misc.Die("password input failed", err)
			}
			password = input
		} else {
			input, err := inputs.Stdin(false)
			if err != nil {
				misc.Die("failed to read stdin", err)
			}
			password = input
		}
		if password == "" {
			misc.Die("empty password provided", errors.New("password can NOT be empty"))
		}
		if err := encrypt.ToFile(entry, []byte(password)); err != nil {
			misc.Die("unable to encrypt object", err)
		}
		fmt.Println("")
		hooks.Run(hooks.Insert, hooks.PostStep)
	case "rm":
		entry := getEntry(fs, args, 2)
		if !misc.PathExists(entry) {
			misc.Die("does not exists", errors.New("can not delete unknown entry"))
		}
		if confirm("remove entry") {
			os.Remove(entry)
			hooks.Run(hooks.Remove, hooks.PostStep)
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
		inEntry := getEntry(fs, args, startEntry)
		isShow := command == "show" || isDump
		entries := []string{inEntry}
		if strings.Contains(inEntry, "*") {
			if inEntry == getEntry(fs, []string{"***"}, 0) {
				all, err := fs.List(store.ViewOptions{})
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
		coloring, err := colors.NewTerminal(colors.Red)
		if err != nil {
			misc.Die("unable to get color for terminal", err)
		}
		dumpData := []dump.ExportEntity{}
		clipboard := platform.Clipboard{}
		if !isShow {
			clipboard, err = platform.NewClipboard()
			if err != nil {
				misc.Die("unable to get clipboard", err)
			}
		}
		for _, entry := range entries {
			if !misc.PathExists(entry) {
				misc.Die("invalid entry", errors.New("entry not found"))
			}
			l, err := encrypt.NewLockbox(encrypt.LockboxOptions{File: entry})
			if err != nil {
				misc.Die("unable to make lockbox model instance", err)
			}
			decrypt, err := l.Decrypt()
			if err != nil {
				misc.Die("unable to decrypt", err)
			}
			value := strings.TrimSpace(string(decrypt))
			entity := dump.ExportEntity{}
			if isShow {
				if isGlob {
					fileName := fs.CleanPath(entry)
					if isDump {
						entity.Path = fileName
					} else {
						fmt.Printf("%s%s:%s\n", coloring.Start, fileName, coloring.End)
					}
				}
				if isDump {
					entity.Value = value
					dumpData = append(dumpData, entity)
				} else {
					fmt.Println(value)
				}
				continue
			}
			clipboard.CopyTo(value, getExecutable())
		}
		if isDump {
			if !options.Yes {
				if !confirm("dump data to stdout as plaintext") {
					return
				}
			}
			d, err := dump.Marshal(dumpData)
			if err != nil {
				misc.Die("failed to marshal items", err)
			}
			fmt.Println(string(d))
		}
	case "clear":
		idx := 0
		val, err := inputs.Stdin(false)
		if err != nil {
			misc.Die("unable to read value to clear", err)
		}
		clipboard, err := platform.NewClipboard()
		if err != nil {
			misc.Die("unable to get paste command", err)
		}
		pCmd, pArgs := clipboard.Args(false)
		val = strings.TrimSpace(val)
		for idx < clipboard.MaxTime {
			idx++
			time.Sleep(1 * time.Second)
			out, err := exec.Command(pCmd, pArgs...).Output()
			if err != nil {
				continue
			}
			fmt.Println(string(out))
			fmt.Println(val)
			if strings.TrimSpace(string(out)) != val {
				return
			}
		}
		clipboard.CopyTo("", getExecutable())
	default:
		lib := os.Getenv("LOCKBOX_LIBEXEC")
		if lib == "" {
			lib = libExec
		}
		tryCommand := fmt.Sprintf(filepath.Join(lib, "lb-%s"), command)
		if !misc.PathExists(tryCommand) {
			misc.Die("unknown subcommand", errors.New(command))
		}
		c := exec.Command(tryCommand, args[2:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			misc.Die("bad command", err)
		}
	}
}

func confirm(prompt string) bool {
	yesNo, err := inputs.ConfirmYesNoPrompt(prompt)
	if err != nil {
		misc.Die("failed to get response", err)
	}
	return yesNo
}
