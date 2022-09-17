// provides the binary runs or calls lockbox commands.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	programError     struct {
		message string
		err     error
	}
)

func getEntry(fs store.FileSystem, args []string, idx int) string {
	if len(args) != idx+1 {
		exit("invalid entry given", errors.New("specific entry required"))
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

func exit(message string, err error) {
	msg := message
	if err != nil {
		msg = fmt.Sprintf("%s (%v)", msg, err)
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func newError(message string, err error) *programError {
	return &programError{message: message, err: err}
}

func main() {
	if err := run(); err != nil {
		exit(err.message, err.err)
	}
}

func run() *programError {
	args := os.Args
	if len(args) < 2 {
		return newError("missing arguments", errors.New("requires subcommand"))
	}
	command := args[1]
	switch command {
	case "ls", "list", "find":
		opts := subcommands.ListFindOptions{Find: command == "find", Search: "", Store: store.NewFileSystemStore()}
		if opts.Find {
			if len(args) < 3 {
				return newError("find requires an argument to search for", errors.New("search term required"))
			}
			opts.Search = args[2]
		}
		files, err := subcommands.ListFindCallback(opts)
		if err != nil {
			return newError("unable to list files", err)
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
			return newError("insert missing required arguments", errors.New("entry required"))
		case 3:
		case 4:
			options = cli.ParseArgs(args[2])
			if !options.Multi {
				return newError("multi-line insert must be after 'insert'", errors.New("invalid command"))
			}
			idx = 3
		default:
			return newError("too many arguments", errors.New("insert can only perform one operation"))
		}
		isPipe := inputs.IsInputFromPipe()
		s := store.NewFileSystemStore()
		entry := getEntry(s, args, idx)
		if store.PathExists(entry) {
			if !isPipe {
				if !confirm("overwrite existing") {
					return nil
				}
			}
		} else {
			dir := filepath.Dir(entry)
			if !store.PathExists(dir) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return newError("failed to create directory structure", err)
				}
			}
		}
		password, err := inputs.GetUserInputPassword(isPipe, options.Multi)
		if err != nil {
			return newError("invalid input", err)
		}
		if err := encrypt.ToFile(entry, password); err != nil {
			return newError("unable to encrypt object", err)
		}
		fmt.Println("")
		hooks.Run(hooks.Insert, hooks.PostStep)
		if err := s.GitCommit(entry); err != nil {
			return newError("failed to git commit changed", err)
		}
	case "rm":
		s := store.NewFileSystemStore()
		value := args[2]
		var deletes []string
		confirmText := "entry"
		if strings.Contains(value, "*") {
			globs, err := s.Globs(value)
			if err != nil {
				return newError("rm glob failed", err)
			}
			if len(globs) > 1 {
				confirmText = "entries"
			}
			deletes = append(deletes, globs...)
		} else {
			deletes = []string{getEntry(s, args, 2)}
		}
		if len(deletes) == 0 {
			return newError("nothing to delete", errors.New("no files to remove"))
		}
		if confirm(fmt.Sprintf("remove %s", confirmText)) {
			for _, entry := range deletes {
				if !store.PathExists(entry) {
					return newError("does not exists", errors.New("can not delete unknown entry"))
				}
			}
			for _, entry := range deletes {
				if err := os.Remove(entry); err != nil {
					return newError("unable to remove entry", err)
				}
			}
			hooks.Run(hooks.Remove, hooks.PostStep)
			if err := s.GitRemove(deletes); err != nil {
				return newError("failed to git remove", err)
			}
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
			return newError("display command failed to retrieve data", err)
		}
		if opts.Dump {
			if !options.Yes {
				if !confirm("dump data to stdout as plaintext") {
					return nil
				}
			}
			d, err := dump.Marshal(dumpData)
			if err != nil {
				return newError("failed to marshal items", err)
			}
			fmt.Println(string(d))
			return nil
		}
		clipboard := platform.Clipboard{}
		if !opts.Show {
			clipboard, err = platform.NewClipboard()
			if err != nil {
				return newError("unable to get clipboard", err)
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
				return newError("clipboard failed", err)
			}
		}
	case "clear":
		if err := subcommands.ClearClipboardCallback(); err != nil {
			return newError("failed to handle clipboard clear", err)
		}
	default:
		a := args[2:]
		callback := internalCallback(command)
		if callback != nil {
			if err := callback(a); err != nil {
				return newError(fmt.Sprintf("%s command failure", command), err)
			}
			return nil
		}
		return newError("unknown command", errors.New(command))
	}
	return nil
}

func confirm(prompt string) bool {
	yesNo, err := inputs.ConfirmYesNoPrompt(prompt)
	if err != nil {
		exit("failed to get response", err)
	}
	return yesNo
}
