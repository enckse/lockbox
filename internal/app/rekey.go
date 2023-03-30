package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/inputs"
)

func getCommandLines(exe string, args ...string) ([]string, error) {
	out, err := exec.Command(exe, args...).Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

// ReKey handles entry rekeying
func ReKey(cmd *DefaultCommand) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	env, err := inputs.GetReKey()
	if err != nil {
		return err
	}
	entries, err := getCommandLines(exe, cli.ListCommand)
	if err != nil {
		return err
	}
	writer := cmd.Writer()
	for _, entry := range entries {
		if _, err := fmt.Fprintf(writer, "rekeying: %s\n", entry); err != nil {
			return err
		}
		stats, err := getCommandLines(exe, cli.StatsCommand, entry)
		if err != nil {
			return fmt.Errorf("failed to get modtime, command failed: %w", err)
		}
		modTime := ""
		for _, stat := range stats {
			if strings.HasPrefix(stat, backend.ModTimeField) {
				if modTime != "" {
					return errors.New("unable to read modtime, too many values")
				}
				modTime = strings.TrimPrefix(stat, backend.ModTimeField)
			}
		}
		modTime = strings.TrimSpace(modTime)
		if modTime == "" {
			return errors.New("did not read modtime")
		}
		data, err := exec.Command(exe, cli.ShowCommand, entry).Output()
		if err != nil {
			return err
		}
		cmd := exec.Command(exe, cli.InsertCommand, entry)
		cmd.Env = append(os.Environ(), env...)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", inputs.ModTimeEnv, modTime))
		in, err := cmd.StdinPipe()
		if nil != err {
			return err
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		go func() {
			defer in.Close()
			in.Write(data)
		}()
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
