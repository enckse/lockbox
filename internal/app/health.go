// Package app can insert
package app

import (
	"errors"
	"fmt"
	"io"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/platform"
)

func report(w io.Writer, cat string, err error) {
	msg := "ok"
	if err != nil {
		msg = fmt.Sprintf("error: %v", err)
	}
	text := fmt.Sprintf("%s\n  -> %s\n", cat, msg)
	w.Write([]byte(text))
}

// Health will display configuration/system health
func Health(cmd CommandOptions) error {
	key, err := config.NewKey(config.DefaultKeyMode)
	w := cmd.Writer()
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
