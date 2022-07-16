package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal"
	"github.com/enckse/lockbox/internal/cli"
	otp "github.com/pquerna/otp/totp"
)

var (
	mainExe = ""
)

func list() ([]string, error) {
	files := []string{}
	token := totpToken()
	store := internal.GetStore()
	err := filepath.Walk(store, func(path string, info fs.FileInfo, err error) error {
		name := info.Name()
		if name != token {
			return nil
		}
		dir := strings.TrimPrefix(filepath.Dir(path), store)
		if strings.HasSuffix(dir, "/") {
			dir = dir[0 : len(dir)-1]
		}
		if strings.HasPrefix(dir, "/") {
			dir = dir[1:]
		}
		files = append(files, dir)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var results []string
	for _, obj := range files {
		results = append(results, obj)
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

func totpToken() string {
	t := os.Getenv("LOCKBOX_TOTP")
	if t == "" {
		t = "totp"
	}
	return t + internal.Extension
}

func display(token string, args cli.Arguments) error {
	interactive, err := internal.IsInteractive()
	if err != nil {
		return err
	}
	if args.Short {
		interactive = false
	}
	if !interactive && args.Clip {
		return errors.New("clipboard not available in non-interactive mode")
	}
	redStart, redEnd, err := internal.GetColor(internal.ColorRed)
	if err != nil {
		return err
	}
	tok := strings.TrimSpace(token)
	store := filepath.Join(internal.GetStore(), tok, totpToken())
	if !internal.PathExists(store) {
		return errors.New("object does not exist")
	}
	l, err := internal.NewLockbox(internal.LockboxOptions{File: store})
	if err != nil {
		return err
	}
	val, err := l.Decrypt()
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
	clip := internal.ClipboardCommands{}
	if args.Clip {
		clip, err = internal.NewClipboardCommands()
		if err != nil {
			internal.Die("invalid clipboard", err)
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
			startColor = redStart
			endColor = redEnd
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
			clip.CopyToClipboard(code, exe)
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
		internal.Die("subkey required", errors.New("invalid arguments"))
	}
	cmd := args[1]
	options := cli.ParseArgs(cmd)
	if options.List {
		result, err := list()
		if err != nil {
			internal.Die("invalid list response", err)
		}
		sort.Strings(result)
		for _, entry := range result {
			fmt.Println(entry)
		}
		return
	}
	if len(args) == 3 {
		if !options.Clip && !options.Short && !options.Once {
			internal.Die("subcommand not supported", errors.New("invalid sub command"))
		}
		cmd = args[2]
	}
	if err := display(cmd, options); err != nil {
		internal.Die("failed to show totp token", err)
	}
}
