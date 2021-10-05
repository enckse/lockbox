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
	"voidedtech.com/stock"
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
	redStart, redEnd, err := internal.GetColor(internal.ColorRed)
	if err != nil {
		return err
	}
	tok := strings.TrimSpace(token)
	store := filepath.Join(getEnv(), tok+internal.Extension)
	if !stock.PathExists(store) {
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
		startColor := ""
		endColor := ""
		if left < 10 {
			startColor = redStart
			endColor = redEnd
		}
		if !clip {
			outputs = append(outputs, fmt.Sprintf("%s\n    %s", tok, code))
			outputs = append(outputs, "-> CTRL+C to exit")
		} else {
			colorize(startColor, fmt.Sprintf("\n  -> %s\n", expires), endColor)
			internal.CopyToClipboard(code)
			return nil
		}
		clear()
		colorize(startColor, strings.Join(outputs, "\n\n"), endColor)
	}
}

func colorize(start, text, end string) {
	fmt.Printf("%s%s%s\n", start, text, end)
}

func main() {
	args := os.Args
	if len(args) > 3 || len(args) < 2 {
		stock.Die("subkey required", internal.NewLockboxError("invalid arguments"))
	}
	cmd := args[1]
	if cmd == "list" || cmd == "ls" {
		result, err := list()
		if err != nil {
			stock.Die("invalid list response", err)
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
			stock.Die("subcommand not supported", internal.NewLockboxError("invalid sub command"))
		}
		clip = true
		cmd = args[2]
	}
	if err := display(cmd, clip); err != nil {
		stock.Die("failed to show totp token", err)
	}
}
