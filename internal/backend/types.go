// Package backend has types
package backend

import (
	"github.com/tobischo/gokeepasslib/v3"
)

type (
	// QueryMode indicates HOW an entity will be found
	QueryMode int
	// ValueMode indicates what to do with the store value of the entity
	ValueMode int
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
		valid  bool
		file   string
		exists bool
		write  bool
	}
	// Context handles operating on the underlying database
	Context struct {
		db *gokeepasslib.Database
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
)
