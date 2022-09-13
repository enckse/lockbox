// provides the binary runs or calls lockbox commands.
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
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/store"
	"github.com/enckse/lockbox/internal/subcommands"
)

var (
	version = "development"
)

type (
	callbackFunction func([]string) error
)

func getEntry(fs store.FileSystem, args []string, idx int) string {
	if len(args) != idx+1 {
		die("invalid entry given", errors.New("specific entry required"))
	}
	return fs.NewPath(args[idx])
}

func internalCallback(name string) callbackFunction {
	switch name {
	case "gitdiff":
		return subcommands.GitDiff
	case "rekey":
		return subcommands.Rekey
	case "rw":
		return subcommands.ReadWrite
	case "totp":
		return subcommands.TOTP
	}
	return nil
}

func die(message string, err error) {
	msg := message
	if err != nil {
		msg = fmt.Sprintf("%s (%v)", msg, err)
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func main() {
	args := os.Args
	if len(args) < 2 {
		die("missing arguments", errors.New("requires subcommand"))
	}
	command := args[1]
	switch command {
	case "ls", "list", "find":
		opts := subcommands.ListFindOptions{Find: command == "find", Search: "", Store: store.NewFileSystemStore()}
		if opts.Find {
			if len(args) < 3 {
				die("find requires an argument to search for", errors.New("search term required"))
			}
			opts.Search = args[2]
		}
		files, err := subcommands.ListFindCallback(opts)
		if err != nil {
			die("unable to list files", err)
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
			die("insert missing required arguments", errors.New("entry required"))
		case 3:
		case 4:
			options = cli.ParseArgs(args[2])
			if !options.Multi {
				die("multi-line insert must be after 'insert'", errors.New("invalid command"))
			}
			idx = 3
		default:
			die("too many arguments", errors.New("insert can only perform one operation"))
		}
		isPipe := inputs.IsInputFromPipe()
		entry := getEntry(store.NewFileSystemStore(), args, idx)
		if store.PathExists(entry) {
			if !isPipe {
				if !confirm("overwrite existing") {
					return
				}
			}
		} else {
			dir := filepath.Dir(entry)
			if !store.PathExists(dir) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					die("failed to create directory structure", err)
				}
			}
		}
		password, err := inputs.GetUserInputPassword(isPipe, options.Multi)
		if err != nil {
			die("invalid input", err)
		}
		if err := encrypt.ToFile(entry, password); err != nil {
			die("unable to encrypt object", err)
		}
		fmt.Println("")
		hooks.Run(hooks.Insert, hooks.PostStep)
	case "rm":
		entry := getEntry(store.NewFileSystemStore(), args, 2)
		if !store.PathExists(entry) {
			die("does not exists", errors.New("can not delete unknown entry"))
		}
		if confirm("remove entry") {
			os.Remove(entry)
			hooks.Run(hooks.Remove, hooks.PostStep)
		}
	case "show", "clip", "dump":
		fs := store.NewFileSystemStore()
		opts := subcommands.DisplayOptions{Dump: command == "dump", Show: command == "show", Glob: getEntry(fs, []string{"***"}, 0), Store: fs}
		opts.Show = opts.Show || opts.Dump
		startEntry := 2
		options := cli.Arguments{}
		if opts.Dump {
			if len(args) > 2 {
				options = cli.ParseArgs(args[2])
				if options.Yes {
					startEntry = 3
				}
			}
		}
		opts.Entry = getEntry(fs, args, startEntry)
		var err error
		dumpData, err := subcommands.DisplayCallback(opts)
		if err != nil {
			die("display command failed to retrieve data", err)
		}
		if opts.Dump {
			if !options.Yes {
				if !confirm("dump data to stdout as plaintext") {
					return
				}
			}
			d, err := dump.Marshal(dumpData)
			if err != nil {
				die("failed to marshal items", err)
			}
			fmt.Println(string(d))
			return
		}
		clipboard := platform.Clipboard{}
		if !opts.Show {
			clipboard, err = platform.NewClipboard()
			if err != nil {
				die("unable to get clipboard", err)
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
			if err := clipboard.CopyTo(obj.Value); err != nil {
				die("clipboard failed", err)
			}
		}
	case "clear":
		if err := subcommands.ClearClipboardCallback(); err != nil {
			die("failed to handle clipboard clear", err)
		}
	default:
		a := args[2:]
		callback := internalCallback(command)
		if callback != nil {
			if err := callback(a); err != nil {
				die(fmt.Sprintf("%s command failure", command), err)
			}
			return
		}
		die("unknown command", errors.New(command))
	}
}

func confirm(prompt string) bool {
	yesNo, err := inputs.ConfirmYesNoPrompt(prompt)
	if err != nil {
		die("failed to get response", err)
	}
	return yesNo
}
