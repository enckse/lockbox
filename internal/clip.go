package internal

import (
	"fmt"
	"os/exec"
)

const (
	// MaxClipTime is the max time to let something stay in the clipboard.
	MaxClipTime = 45
)

// CopyToClipboard will copy to clipboard, if non-empty will clear later.
func CopyToClipboard(value string) {
	pipeTo("pbcopy", value, true)
	if value != "" {
		fmt.Printf("clipboard will clear in %d seconds\n", MaxClipTime)
		pipeTo("lb", value, false, "clear")
	}
}

func pipeTo(command, value string, wait bool, args ...string) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		Die("unable to get stdin pipe", err)
	}

	go func() {
		defer stdin.Close()
		if _, err := stdin.Write([]byte(value)); err != nil {
			fmt.Printf("failed writing to stdin: %v\n", err)
		}
	}()
	var ran error
	if wait {
		ran = cmd.Run()
	} else {
		ran = cmd.Start()
	}
	if ran != nil {
		Die("failed to run command", ran)
	}
}
