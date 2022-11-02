// Package backend has types
package backend

import (
	"errors"

	"github.com/tobischo/gokeepasslib/v3"
)

type (
	// QueryMode indicates HOW an entity will be found
	QueryMode int
	// ValueMode indicates what to do with the store value of the entity
	ValueMode int
	// ActionMode represents activities performed via transactions
	ActionMode string
	// HookMode are hook operations the user can tie to
	HookMode string
	// QueryOptions indicates how to find entities
	QueryOptions struct {
		Mode     QueryMode
		Values   ValueMode
		Criteria string
	}
	// QueryEntity is the result of a query
	QueryEntity struct {
		Path    string
		Value   string
		backing gokeepasslib.Entry
	}
	action func(t Context) error
	// Transaction handles the overall operation of the transaction
	Transaction struct {
		valid    bool
		file     string
		exists   bool
		write    bool
		readonly bool
	}
	// Context handles operating on the underlying database
	Context struct {
		db *gokeepasslib.Database
	}
	// Hook represents a runnable user-defined hook
	Hook struct {
		path    string
		mode    ActionMode
		enabled bool
		scripts []string
	}
	removal struct {
		parts []string
		title string
		hook  Hook
	}
)

const (
	noneMode QueryMode = iota
	// ListMode indicates ALL entities will be listed
	ListMode
	// FindMode indicates a _contains_ search for an entity
	FindMode
	// ExactMode means an entity must MATCH the string exactly
	ExactMode
	// SuffixMode will look for an entity ending in a specific value
	SuffixMode
	// PrefixMode allows for entities starting with a specific value
	PrefixMode
)

const (
	// MoveAction represents changes via moves, like the Move command
	MoveAction ActionMode = "mv"
	// InsertAction represents changes via inserts, like the Insert command
	InsertAction ActionMode = "insert"
	// RemoveAction represents changes via deletions, like Remove or globbed remove commands
	RemoveAction ActionMode = "rm"
	// HookPre are triggers BEFORE an action is performed on an entity
	HookPre HookMode = "pre"
	// HookPost are triggers AFTER an action is performed on an entity
	HookPost HookMode = "post"
)

const (
	// BlankValue will not decrypt secrets, empty value
	BlankValue ValueMode = iota
	// HashedValue will decrypt and then hash the password
	HashedValue
	// SecretValue will have the raw secret onboard
	SecretValue
)

const (
	notesKey = "Notes"
	titleKey = "Title"
	passKey  = "Password"
	pathSep  = "/"
	isGlob   = pathSep + "*"
)

var (
	errPath = errors.New("input paths must contain at LEAST 2 components")
)
