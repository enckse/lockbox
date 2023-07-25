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

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/system"
	coreotp "github.com/pquerna/otp"
	otp "github.com/pquerna/otp/totp"
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
	// Arguments are the parsed TOTP call arguments
	Arguments struct {
		Mode  Mode
		Entry string
		token string
	}
	totpWrapper struct {
		opts otp.ValidateOpts
		code string
	}
	// Options are TOTP call options
	Options struct {
		app           app.CommandOptions
		Clear         func()
		IsNoTOTP      func() (bool, error)
		IsInteractive func() (bool, error)
	}
)

const (
	// UnknownMode is an unknown command
	UnknownMode Mode = iota
	// InsertMode is inserting a new totp token
	InsertMode
	// ShowMode will show the token
	ShowMode
	// ClipMode will copy to clipboard
	ClipMode
	// MinimalMode will display minimal information to display the token
	MinimalMode
	// ListMode lists the available tokens
	ListMode
	// OnceMode will only show the token once and exit
	OnceMode
)

// NewDefaultOptions gets the default option set
func NewDefaultOptions(app app.CommandOptions) Options {
	return Options{
		app:           app,
		Clear:         clear,
		IsInteractive: inputs.IsInteractive,
		IsNoTOTP:      inputs.IsNoTOTP,
	}
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Printf("unable to clear screen: %v\n", err)
	}
}

func colorWhenRules() ([]inputs.ColorWindow, error) {
	envTime := system.EnvironOrDefault(inputs.ColorBetweenEnv, inputs.TOTPDefaultBetween)
	if envTime == inputs.TOTPDefaultBetween {
		return inputs.TOTPDefaultColorWindow, nil
	}
	return inputs.ParseColorWindow(envTime)
}

func (w totpWrapper) generateCode() (string, error) {
	return otp.GenerateCodeCustom(w.code, time.Now(), w.opts)
}

func (args *Arguments) display(opts Options) error {
	interactive, err := opts.IsInteractive()
	if err != nil {
		return err
	}
	if args.Mode == MinimalMode {
		interactive = false
	}
	once := args.Mode == OnceMode
	clip := args.Mode == ClipMode
	if !interactive && clip {
		return errors.New("clipboard not available in non-interactive mode")
	}
	coloring, err := NewTerminal(Red)
	if err != nil {
		return err
	}
	entity, err := opts.app.Transaction().Get(backend.NewPath(args.Entry, args.token), backend.SecretValue)
	if err != nil {
		return err
	}
	if entity == nil {
		return errors.New("object does not exist")
	}
	totpToken := string(entity.Value)
	k, err := coreotp.NewKeyFromURL(inputs.FormatTOTP(totpToken))
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
	if !clip {
		if !once {
			opts.Clear()
		}
	}
	clipboard := platform.Clipboard{}
	if clip {
		clipboard, err = platform.NewClipboard()
		if err != nil {
			return err
		}
	}
	colorRules, err := colorWhenRules()
	if err != nil {
		return err
	}
	runString := system.EnvironOrDefault(inputs.MaxTOTPTime, inputs.MaxTOTPTimeDefault)
	runFor, err := strconv.Atoi(runString)
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
		startColor := ""
		endColor := ""
		for _, when := range colorRules {
			if left < when.End && left >= when.Start {
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
		if !clip {
			outputs = append(outputs, fmt.Sprintf("%s\n    %s", args.Entry, code))
			if !once {
				outputs = append(outputs, "-> CTRL+C to exit")
			}
		} else {
			fmt.Fprintf(writer, "-> %s\n", expires)
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
func (args *Arguments) Do(opts Options) error {
	if args.Mode == UnknownMode {
		return ErrUnknownTOTPMode
	}
	if opts.Clear == nil || opts.IsNoTOTP == nil || opts.IsInteractive == nil {
		return errors.New("invalid option functions")
	}
	off, err := opts.IsNoTOTP()
	if err != nil {
		return err
	}
	if off {
		return ErrNoTOTP
	}
	if args.Mode == ListMode {
		e, err := opts.app.Transaction().QueryCallback(backend.QueryOptions{Mode: backend.SuffixMode, Criteria: backend.NewSuffix(args.token)})
		if err != nil {
			return err
		}
		writer := opts.app.Writer()
		for _, entry := range e {
			fmt.Fprintf(writer, "%s\n", entry.Directory())
		}
		return nil
	}
	return args.display(opts)
}

// NewArguments will parse the input arguments
func NewArguments(args []string, tokenType string) (*Arguments, error) {
	if len(args) == 0 {
		return nil, errors.New("not enough arguments for totp")
	}
	if strings.TrimSpace(tokenType) == "" {
		return nil, errors.New("invalid token type, not set?")
	}
	opts := &Arguments{Mode: UnknownMode}
	opts.token = tokenType
	sub := args[0]
	needs := true
	switch sub {
	case app.TOTPListCommand:
		needs = false
		if len(args) != 1 {
			return nil, errors.New("list takes no arguments")
		}
		opts.Mode = ListMode
	case app.TOTPInsertCommand:
		opts.Mode = InsertMode
	case app.TOTPShowCommand:
		opts.Mode = ShowMode
	case app.TOTPClipCommand:
		opts.Mode = ClipMode
	case app.TOTPMinimalCommand:
		opts.Mode = MinimalMode
	case app.TOTPOnceCommand:
		opts.Mode = OnceMode
	default:
		return nil, ErrUnknownTOTPMode
	}
	if needs {
		if len(args) != 2 {
			return nil, errors.New("invalid arguments")
		}
		opts.Entry = args[1]
		if opts.Mode == InsertMode {
			if !strings.HasSuffix(opts.Entry, tokenType) {
				opts.Entry = backend.NewPath(opts.Entry, tokenType)
			}
		}
	}
	return opts, nil
}
