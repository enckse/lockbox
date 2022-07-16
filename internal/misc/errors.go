package misc

import (
	"fmt"
	"os"
)

// LogError will log an error to stderr.
func LogError(message string, err error) {
	msg := message
	if err != nil {
		msg = fmt.Sprintf("%s (%v)", msg, err)
	}
	fmt.Fprintln(os.Stderr, msg)
}

// Die will print messages and exit.
func Die(message string, err error) {
	LogError(message, err)
	os.Exit(1)
}
