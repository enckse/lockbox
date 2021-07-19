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
		return nil, fmt.Errorf("no objects found")
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

func display(token string) error {
	tok := strings.TrimSpace(token)
	store := filepath.Join(getEnv(), tok+internal.Extension)
	if !internal.PathExists(store) {
		return fmt.Errorf("object does not exist")
	}
	l, err := internal.NewLockbox("", "", store)
	if err != nil {
		return err
	}
	val, err := l.Decrypt()
	if err != nil {
		return err
	}
	clear()
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
		outputs = append(outputs, fmt.Sprintf("%s\n    %s", tok, code))
		outputs = append(outputs, "-> CTRL+C to exit")
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
	if len(args) != 2 {
		internal.Die("subkey required", fmt.Errorf("invalid arguments"))
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
	if err := display(cmd); err != nil {
		internal.Die("failed to show totp token", err)
	}
}
