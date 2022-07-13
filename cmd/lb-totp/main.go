package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal"
	otp "github.com/pquerna/otp/totp"
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
		return nil, internal.NewLockboxError("no objects found")
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

func display(token string, clip, once, short bool) error {
	interactive, err := internal.IsInteractive()
	if err != nil {
		return err
	}
	if short {
		interactive = false
	}
	if !interactive && clip {
		return internal.NewLockboxError("clipboard not available in non-interactive mode")
	}
	redStart, redEnd, err := internal.GetColor(internal.ColorRed)
	if err != nil {
		return err
	}
	tok := strings.TrimSpace(token)
	store := filepath.Join(internal.GetStore(), tok, totpToken())
	if !internal.PathExists(store) {
		return internal.NewLockboxError("object does not exist")
	}
	l, err := internal.NewLockbox("", "", store)
	if err != nil {
		return err
	}
	val, err := l.Decrypt()
	if err != nil {
		return err
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
	if !clip {
		if !once {
			clear()
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
		if !clip {
			outputs = append(outputs, fmt.Sprintf("%s\n    %s", tok, code))
			if !once {
				outputs = append(outputs, "-> CTRL+C to exit")
			}
		} else {
			fmt.Printf("-> %s\n", expires)
			internal.CopyToClipboard(code)
			return nil
		}
		if !once {
			clear()
		}
		fmt.Printf("%s\n", strings.Join(outputs, "\n\n"))
		if once {
			return nil
		}
	}
}

func main() {
	args := os.Args
	if len(args) > 3 || len(args) < 2 {
		internal.Die("subkey required", internal.NewLockboxError("invalid arguments"))
	}
	cmd := args[1]
	if cmd == "-list" || cmd == "-ls" {
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
	clip := false
	once := false
	short := false
	if len(args) == 3 {
		if cmd != "-c" && cmd != "clip" && cmd != "-once" && cmd != "-short" {
			internal.Die("subcommand not supported", internal.NewLockboxError("invalid sub command"))
		}
		clip = cmd == "-clip" || cmd == "-c"
		once = cmd == "-once"
		short = cmd == "-short"
		cmd = args[2]
	}
	if err := display(cmd, clip, once, short); err != nil {
		internal.Die("failed to show totp token", err)
	}
}
