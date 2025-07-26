// Package app can move entries
package app

import (
	"errors"
	"fmt"

	"github.com/enckse/lockbox/internal/kdbx"
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
		if !kdbx.IsDirectory(dst) {
			return fmt.Errorf("%s must be a path, not an entry", dst)
		}
		srcDir := kdbx.Directory(src)
		dir := kdbx.Directory(dst)
		for _, e := range m {
			srcPath := kdbx.Directory(e.Path)
			if srcPath != srcDir {
				return fmt.Errorf("multiple moves can only be done at a leaf level")
			}
			r := moveRequest{cmd: cmd, src: e.Path, dst: kdbx.NewPath(dir, kdbx.Base(e.Path)), overwrite: false}
			if _, err := r.do(true); err != nil {
				return err
			}
			requests = append(requests, r)
		}
	}
	rCount := len(requests)
	if rCount == 0 {
		return errors.New("no source entries matched")
	}
	var moving []kdbx.MoveRequest
	for _, r := range requests {
		req, err := r.do(false)
		if err != nil {
			return err
		}
		if req != nil {
			moving = append(moving, *req)
		}
	}
	return t.Move(moving...)
}

func (r moveRequest) do(dryRun bool) (*kdbx.MoveRequest, error) {
	tx := r.cmd.Transaction()
	if !dryRun {
		use, err := kdbx.NewTransaction()
		if err != nil {
			return nil, err
		}
		tx = use

	}
	srcExists, err := tx.Get(r.src, kdbx.SecretValue)
	if err != nil {
		return nil, errors.New("unable to get source entry")
	}
	if srcExists == nil {
		return nil, errors.New("no source object found")
	}
	dstExists, err := tx.Get(r.dst, kdbx.BlankValue)
	if err != nil {
		return nil, errors.New("unable to get destination object")
	}
	if dstExists != nil {
		if r.overwrite {
			if !r.cmd.Confirm("overwrite destination") {
				return nil, nil
			}
		} else {
			return nil, errors.New("unable to overwrite entries when moving multiple items")
		}
	}
	if dryRun {
		return nil, nil
	}
	return &kdbx.MoveRequest{Source: srcExists, Destination: r.dst}, nil
}
