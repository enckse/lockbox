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
	"git.sr.ht/~enckse/lockbox/internal/backend"
	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/platform/clip"
)

var (
	// ErrNoTOTP is used when TOTP is requested BUT is disabled
	ErrNoTOTP = errors.New("totp is disabled")
	// ErrUnknownTOTPMode indicates an unknown totp argument type
	ErrUnknownTOTPMode = errors.New("unknown totp mode")
)

type (
	// Mode is the operating mode for TOTP operations
	Mode int
	// TOTPArguments are the parsed TOTP call arguments
	TOTPArguments struct {
		Entry string
		Mode  Mode
	}
	totpWrapper struct {
		code string
		opts otp.ValidateOpts
	}
	// TOTPOptions are TOTP call options
	TOTPOptions struct {
		app           CommandOptions
		Clear         func()
		CanTOTP       func() bool
		IsInteractive func() bool
	}
)

const (
	// UnknownTOTPMode is an unknown command
	UnknownTOTPMode Mode = iota
	// ShowTOTPMode will show the token
	ShowTOTPMode
	// ClipTOTPMode will copy to clipboard
	ClipTOTPMode
	// MinimalTOTPMode will display minimal information to display the token
	MinimalTOTPMode
	// ListTOTPMode lists the available tokens
	ListTOTPMode
	// FindTOTPMode is list but with a regexp filter
	FindTOTPMode
	// OnceTOTPMode will only show the token once and exit
	OnceTOTPMode
)

// NewDefaultTOTPOptions gets the default option set
func NewDefaultTOTPOptions(app CommandOptions) TOTPOptions {
	return TOTPOptions{
		app:           app,
		Clear:         clearFunc,
		IsInteractive: config.EnvInteractive.Get,
		CanTOTP:       config.EnvTOTPEnabled.Get,
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
	interactive := opts.IsInteractive()
	if args.Mode == MinimalTOTPMode {
		interactive = false
	}
	once := args.Mode == OnceTOTPMode
	clipMode := args.Mode == ClipTOTPMode
	if !interactive && clipMode {
		return errors.New("clipboard not available in non-interactive mode")
	}
	if !backend.IsLeafAttribute(args.Entry, backend.OTP) {
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
	allowColor, err := config.CanColor()
	if err != nil {
		return err
	}
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
	if args.Mode == UnknownTOTPMode {
		return ErrUnknownTOTPMode
	}
	if opts.Clear == nil || opts.CanTOTP == nil || opts.IsInteractive == nil {
		return errors.New("invalid option functions")
	}
	if !opts.CanTOTP() {
		return ErrNoTOTP
	}
	if args.Mode == ListTOTPMode || args.Mode == FindTOTPMode {
		return doList(backend.OTP, args.Entry, opts.app, false)
	}
	return args.display(opts)
}

// NewTOTPArguments will parse the input arguments
func NewTOTPArguments(args []string) (*TOTPArguments, error) {
	if len(args) == 0 {
		return nil, errors.New("not enough arguments for totp")
	}
	opts := &TOTPArguments{Mode: UnknownTOTPMode}
	sub := args[0]
	needs := true
	switch sub {
	case commands.TOTPList:
		needs = false
		if len(args) != 1 {
			return nil, errors.New("list takes no arguments")
		}
		opts.Mode = ListTOTPMode
	case commands.TOTPFind:
		opts.Mode = FindTOTPMode
	case commands.TOTPShow:
		opts.Mode = ShowTOTPMode
	case commands.TOTPClip:
		opts.Mode = ClipTOTPMode
	case commands.TOTPMinimal:
		opts.Mode = MinimalTOTPMode
	case commands.TOTPOnce:
		opts.Mode = OnceTOTPMode
	default:
		return nil, ErrUnknownTOTPMode
	}
	if needs {
		if len(args) != 2 {
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
