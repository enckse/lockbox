// Package subcommands handles listing items from the lockbox store.
package subcommands

import (
	"strings"

	"github.com/enckse/lockbox/internal/store"
)

type (
	// ListFindOptions for listing/finding entries in a store.
	ListFindOptions struct {
		Find   bool
		Search string
		Store  store.FileSystem
	}
)

// ListFindCallback for searching/finding/listing entries.
func ListFindCallback(args ListFindOptions) ([]string, error) {
	viewOptions := store.ViewOptions{Display: true}
	if args.Find {
		viewOptions.Filter = func(inPath string) string {
			if strings.Contains(inPath, args.Search) {
				return inPath
			}
			return ""
		}
	}
	files, err := args.Store.List(viewOptions)
	if err != nil {
		return nil, err
	}
	return files, nil
}
