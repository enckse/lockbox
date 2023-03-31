// Package inputs handles user inputs/UI elements.
package inputs

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/enckse/pgl/os/env"
)

const (
	otpAuth   = "otpauth"
	otpIssuer = "lbissuer"
)

// FormatTOTP will format a totp otpauth url
func FormatTOTP(value string) string {
	if strings.HasPrefix(value, otpAuth) {
		return value
	}
	override := env.GetOrDefault(formatTOTPEnv, "")
	if override != "" {
		return fmt.Sprintf(override, value)
	}
	v := url.Values{}
	v.Set("secret", value)
	v.Set("issuer", otpIssuer)
	v.Set("period", "30")
	v.Set("algorithm", "SHA1")
	v.Set("digits", "6")
	u := url.URL{
		Scheme:   otpAuth,
		Host:     "totp",
		Path:     "/" + otpIssuer + ":" + "lbaccount",
		RawQuery: v.Encode(),
	}
	return u.String()
}
