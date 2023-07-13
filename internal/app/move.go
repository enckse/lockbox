package app

import (
	"errors"
	"fmt"

	"github.com/enckse/lockbox/internal/backend"
)

type (
	moveRequest struct {
		cmd       CommandOptions
		src       string
		dst       string
		overwrite bool
	}
)

// Move is the CLI command to move entries
func Move(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) != 2 {
		return errors.New("src/dst required for move")
	}
	t := cmd.Transaction()
	src := args[0]
	dst := args[1]
	m, err := t.MatchPath(src)
	if err != nil {
		return err
	}
	var requests []moveRequest
	switch len(m) {
	case 1:
		requests = append(requests, moveRequest{cmd: cmd, src: m[0].Path, dst: dst, overwrite: true})
	case 0:
		break
	default:
		if !backend.IsDirectory(dst) {
			return fmt.Errorf("%s must be a path, not an entry", dst)
		}
		srcDir := backend.Directory(src)
		dir := backend.Directory(dst)
		for _, e := range m {
			srcPath := backend.Directory(e.Path)
			if srcPath != srcDir {
				return fmt.Errorf("multiple moves can only be done at a leaf level")
			}
			r := moveRequest{cmd: cmd, src: e.Path, dst: backend.NewPath(dir, backend.Base(e.Path)), overwrite: false}
			if err := r.do(true); err != nil {
				return err
			}
			requests = append(requests, r)
		}
	}
	rCount := len(requests)
	if rCount == 0 {
		return errors.New("no source entries matched")
	}
	for _, r := range requests {
		if err := r.do(false); err != nil {
			return err
		}
	}
	return nil
}

func (r moveRequest) do(dryRun bool) error {
	tx := r.cmd.Transaction()
	if !dryRun {
		use, err := backend.NewTransaction()
		if err != nil {
			return err
		}
		tx = use

	}
	srcExists, err := tx.Get(r.src, backend.SecretValue)
	if err != nil {
		return errors.New("unable to get source entry")
	}
	if srcExists == nil {
		return errors.New("no source object found")
	}
	dstExists, err := tx.Get(r.dst, backend.BlankValue)
	if err != nil {
		return errors.New("unable to get destination object")
	}
	if dstExists != nil {
		if r.overwrite {
			if !r.cmd.Confirm("overwrite destination") {
				return nil
			}
		} else {
			return errors.New("unable to overwrite entries when moving multiple items")
		}
	}
	if dryRun {
		return nil
	}
	return tx.Move(*srcExists, r.dst)
}
