package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/dump"
	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/hooks"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/misc"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/store"
	"github.com/enckse/lockbox/internal/subcommands"
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
		files, err := subcommands.ListFindCallback(subcommands.ListFindOptions{Find: isFind, Search: searchTerm, Store: store.NewFileSystemStore()})
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
		fs := store.NewFileSystemStore()
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
		fs := store.NewFileSystemStore()
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
		fs := store.NewFileSystemStore()
		inEntry := getEntry(fs, args, startEntry)
		opts := subcommands.DisplayOptions{Dump: isDump, Entry: inEntry, Show: command == "show" || isDump, Glob: getEntry(fs, []string{"***"}, 0), Store: fs}
		var err error
		dumpData, err := subcommands.DisplayCallback(opts)
		if err != nil {
			misc.Die("display command failed to retrieve data", err)
		}
		if opts.Dump {
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
			return
		}
		clipboard := platform.Clipboard{}
		if !opts.Show {
			clipboard, err = platform.NewClipboard()
			if err != nil {
				misc.Die("unable to get clipboard", err)
			}
		}
		for _, obj := range dumpData {
			if opts.Show {
				if obj.Path != "" {
					fmt.Println(obj.Path)
				}
				fmt.Println(obj.Value)
				continue
			}
			clipboard.CopyTo(obj.Value, getExecutable())
		}
	case "clear":
		if err := subcommands.ClearClipboardCallback(); err != nil {
			misc.Die("failed to handle clipboard clear", err)
		}
	default:
		lib := os.Getenv("LOCKBOX_LIBEXEC")
		if lib == "" {
			lib = libExec
		}
		if err := subcommands.LibExecCallback(subcommands.LibExecOptions{Directory: lib, Command: command, Args: args[2:]}); err != nil {
			misc.Die("subcommand failed", err)
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
