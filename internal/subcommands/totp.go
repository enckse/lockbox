// Package subcommands handles TOTP tokens.
package subcommands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/colors"
	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/store"
	otp "github.com/pquerna/otp/totp"
)

type (
	colorWhen struct {
		start int
		end   int
	}
)

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Printf("unable to clear screen: %v\n", err)
	}
}

func totpEnv() string {
	return inputs.EnvOrDefault(inputs.TotpEnv, "totp")
}

func colorWhenRules() ([]colorWhen, error) {
	envTime := os.Getenv(inputs.ColorBetweenEnv)
	if envTime == "" {
		return []colorWhen{
			{start: 0, end: 5},
			{start: 30, end: 35},
		}, nil
	}
	var rules []colorWhen
	for _, item := range strings.Split(envTime, ",") {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid colorization rule found: %s", line)
		}
		s, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		e, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		if s < 0 || e < 0 || e < s || s > 59 || e > 59 {
			return nil, fmt.Errorf("invalid time found for colorization rule: %s", line)
		}
		rules = append(rules, colorWhen{start: s, end: e})
	}
	if len(rules) == 0 {
		return nil, errors.New("invalid colorization rules for totp, none found")
	}
	return rules, nil
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
	tok := filepath.Join(strings.TrimSpace(token), totpEnv())
	pathing := f.NewPath(tok)
	if !f.Exists(pathing) {
		return errors.New("object does not exist")
	}
	val, err := encrypt.FromFile(pathing)
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
	if !args.Clip {
		if !args.Once {
			clear()
		}
	}
	clipboard := platform.Clipboard{}
	if args.Clip {
		clipboard, err = platform.NewClipboard()
		if err != nil {
			return err
		}
	}
	colorRules, err := colorWhenRules()
	if err != nil {
		return err
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
		for _, when := range colorRules {
			if left < when.end && left >= when.start {
				startColor = coloring.Start
				endColor = coloring.End
			}
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
			return clipboard.CopyTo(code)
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

// TOTP handles UI for TOTP tokens.
func TOTP(args []string) error {
	if len(args) > 2 || len(args) < 1 {
		return errors.New("invalid arguments, subkey and entry required")
	}
	cmd := args[0]
	options := cli.ParseArgs(cmd)
	if options.List {
		f := store.NewFileSystemStore()
		token := f.NewFile(totpEnv())
		results, err := f.List(store.ViewOptions{ErrorOnEmpty: true, Filter: func(path string) string {
			if filepath.Base(path) == token {
				return filepath.Dir(f.CleanPath(path))
			}
			return ""
		}})
		if err != nil {
			return err
		}
		sort.Strings(results)
		for _, entry := range results {
			fmt.Println(entry)
		}
		return nil
	}
	if len(args) == 2 {
		if !options.Clip && !options.Short && !options.Once {
			return errors.New("invalid sub command")
		}
		cmd = args[1]
	}
	if err := display(cmd, options); err != nil {
		return err
	}
	return nil
}
