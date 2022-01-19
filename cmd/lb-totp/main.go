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
		return nil, stock.NewBasicError("no objects found")
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

func display(token string, clip, once, short bool) error {
	interactive, err := internal.IsInteractive()
	if err != nil {
		return err
	}
	if short {
		interactive = false
	}
	if !interactive && clip {
		return stock.NewBasicError("clipboard not available in non-interactive mode")
	}
	redStart, redEnd, err := internal.GetColor(internal.ColorRed)
	if err != nil {
		return err
	}
	tok := strings.TrimSpace(token)
	store := filepath.Join(getEnv(), tok+internal.Extension)
	if !stock.PathExists(store) {
		return stock.NewBasicError("object does not exist")
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
		if len(leftString) < 1 {
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
		stock.Die("subkey required", stock.NewBasicError("invalid arguments"))
	}
	cmd := args[1]
	if cmd == "-list" || cmd == "-ls" {
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
	once := false
	short := false
	if len(args) == 3 {
		if cmd != "-c" && cmd != "clip" && cmd != "-once" && cmd != "-short" {
			stock.Die("subcommand not supported", stock.NewBasicError("invalid sub command"))
		}
		clip = cmd == "-clip" || cmd == "-c"
		once = cmd == "-once"
		short = cmd == "-short"
		cmd = args[2]
	}
	if err := display(cmd, clip, once, short); err != nil {
		stock.Die("failed to show totp token", err)
	}
}
