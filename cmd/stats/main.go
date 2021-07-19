package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"voidedtech.com/lockbox/internal"
)

type (
	// History is commit information for the object.
	History struct {
		Hash string `json:"hash"`
		Date string `json:"date"`
	}

	// Stats are general entry stats
	Stats struct {
		Entry   string    `json:"entry"`
		Name    string    `json:"name"`
		Dir     string    `json:"dir"`
		History []History `json:"history"`
	}
)

func main() {
	args := os.Args
	filtering := len(args) > 1
	filter := ""
	if filtering {
		filter = args[1]
	}
	store := internal.GetStore()
	items, err := internal.Find(store, true)
	if err != nil {
		internal.Die("unable to find entries", err)
	}
	results := []Stats{}
	for _, item := range items {
		if filtering {
			if !strings.HasPrefix(item, filter) {
				continue
			}
		}
		stat := Stats{}
		stat.Entry = item
		stat.Name = filepath.Base(item)
		stat.Dir = filepath.Dir(item)
		cmd := exec.Command("git", "-C", store, "log", "--format=%h %aI", fmt.Sprintf("%s%s", item, internal.Extension))
		b, err := cmd.Output()
		if err != nil {
			internal.Die("failed to get git history", err)
		}
		history := []History{}
		for _, value := range strings.Split(string(b), "\n") {
			cleaned := strings.TrimSpace(value)
			if len(cleaned) == 0 {
				continue
			}
			parts := strings.Split(cleaned, " ")
			if len(parts) != 2 {
				internal.Die("invalid format entry", fmt.Errorf("mismatch between format string and struct?"))
			}
			history = append(history, History{Hash: parts[0], Date: parts[1]})
		}
		stat.History = history
		results = append(results, stat)
	}
	if len(results) == 0 {
		internal.Die("found no entries", fmt.Errorf("no entries"))
	}
	j, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		internal.Die("unable to prep json", err)
	}
	fmt.Println(string(j))
}
