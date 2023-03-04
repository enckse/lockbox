// Package app common objects
package app

type (
	// Confirm user inputs
	Confirm func(string) bool
	// InsertOptions are functions required for insert
	InsertOptions struct {
		IsPipe  func() bool
		Input   func(bool, bool) ([]byte, error)
		Confirm Confirm
	}
	// InsertArgs are parsed insert settings for insert commands
	InsertArgs struct {
		Entry string
		Multi bool
		Opts  InsertOptions
	}
)
