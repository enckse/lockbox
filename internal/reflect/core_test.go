package reflect_test

import (
	"fmt"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/reflect"
)

type mock struct {
	Name  string
	Field string
}

func TestListFields(t *testing.T) {
	fields := reflect.ListFields(mock{"abc", "xyz"})
	if len(fields) != 2 || fmt.Sprintf("%v", fields) != "[abc xyz]" {
		t.Errorf("invalid fields: %v", fields)
	}
}
