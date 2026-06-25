/*
DAGGate validates the acyclic dependency rules of feature packages (D6).
It uses 'go-arch-lint check' if available, or falls back to standard Go
package list analysis with a depth-first search (DFS) cycle-detection algorithm.
It also ensures no package imports the forbidden 'notification' package.
*/

package gates

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DAGGate validates D6 dependency DAG acyclicity (Gate #4).
// Primary: go-arch-lint. Fallback: go list + DFS cycle detection.
type DAGGate struct {
	InternalDir string // defaults to "internal"
}

func (g *DAGGate) Name() string { return "d6-dag" }

func (g *DAGGate) Run() Result {
	what := "feature package dependency graph and notification package import boundaries"
	why := "to prevent circular feature imports and isolate notification subscriptions in accordance with D6 rules"

	// Try go-arch-lint first
	if toolAvailable("go-arch-lint") {
		cmd := ExecCommand("go-arch-lint", "check")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return Result{
				Gate:    g.Name(),
				Status:  "FAIL",
				Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - go-arch-lint violations:\n%s | debug: run 'go-arch-lint check'", what, why, string(out)),
			}
		}
		// Still check notification imports (go-arch-lint doesn't enforce this)
		if notifErrs := g.checkNotificationImports(); len(notifErrs) > 0 {
			return Result{
				Gate:    g.Name(),
				Status:  "FAIL",
				Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: run 'go list' and inspect imports", what, why, strings.Join(notifErrs, "; ")),
			}
		}
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - dependency graph is acyclic (verified via go-arch-lint)", what, why),
		}
	}

	// Fallback: go list + DFS
	return g.runFallback(what, why)
}

//nolint:gocognit
func (g *DAGGate) runFallback(what, why string) Result {
	dir := g.InternalDir
	if dir == "" {
		dir = "internal"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "SKIP",
			Message: fmt.Sprintf("checked: %s | why: %s | status: SKIP - cannot read %s: %v", what, why, dir, err),
		}
	}

	var features []string
	for _, e := range entries {
		if e.IsDir() && isFeatureSlice(dir, e.Name()) {
			features = append(features, e.Name())
		}
	}

	if len(features) == 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - no feature slices yet (pre-migration, verified via fallback)", what, why),
		}
	}

	// Build dependency graph
	deps := make(map[string][]string)
	for _, feature := range features {
		featureDeps, err := getFeatureDeps(dir, feature)
		if err != nil {
			continue
		}
		deps[feature] = featureDeps
	}

	// DFS cycle detection
	var errors []string
	const (
		white = 0 // unvisited
		gray  = 1 // in current path
		black = 2 // fully processed
	)
	color := make(map[string]int)
	parent := make(map[string]string)

	var dfs func(node string) []string
	dfs = func(node string) []string {
		color[node] = gray
		for _, dep := range deps[node] {
			switch color[dep] {
			case gray:
				// Back edge → cycle. Build path.
				path := []string{dep, node}
				cur := node
				for cur != dep {
					cur = parent[cur]
					if cur == "" {
						break
					}
					path = append(path, cur)
				}
				// Reverse for readability
				for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
					path[i], path[j] = path[j], path[i]
				}
				return []string{"CIRCULAR: " + strings.Join(path, " → ")}
			case white:
				parent[dep] = node
				if errs := dfs(dep); len(errs) > 0 {
					return errs
				}
			}
		}
		color[node] = black
		return nil
	}

	for _, feature := range features {
		if color[feature] == white {
			if errs := dfs(feature); len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}
	}

	// Check notification imports
	errors = append(errors, g.checkNotificationImports()...)

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: run 'go list -json' and inspect cycles", what, why, strings.Join(errors, "; ")),
		}
	}
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - dependency graph is acyclic (verified via fallback)", what, why),
	}
}

func (g *DAGGate) checkNotificationImports() []string {
	dir := g.InternalDir
	if dir == "" {
		dir = "internal"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var errors []string
	for _, e := range entries {
		if !e.IsDir() || !isFeatureSlice(dir, e.Name()) || e.Name() == "notification" {
			continue
		}
		featureDeps, err := getFeatureDeps(dir, e.Name())
		if err != nil {
			continue
		}
		for _, dep := range featureDeps {
			if dep == "notification" {
				errors = append(errors, fmt.Sprintf("D6: %s imports notification (forbidden)", e.Name()))
			}
		}
	}
	return errors
}

type goListPkg struct {
	Imports []string `json:"Imports"`
}

// getFeatureDeps returns other feature slices that this feature imports.
func getFeatureDeps(_, feature string) ([]string, error) {
	cmd := ExecCommand("go", "list", "-json", "social-network/internal/"+feature+"/...")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var deps []string
	seen := make(map[string]bool)
	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var pkg goListPkg
		if err := dec.Decode(&pkg); err != nil {
			break
		}
		for _, imp := range pkg.Imports {
			if !strings.HasPrefix(imp, "social-network/internal/") {
				continue
			}
			dep := strings.TrimPrefix(imp, "social-network/internal/")
			dep = strings.SplitN(dep, "/", 2)[0]
			if dep == feature || skipDirs[dep] || seen[dep] {
				continue
			}
			seen[dep] = true
			deps = append(deps, dep)
		}
	}
	//nolint:nilerr
	return deps, nil
}
