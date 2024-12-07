package core_test

import (
	"fmt"
	"testing"

	"github.com/seanenck/lockbox/internal/core"
)

type mock struct {
	Name  string
	Field string
}

func TestListFields(t *testing.T) {
	fields := core.ListFields(mock{"abc", "xyz"})
	if len(fields) != 2 || fmt.Sprintf("%v", fields) != "[abc xyz]" {
		t.Errorf("invalid fields: %v", fields)
	}
}
