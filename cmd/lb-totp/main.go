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
	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/misc"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/store"
	otp "github.com/pquerna/otp/totp"
)

var (
	mainExe = ""
)

func list() ([]string, error) {
	f := store.NewFileSystemStore()
	token := totpToken(f, true)
	results, err := f.List(store.ViewOptions{Filter: func(path string) string {
		if filepath.Base(path) == token {
			return filepath.Dir(f.CleanPath(path))
		}
		return ""
	}})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no objects found")
	}
	return results, nil
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Printf("unable to clear screen: %v\n", err)
	}
}

func totpToken(f store.FileSystem, extension bool) string {
	t := os.Getenv("LOCKBOX_TOTP")
	if t == "" {
		t = "totp"
	}
	if !extension {
		return t
	}
	return f.NewFile(t)
}

func display(token string, args cli.Arguments) error {
	interactive, err := inputs.IsInteractive()
	if err != nil {
		return err
	}
	if args.Short {
		interactive = false
	}
	if !interactive && args.Clip {
		return errors.New("clipboard not available in non-interactive mode")
	}
	coloring, err := colors.NewTerminal(colors.Red)
	if err != nil {
		return err
	}
	f := store.NewFileSystemStore()
	tok := filepath.Join(strings.TrimSpace(token), totpToken(f, false))
	pathing := f.NewPath(tok)
	if !misc.PathExists(pathing) {
		return errors.New("object does not exist")
	}
	val, err := encrypt.FromFile(pathing)
	if err != nil {
		return err
	}
	exe := os.Getenv("LOCKBOX_EXE")
	if exe == "" {
		exe = mainExe
	}
	totpToken := string(val)
	if !interactive {
		code, err := otp.GenerateCode(totpToken, time.Now())
		if err != nil {
			return err
		}
		fmt.Println(code)
		return nil
	}
	first := true
	running := 0
	lastSecond := -1
	if !args.Clip {
		if !args.Once {
			clear()
		}
	}
	clipboard := platform.Clipboard{}
	if args.Clip {
		clipboard, err = platform.NewClipboard()
		if err != nil {
			misc.Die("invalid clipboard", err)
		}
	}
	for {
		if !first {
			time.Sleep(500 * time.Millisecond)
		}
		first = false
		running++
		if running > 120 {
			fmt.Println("exiting (timeout)")
			return nil
		}
		now := time.Now()
		last := now.Second()
		if last == lastSecond {
			continue
		}
		lastSecond = last
		left := 60 - last
		code, err := otp.GenerateCode(totpToken, now)
		if err != nil {
			return err
		}
		startColor := ""
		endColor := ""
		if left < 5 || (left < 35 && left >= 30) {
			startColor = coloring.Start
			endColor = coloring.End
		}
		leftString := fmt.Sprintf("%d", left)
		if len(leftString) < 2 {
			leftString = "0" + leftString
		}
		expires := fmt.Sprintf("%s%s (%s)%s", startColor, now.Format("15:04:05"), leftString, endColor)
		outputs := []string{expires}
		if !args.Clip {
			outputs = append(outputs, fmt.Sprintf("%s\n    %s", tok, code))
			if !args.Once {
				outputs = append(outputs, "-> CTRL+C to exit")
			}
		} else {
			fmt.Printf("-> %s\n", expires)
			clipboard.CopyTo(code, exe)
			return nil
		}
		if !args.Once {
			clear()
		}
		fmt.Printf("%s\n", strings.Join(outputs, "\n\n"))
		if args.Once {
			return nil
		}
	}
}

func main() {
	args := os.Args
	if len(args) > 3 || len(args) < 2 {
		misc.Die("subkey required", errors.New("invalid arguments"))
	}
	cmd := args[1]
	options := cli.ParseArgs(cmd)
	if options.List {
		result, err := list()
		if err != nil {
			misc.Die("invalid list response", err)
		}
		sort.Strings(result)
		for _, entry := range result {
			fmt.Println(entry)
		}
		return
	}
	if len(args) == 3 {
		if !options.Clip && !options.Short && !options.Once {
			misc.Die("subcommand not supported", errors.New("invalid sub command"))
		}
		cmd = args[2]
	}
	if err := display(cmd, options); err != nil {
		misc.Die("failed to show totp token", err)
	}
}
