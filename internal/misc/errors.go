// Package misc handles error logging/handling for UI outputs.
package misc

import (
	"fmt"
	"os"
)

// Die will print messages and exit.
func Die(message string, err error) {
	msg := message
	if err != nil {
		msg = fmt.Sprintf("%s (%v)", msg, err)
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
