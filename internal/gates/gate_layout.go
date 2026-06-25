/*
LayoutGate validates the physical structure of vertical slice feature packages (D1).
It ensures each feature folder under the internal directory has the required
structure: a main feature file (<feature>.go) and subdirectories for commands,
queries, transport, and store.
*/

package gates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// skipDirs are directories under internal/ that are not feature slices.
var skipDirs = map[string]bool{
	"core": true, "platform": true, "pkg": true, "config": true,
	"bootstrap": true, "domain": true, "infra": true, "app": true,
	"gates": true,
}

// LayoutGate validates D1 vertical slice layout (Gate #2).
type LayoutGate struct {
	InternalDir string // defaults to "internal"
}

func (g *LayoutGate) Name() string { return "d1-layout" }

func (g *LayoutGate) Run() Result {
	dir := g.InternalDir
	if dir == "" {
		dir = "internal"
	}

	what := "presence of active feature packages and core vertical slice directories (commands, queries, transport, store)"
	why := "to enforce the physical vertical slice layout pattern (CQRS + transport/store separation)"

	entries, err := os.ReadDir(dir)
	if err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "SKIP",
			Message: fmt.Sprintf("checked: %s | why: %s | status: SKIP - cannot read %s: %v", what, why, dir, err),
		}
	}

	var errors []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !isFeatureSlice(dir, name) {
			continue
		}

		featureDir := filepath.Join(dir, name)
		required := []string{
			filepath.Join(featureDir, name+".go"),
		}
		requiredDirs := []string{"commands", "queries", "transport", "store"}

		for _, r := range required {
			if _, err := os.Stat(r); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("%s: missing %s", featureDir, filepath.Base(r)))
			}
		}
		for _, d := range requiredDirs {
			if _, err := os.Stat(filepath.Join(featureDir, d)); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("%s: missing %s/", featureDir, d))
			}
		}
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: run 'tree %s' or verify vertical slice directory structure", what, why, strings.Join(errors, "; "), dir),
		}
	}
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - all active features contain expected vertical slice subfolders", what, why),
	}
}
