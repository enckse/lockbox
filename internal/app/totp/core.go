// Package totp handles TOTP operations
package totp

import (
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/ja7ad/otp"

	"github.com/enckse/lockbox/internal/config"
)

type (
	// Generator is used to generate TOTP codes
	Generator struct {
		secret string
		opts   *otp.Param
		url    *url.URL
	}
)

// Code will generate a new code for the specified TOTP object
func (g Generator) Code() (string, error) {
	return otp.GenerateTOTP(g.secret, time.Now(), g.opts)
}

// Print will print information about the generator to the writer
func (g Generator) Print(w io.Writer, details bool) {
	if details {
		fmt.Fprintf(w, "url:       %s\n", g.url)
		fmt.Fprintf(w, "seed:      %s\n", g.secret)
		fmt.Fprintf(w, "digits:    %d\n", g.opts.Digits)
		fmt.Fprintf(w, "algorithm: %s\n", g.opts.Algorithm)
		fmt.Fprintf(w, "period:    %d\n", g.opts.Period)
		return
	}
	fmt.Fprintln(w, g.secret)
}

// New will create a new generator
func New(input string) (Generator, error) {
	u, err := url.Parse(config.EnvTOTPFormat.Get(input))
	if err != nil {
		return Generator{}, err
	}

	obj, err := otp.ParseOTPAuthURL(u)
	if err != nil {
		return Generator{}, err
	}
	wrapper := Generator{}
	wrapper.secret = obj.Secret
	wrapper.opts = &otp.Param{}
	wrapper.opts.Algorithm = obj.Algorithm
	wrapper.opts.Digits = obj.Digits
	wrapper.opts.Period = obj.Period
	wrapper.url = u
	return wrapper, nil
}
