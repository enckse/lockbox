// Package totp handles TOTP tokens.
package totp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/colors"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	coreotp "github.com/pquerna/otp"
	otp "github.com/pquerna/otp/totp"
)

const (
	// ClipCommand is the argument for copying totp codes to clipboard
	ClipCommand = "-clip"
	// ShortCommand is the argument for getting the short version of a code
	ShortCommand = "-short"
	// ListCommand will list the totp-enabled entries
	ListCommand = "-list"
	// OnceCommand will perform like a normal totp request but not refresh
	OnceCommand = "-once"
)

type (
	colorWhen struct {
		start int
		end   int
	}
	arguments struct {
		Clip  bool
		Once  bool
		Short bool
		List  bool
	}
	totpWrapper struct {
		opts otp.ValidateOpts
		code string
	}
)

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Printf("unable to clear screen: %v\n", err)
	}
}

func colorWhenRules() ([]colorWhen, error) {
	envTime := inputs.EnvOrDefault(inputs.ColorBetweenEnv, inputs.TOTPDefaultBetween)
	if envTime == "" || envTime == inputs.TOTPDefaultBetween {
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

func (w totpWrapper) generateCode() (string, error) {
	return otp.GenerateCodeCustom(w.code, time.Now(), w.opts)
}

func display(token string, args arguments) error {
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
	t, err := backend.NewTransaction()
	if err != nil {
		return err
	}
	entity, err := t.Get(backend.NewPath(token, inputs.TOTPToken()), backend.SecretValue)
	if err != nil {
		return err
	}
	if entity == nil {
		return errors.New("object does not exist")
	}
	totpToken := string(entity.Value)
	k, err := coreotp.NewKeyFromURL(totpToken)
	if err != nil {
		return err
	}
	wrapper := totpWrapper{}
	wrapper.code = k.Secret()
	wrapper.opts = otp.ValidateOpts{}
	wrapper.opts.Digits = k.Digits()
	wrapper.opts.Algorithm = k.Algorithm()
	wrapper.opts.Period = uint(k.Period())
	if !interactive {
		code, err := wrapper.generateCode()
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
		code, err := wrapper.generateCode()
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
			outputs = append(outputs, fmt.Sprintf("%s\n    %s", token, code))
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

// Call handles UI for TOTP tokens.
func Call(args []string) error {
	if len(args) > 2 || len(args) < 1 {
		return errors.New("invalid arguments, subkey and entry required")
	}
	cmd := args[0]
	options := parseArgs(cmd)
	if options.List {
		t, err := backend.NewTransaction()
		if err != nil {
			return err
		}
		e, err := t.QueryCallback(backend.QueryOptions{Mode: backend.SuffixMode, Criteria: backend.NewSuffix(inputs.TOTPToken())})
		if err != nil {
			return err
		}
		for _, entry := range e {
			fmt.Println(entry.Directory())
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

func parseArgs(arg string) arguments {
	args := arguments{}
	args.Clip = arg == ClipCommand
	args.Once = arg == OnceCommand
	args.Short = arg == ShortCommand
	args.List = arg == ListCommand
	return args
}
