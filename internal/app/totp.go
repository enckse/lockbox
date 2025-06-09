// Package app handles TOTP tokens.
package app

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	coreotp "github.com/pquerna/otp"
	otp "github.com/pquerna/otp/totp"

	"git.sr.ht/~enckse/lockbox/internal/app/commands"
	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/kdbx"
	"git.sr.ht/~enckse/lockbox/internal/platform/clip"
)

var (
	// ErrNoTOTP is used when TOTP is requested BUT is disabled
	ErrNoTOTP = errors.New("totp is disabled")
	// ErrUnknownTOTPMode indicates an unknown totp argument type
	ErrUnknownTOTPMode = errors.New("unknown totp mode")
)

type (
	// TOTPArguments are the parsed TOTP call arguments
	TOTPArguments struct {
		Entry string
		Mode  string
	}
	totpWrapper struct {
		code string
		opts otp.ValidateOpts
	}
	// TOTPOptions are TOTP call options
	TOTPOptions struct {
		app   CommandOptions
		Clear func()
	}
)

// NewDefaultTOTPOptions gets the default option set
func NewDefaultTOTPOptions(app CommandOptions) TOTPOptions {
	return TOTPOptions{
		app:   app,
		Clear: clearFunc,
	}
}

func clearFunc() {
	fmt.Print("\033[H\033[2J")
}

func colorWhenRules() ([]config.TimeWindow, error) {
	envTime := config.EnvTOTPColorBetween.Get()
	if slices.Compare(envTime, config.TOTPDefaultBetween) == 0 {
		return config.TOTPDefaultColorWindow, nil
	}
	return ParseTimeWindow(envTime...)
}

func (w totpWrapper) generateCode() (string, error) {
	return otp.GenerateCodeCustom(w.code, time.Now(), w.opts)
}

func (args *TOTPArguments) display(opts TOTPOptions) error {
	interactive := !slices.Contains([]string{commands.TOTPMinimal, commands.TOTPSeed, commands.TOTPURL}, args.Mode)
	once := args.Mode == commands.TOTPOnce
	clipMode := args.Mode == commands.TOTPClip
	if !interactive && clipMode {
		return errors.New("clipboard not available in non-interactive mode")
	}
	if !kdbx.IsLeafAttribute(args.Entry, kdbx.OTPField) {
		return fmt.Errorf("'%s' is not a TOTP entry", args.Entry)
	}
	entity, err := getEntity(args.Entry, opts.app)
	if err != nil {
		return err
	}
	k, err := coreotp.NewKeyFromURL(config.EnvTOTPFormat.Get(entity))
	if err != nil {
		return err
	}
	wrapper := totpWrapper{}
	wrapper.code = k.Secret()
	wrapper.opts = otp.ValidateOpts{}
	wrapper.opts.Digits = k.Digits()
	wrapper.opts.Algorithm = k.Algorithm()
	wrapper.opts.Period = uint(k.Period())
	writer := opts.app.Writer()
	switch args.Mode {
	case commands.TOTPSeed:
		fmt.Fprintln(writer, wrapper.code)
		return nil
	case commands.TOTPURL:
		fmt.Fprintf(writer, "url:       %s\n", k.URL())
		fmt.Fprintf(writer, "seed:      %s\n", wrapper.code)
		fmt.Fprintf(writer, "digits:    %s\n", wrapper.opts.Digits)
		fmt.Fprintf(writer, "algorithm: %s\n", wrapper.opts.Algorithm)
		fmt.Fprintf(writer, "period:    %d\n", wrapper.opts.Period)
		return nil
	}
	if !interactive {
		code, err := wrapper.generateCode()
		if err != nil {
			return err
		}
		fmt.Fprintf(writer, "%s\n", code)
		return nil
	}
	first := true
	var running int64
	lastSecond := -1
	if !clipMode {
		if !once {
			opts.Clear()
		}
	}
	clipboard := clip.Board{}
	if clipMode {
		clipboard, err = clip.New()
		if err != nil {
			return err
		}
	}
	colorRules, err := colorWhenRules()
	if err != nil {
		return err
	}
	runFor, err := config.EnvTOTPTimeout.Get()
	if err != nil {
		return err
	}
	allowColor := config.CanColor()
	for {
		if !first {
			time.Sleep(500 * time.Millisecond)
		}
		first = false
		running++
		if running > runFor {
			fmt.Fprint(writer, "exiting (timeout)\n")
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
		isColor := false
		if allowColor {
			for _, when := range colorRules {
				if left < when.End && left >= when.Start {
					isColor = true
				}
			}
		}
		leftString := fmt.Sprintf("%d", left)
		if len(leftString) < 2 {
			leftString = "0" + leftString
		}
		txt := fmt.Sprintf("%s (%s)", now.Format("15:04:05"), leftString)
		if isColor {
			txt = fmt.Sprintf("\x1b[31m%s\x1b[39m", txt)
		}
		outputs := []string{txt}
		if !clipMode {
			outputs = append(outputs, fmt.Sprintf("%s\n    %s", args.Entry, code))
			if !once {
				outputs = append(outputs, "-> CTRL+C to exit")
			}
		} else {
			fmt.Fprintf(writer, "-> %s\n", txt)
			return clipboard.CopyTo(code)
		}
		if !once {
			opts.Clear()
		}
		fmt.Fprintf(writer, "%s\n", strings.Join(outputs, "\n\n"))
		if once {
			return nil
		}
	}
}

// Do will perform the TOTP operation
func (args *TOTPArguments) Do(opts TOTPOptions) error {
	if args.Mode == "" {
		return ErrUnknownTOTPMode
	}
	if opts.Clear == nil {
		return errors.New("invalid option functions")
	}
	if args.Mode == commands.TOTPList {
		return doList(kdbx.OTPField, args.Entry, opts.app, false)
	}
	return args.display(opts)
}

// NewTOTPArguments will parse the input arguments
func NewTOTPArguments(args []string) (*TOTPArguments, error) {
	if len(args) == 0 {
		return nil, errors.New("not enough arguments for totp")
	}
	opts := &TOTPArguments{}
	sub := args[0]
	needs := true
	length := len(args)
	switch sub {
	case commands.TOTPList:
		needs = false
		if length != 1 {
			needs = true
			if length != 2 {
				return nil, errors.New("list takes only a filter (if any)")
			}
		}
	case commands.TOTPURL:
	case commands.TOTPSeed:
	case commands.TOTPShow:
	case commands.TOTPClip:
	case commands.TOTPMinimal:
	case commands.TOTPOnce:
	default:
		return nil, ErrUnknownTOTPMode
	}
	opts.Mode = sub
	if needs {
		if length != 2 {
			return nil, errors.New("invalid arguments")
		}
		opts.Entry = args[1]
	}
	return opts, nil
}

// ParseTimeWindow will handle parsing a window of colors for TOTP operations
func ParseTimeWindow(windows ...string) ([]config.TimeWindow, error) {
	var rules []config.TimeWindow
	for _, item := range windows {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		parts := strings.Split(line, config.TimeWindowSpan)
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
		rules = append(rules, config.TimeWindow{Start: s, End: e})
	}
	if len(rules) == 0 {
		return nil, errors.New("invalid colorization rules for totp, none found")
	}
	return rules, nil
}
