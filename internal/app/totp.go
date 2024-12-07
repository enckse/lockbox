// Package app handles TOTP tokens.
package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	coreotp "github.com/pquerna/otp"
	otp "github.com/pquerna/otp/totp"

	"github.com/seanenck/lockbox/internal/backend"
	"github.com/seanenck/lockbox/internal/config"
	"github.com/seanenck/lockbox/internal/platform/clip"
	"github.com/seanenck/lockbox/internal/util"
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
		Mode  Mode
		Entry string
		token string
	}
	totpWrapper struct {
		opts otp.ValidateOpts
		code string
	}
	// TOTPOptions are TOTP call options
	TOTPOptions struct {
		app           CommandOptions
		Clear         func()
		CanTOTP       func() (bool, error)
		IsInteractive func() (bool, error)
	}
)

const (
	// UnknownTOTPMode is an unknown command
	UnknownTOTPMode Mode = iota
	// InsertTOTPMode is inserting a new totp token
	InsertTOTPMode
	// ShowTOTPMode will show the token
	ShowTOTPMode
	// ClipTOTPMode will copy to clipboard
	ClipTOTPMode
	// MinimalTOTPMode will display minimal information to display the token
	MinimalTOTPMode
	// ListTOTPMode lists the available tokens
	ListTOTPMode
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

func colorWhenRules() ([]util.TimeWindow, error) {
	envTime := config.EnvTOTPColorBetween.Get()
	if envTime == config.TOTPDefaultBetween {
		return config.TOTPDefaultColorWindow, nil
	}
	return util.ParseTimeWindow(envTime)
}

func (w totpWrapper) generateCode() (string, error) {
	return otp.GenerateCodeCustom(w.code, time.Now(), w.opts)
}

func (args *TOTPArguments) display(opts TOTPOptions) error {
	interactive, err := opts.IsInteractive()
	if err != nil {
		return err
	}
	if args.Mode == MinimalTOTPMode {
		interactive = false
	}
	once := args.Mode == OnceTOTPMode
	clipMode := args.Mode == ClipTOTPMode
	if !interactive && clipMode {
		return errors.New("clipboard not available in non-interactive mode")
	}
	entity, err := opts.app.Transaction().Get(backend.NewPath(args.Entry, args.token), backend.SecretValue)
	if err != nil {
		return err
	}
	if entity == nil {
		return errors.New("object does not exist")
	}
	totpToken := string(entity.Value)
	k, err := coreotp.NewKeyFromURL(config.EnvTOTPFormat.Get(totpToken))
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
	running := 0
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
	can, err := opts.CanTOTP()
	if err != nil {
		return err
	}
	if !can {
		return ErrNoTOTP
	}
	if args.Mode == ListTOTPMode {
		e, err := opts.app.Transaction().QueryCallback(backend.QueryOptions{Mode: backend.SuffixMode, Criteria: backend.NewSuffix(args.token)})
		if err != nil {
			return err
		}
		writer := opts.app.Writer()
		for entry, err := range e {
			if err != nil {
				return err
			}
			fmt.Fprintf(writer, "%s\n", entry.Directory())
		}
		return nil
	}
	return args.display(opts)
}

// NewTOTPArguments will parse the input arguments
func NewTOTPArguments(args []string, tokenType string) (*TOTPArguments, error) {
	if len(args) == 0 {
		return nil, errors.New("not enough arguments for totp")
	}
	if strings.TrimSpace(tokenType) == "" {
		return nil, errors.New("invalid token type, not set?")
	}
	opts := &TOTPArguments{Mode: UnknownTOTPMode}
	opts.token = tokenType
	sub := args[0]
	needs := true
	switch sub {
	case TOTPListCommand:
		needs = false
		if len(args) != 1 {
			return nil, errors.New("list takes no arguments")
		}
		opts.Mode = ListTOTPMode
	case TOTPInsertCommand:
		opts.Mode = InsertTOTPMode
	case TOTPShowCommand:
		opts.Mode = ShowTOTPMode
	case TOTPClipCommand:
		opts.Mode = ClipTOTPMode
	case TOTPMinimalCommand:
		opts.Mode = MinimalTOTPMode
	case TOTPOnceCommand:
		opts.Mode = OnceTOTPMode
	default:
		return nil, ErrUnknownTOTPMode
	}
	if needs {
		if len(args) != 2 {
			return nil, errors.New("invalid arguments")
		}
		opts.Entry = args[1]
		if opts.Mode == InsertTOTPMode {
			if !strings.HasSuffix(opts.Entry, tokenType) {
				opts.Entry = backend.NewPath(opts.Entry, tokenType)
			}
		}
	}
	return opts, nil
}
