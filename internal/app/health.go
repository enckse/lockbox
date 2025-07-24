// Package app can insert
package app

import (
	"errors"
	"fmt"
	"io"

	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/platform"
)

func report(w io.Writer, cat string, err error) {
	msg := "ok"
	if err != nil {
		msg = fmt.Sprintf("error: %v", err)
	}
	rawReport(w, cat, msg)
}

func rawReport(w io.Writer, cat, msg string) {
	text := fmt.Sprintf("%-15s %s\n", cat, msg)
	w.Write([]byte(text))
}

// Health will display configuration/system health
func Health(cmd CommandOptions) error {
	w := cmd.Writer()
	rawReport(w, "item", "status")
	rawReport(w, "---", "---")
	key, err := config.NewKey(config.DefaultKeyMode)
	if err == nil {
		_, err = key.Read()
	}
	report(w, "key", err)
	err = nil
	file := config.EnvKeyFile.Get()
	if file != "" {
		err = errors.New("key file set, does not exist")

		if platform.PathExists(file) {
			err = nil
		}
	}
	report(w, "keyfile", err)
	_, err = platform.NewClipboard(platform.DefaultClipboardLoader{})
	report(w, "clipboard", err)
	store := config.EnvStore.Get()
	err = errors.New("store not set")
	if store != "" {
		err = errors.New("store does not exist")
		if platform.PathExists(store) {
			err = nil
		}
	}
	report(w, "store", err)
	return nil
}
