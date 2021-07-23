package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	otp "github.com/pquerna/otp/totp"
	"voidedtech.com/lockbox/internal"
)

func getEnv() string {
	return filepath.Join(internal.GetStore(), os.Getenv("LOCKBOX_TOTP"))
}

func list() ([]string, error) {
	path := getEnv()
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, obj := range files {
		f := obj.Name()
		if strings.HasSuffix(f, internal.Extension) {
			results = append(results, strings.TrimSuffix(f, internal.Extension))
		}
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

func display(token string, clip bool) error {
	tok := strings.TrimSpace(token)
	store := filepath.Join(getEnv(), tok+internal.Extension)
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
	if !clip {
		clear()
	}
	totpToken := string(val)
	first := true
	running := 0
	lastSecond := -1
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
		expires := fmt.Sprintf("%s, expires: %2d (seconds)", now.Format("15:04:05"), left)
		outputs := []string{expires}
		code, err := otp.GenerateCode(totpToken, now)
		if err != nil {
			return err
		}
		if !clip {
			outputs = append(outputs, fmt.Sprintf("%s\n    %s", tok, code))
			outputs = append(outputs, "-> CTRL+C to exit")
		} else {
			fmt.Printf("\n  -> %s\n\n", expires)
			internal.CopyToClipboard(code)
			return nil
		}
		startColor := ""
		endColor := ""
		if left < 10 {
			startColor = "\033[1;31m"
			endColor = "\033[0m"
		}
		clear()
		fmt.Printf("%s%s%s\n", startColor, strings.Join(outputs, "\n\n"), endColor)
	}
}

func main() {
	args := os.Args
	if len(args) > 3 {
		internal.Die("subkey required", internal.NewLockboxError("invalid arguments"))
	}
	cmd := args[1]
	if cmd == "list" || cmd == "ls" {
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
	if len(args) == 3 {
		if cmd != "-c" && cmd != "clip" {
			internal.Die("subcommand not supported", internal.NewLockboxError("invalid sub command"))
		}
		clip = true
		cmd = args[2]
	}
	if err := display(cmd, clip); err != nil {
		internal.Die("failed to show totp token", err)
	}
}
