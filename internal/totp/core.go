// Package totp handles TOTP tokens.
package totp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/colors"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	coreotp "github.com/pquerna/otp"
	otp "github.com/pquerna/otp/totp"
)

// ErrNoTOTP is used when TOTP is requested BUT is disabled
var ErrNoTOTP = errors.New("TOTP is disabled")

type (
	// Mode is the operating mode for TOTP operations
	Mode int
	// Arguments are the parsed TOTP call arguments
	Arguments struct {
		Mode  Mode
		Entry string
	}
	totpWrapper struct {
		opts otp.ValidateOpts
		code string
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
	// ShortMode will display minimal information to display the token
	ShortMode
	// ListMode lists the available tokens
	ListMode
	// OnceMode will only show the token once and exit
	OnceMode
)

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Printf("unable to clear screen: %v\n", err)
	}
}

func colorWhenRules() ([]inputs.ColorWindow, error) {
	envTime := inputs.EnvOrDefault(inputs.ColorBetweenEnv, inputs.TOTPDefaultBetween)
	if envTime == inputs.TOTPDefaultBetween {
		return inputs.TOTPDefaultColorWindow, nil
	}
	return inputs.ParseColorWindow(envTime)
}

func (w totpWrapper) generateCode() (string, error) {
	return otp.GenerateCodeCustom(w.code, time.Now(), w.opts)
}

func (args *Arguments) display(tx *backend.Transaction) error {
	interactive, err := inputs.IsInteractive()
	if err != nil {
		return err
	}
	if args.Mode == ShortMode {
		interactive = false
	}
	once := args.Mode == OnceMode
	clip := args.Mode == ClipMode
	if !interactive && clip {
		return errors.New("clipboard not available in non-interactive mode")
	}
	coloring, err := colors.NewTerminal(colors.Red)
	if err != nil {
		return err
	}
	entity, err := tx.Get(backend.NewPath(args.Entry, inputs.TOTPToken()), backend.SecretValue)
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
	if !clip {
		if !once {
			clear()
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
			fmt.Printf("-> %s\n", expires)
			return clipboard.CopyTo(code)
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

// Do will perform the TOTP operation
func (args *Arguments) Do(tx *backend.Transaction) error {
	if args == nil || args.Mode == UnknownMode {
		return errors.New("unknown totp mode")
	}
	off, err := inputs.IsNoTOTP()
	if err != nil {
		return err
	}
	if off {
		return ErrNoTOTP
	}
	if args.Mode == ListMode {
		e, err := tx.QueryCallback(backend.QueryOptions{Mode: backend.SuffixMode, Criteria: backend.NewSuffix(inputs.TOTPToken())})
		if err != nil {
			return err
		}
		for _, entry := range e {
			fmt.Println(entry.Directory())
		}
		return nil
	}
	return args.display(tx)
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
	sub := args[0]
	needs := true
	switch sub {
	case cli.TOTPListCommand:
		needs = false
		if len(args) != 1 {
			return nil, errors.New("list takes no arguments")
		}
		opts.Mode = ListMode
	case cli.TOTPInsertCommand:
		opts.Mode = InsertMode
	case cli.TOTPShowCommand:
		opts.Mode = ShowMode
	case cli.TOTPClipCommand:
		opts.Mode = ClipMode
	case cli.TOTPShortCommand:
		opts.Mode = ShortMode
	case cli.TOTPOnceCommand:
		opts.Mode = OnceMode
	default:
		return nil, errors.New("unknown totp command")
	}
	if needs {
		if len(args) != 2 {
			return nil, errors.New("missing entry")
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
