// Package subcommands handles clearing the clipboard.
package subcommands

import (
	"os/exec"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
)

// ClearClipboardCallback handles clearing the clipboard over time.
func ClearClipboardCallback() error {
	idx := 0
	val, err := inputs.Stdin(false)
	if err != nil {
		return err
	}
	clipboard, err := platform.NewClipboard()
	if err != nil {
		return err
	}
	pCmd, pArgs := clipboard.Args(false)
	val = strings.TrimSpace(val)
	for idx < clipboard.MaxTime {
		idx++
		time.Sleep(1 * time.Second)
		out, err := exec.Command(pCmd, pArgs...).Output()
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(out)) != val {
			return nil
		}
	}
	return clipboard.CopyTo("")
}
