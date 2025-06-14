// Package features controls build flag features
package features

import "fmt"

var (
	clipFeature = true
	totpFeature = true
)

// CanTOTP indicates if TOTP is enabled
func CanTOTP() bool {
	return totpFeature
}

// CanClip indicates if clip(board) is enabled
func CanClip() bool {
	return clipFeature
}

// NewError creates an error if a feature is not enabled
func NewError(name string) error {
	return fmt.Errorf("%s feature is disabled", name)
}
