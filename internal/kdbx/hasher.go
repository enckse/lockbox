package kdbx

import (
	"crypto/sha512"
	"fmt"
	"slices"
	"strings"

	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/output"
)

type (
	checksummable struct {
		value  string
		typeof string
		index  int
	}
	// Hasher is the manager of data output transform via hashing
	Hasher struct {
		isRaw          bool
		isHashed       bool
		isChecksum     bool
		checksums      []checksummable
		checksumTo     int
		requiredLength int
	}
)

// Reset will clear any current hashing data
func (h *Hasher) Reset() {
	h.checksums = nil
}

// NewHasher will create a new hasher to manage data outputs
func NewHasher(mode ValueMode) (*Hasher, error) {
	jsonMode := output.JSONModes.Blank
	if mode == JSONValue {
		m, err := output.ParseJSONMode(config.EnvJSONMode.Get())
		if err != nil {
			return nil, err
		}
		jsonMode = m
	}

	obj := &Hasher{}
	obj.checksumTo = 1
	obj.isRaw = mode == SecretValue
	switch jsonMode {
	case output.JSONModes.Raw:
		obj.isRaw = true
	case output.JSONModes.Hash:
		obj.isHashed = true
		obj.isChecksum = mode == JSONValue
		length, err := config.EnvJSONHashLength.Get()
		if err != nil {
			return nil, err
		}
		obj.checksumTo = max(int(length), obj.checksumTo)
	}
	obj.requiredLength = (len(AllowedFields) + 2) * obj.checksumTo
	return obj, nil
}

// Transform will transform a value to the preferred output type
func (h *Hasher) Transform(value string) string {
	if h.isChecksum {
		return value
	}
	return h.compute(value)
}

func (h *Hasher) compute(value string) string {
	if h.isRaw {
		return value
	}
	if h.isHashed {
		return fmt.Sprintf("%x", sha512.Sum512([]byte(value)))
	}
	return ""
}

// Add will add a value for checksum computation if needed
func (h *Hasher) Add(field, value string) bool {
	if h.isChecksum {
		if len(field) > 0 && len(value) > 0 {
			if r := h.compute(value); len(r) > 0 {
				typeof := field[0]
				h.checksums = append(h.checksums, checksummable{r[0:h.checksumTo], string(typeof), int(typeof)})
			}
		}
		return true
	}
	return false
}

// Calculate will generate the output checksum
func (h *Hasher) Calculate(key string) (string, bool) {
	if !h.isChecksum {
		return "", false
	}
	var check string
	if len(h.checksums) > 0 {
		h.Add("d", key)
		slices.SortFunc(h.checksums, func(x, y checksummable) int {
			return x.index - y.index
		})
		var vals []string
		for _, item := range h.checksums {
			vals = append(vals, fmt.Sprintf("%s%s", item.value, string(item.typeof)))
		}
		for len(vals) < h.requiredLength {
			vals = append([]string{"0" + strings.Repeat("0", h.checksumTo)}, vals...)
		}
		check = fmt.Sprintf("[%s]", strings.Join(vals, " "))
	}
	return check, true
}
