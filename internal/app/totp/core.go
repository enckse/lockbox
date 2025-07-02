// Package totp handles TOTP operations
package totp

import (
	"fmt"
	"io"
	"time"

	coreotp "github.com/pquerna/otp"
	otp "github.com/pquerna/otp/totp"

	"git.sr.ht/~enckse/lockbox/internal/config"
)

type (
	// Generator is used to generate TOTP codes
	Generator struct {
		key    *coreotp.Key
		secret string
		opts   otp.ValidateOpts
	}
)

// Code will generate a new code for the specified TOTP object
func (g Generator) Code() (string, error) {
	return otp.GenerateCodeCustom(g.secret, time.Now(), g.opts)
}

// Print will print information about the generator to the writer
func (g Generator) Print(w io.Writer, details bool) {
	if details {
		fmt.Fprintf(w, "url:       %s\n", g.key.URL())
		fmt.Fprintf(w, "seed:      %s\n", g.secret)
		fmt.Fprintf(w, "digits:    %s\n", g.opts.Digits)
		fmt.Fprintf(w, "algorithm: %s\n", g.opts.Algorithm)
		fmt.Fprintf(w, "period:    %d\n", g.opts.Period)
		return
	}
	fmt.Fprintln(w, g.secret)
}

// New will create a new generator
func New(code string) (Generator, error) {
	k, err := coreotp.NewKeyFromURL(config.EnvTOTPFormat.Get(code))
	if err != nil {
		return Generator{}, err
	}
	wrapper := Generator{}
	wrapper.secret = k.Secret()
	wrapper.opts = otp.ValidateOpts{}
	wrapper.opts.Digits = k.Digits()
	wrapper.opts.Algorithm = k.Algorithm()
	wrapper.opts.Period = uint(k.Period())
	wrapper.key = k
	return wrapper, nil
}
