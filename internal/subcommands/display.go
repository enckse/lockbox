// Package subcommands handles displaying various lockbox structures to the UI.
package subcommands

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/enckse/lockbox/internal/colors"
	"github.com/enckse/lockbox/internal/dump"
	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/store"
)

type (
	// DisplayOptions for getting a set of items for display uses.
	DisplayOptions struct {
		Dump  bool
		Entry string
		Show  bool
		Glob  string
		All   bool
		Store store.FileSystem
	}
)

// DisplayCallback handles getting entries for display.
func DisplayCallback(args DisplayOptions) ([]dump.ExportEntity, error) {
	entries := []string{args.Entry}
	if strings.Contains(args.Entry, "*") || args.All {
		if args.Entry == args.Glob || args.All {
			all, err := args.Store.List(store.ViewOptions{})
			if err != nil {
				return nil, err
			}
			entries = all
		} else {
			matches, err := filepath.Glob(args.Entry)
			if err != nil {
				return nil, err
			}
			entries = matches
		}
	}
	isGlob := len(entries) > 1
	if isGlob {
		if !args.Show {
			return nil, errors.New("bad glob request")
		}
		sort.Strings(entries)
	}
	coloring, err := colors.NewTerminal(colors.Red)
	if err != nil {
		return nil, err
	}
	dumpData := []dump.ExportEntity{}
	for _, entry := range entries {
		if !store.PathExists(entry) {
			return nil, errors.New("entry not found")
		}
		decrypt, err := encrypt.FromFile(entry)
		if err != nil {
			return nil, err
		}
		entity := dump.ExportEntity{Value: strings.TrimSpace(string(decrypt))}
		if args.Show && isGlob {
			fileName := args.Store.CleanPath(entry)
			if args.Dump {
				entity.Path = fileName
			} else {
				entity.Path = fmt.Sprintf("%s%s:%s", coloring.Start, fileName, coloring.End)
			}
		}
		dumpData = append(dumpData, entity)
	}
	return dumpData, nil
}
