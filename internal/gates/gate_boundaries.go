/*
BoundariesGate validates D5 import boundary rules between packages.
It uses 'golangci-lint run --enable-only=depguard' to check boundaries,
with an AST-based parser fallback if golangci-lint is not installed.
*/

package gates

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// BoundariesGate validates D5 import boundary rules (Gate #3).
// Primary: golangci-lint with depguard. Fallback: AST scan.
type BoundariesGate struct {
	InternalDir string // defaults to "internal"
}

func (g *BoundariesGate) Name() string { return "d5-boundaries" }

func (g *BoundariesGate) Run() Result {
	what := "import boundaries between vertical slices and layers"
	why := "to prevent modularity leaks and ensure domain logic remains decoupled from transport and store implementations"

	// Try golangci-lint first (depguard rules in .golangci.yml)
	if toolAvailable("golangci-lint") {
		cmd := ExecCommand("golangci-lint", "run", "--enable-only=depguard", "--timeout=5m")
		out, err := cmd.CombinedOutput()
		if err != nil {
			// Non-zero exit = violations found
			return Result{
				Gate:    g.Name(),
				Status:  "FAIL",
				Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - depguard violations:\n%s | debug: run 'golangci-lint run --enable-only=depguard'", what, why, string(out)),
			}
		}
		return Result{
			Gate:    g.Name(),
			Status:  "PASS",
			Message: fmt.Sprintf("checked: %s | why: %s | status: OK - no forbidden imports found (verified via golangci-lint depguard)", what, why),
		}
	}

	// Fallback: AST-based check
	return g.runAST(what, why)
}

func (g *BoundariesGate) runAST(what, why string) Result {
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

	// Get working directory for path trimming
	wd, _ := os.Getwd()

	var errors []string

	for _, e := range entries {
		if !e.IsDir() || !isFeatureSlice(dir, e.Name()) {
			continue
		}
		featureDir := filepath.Join(dir, e.Name())

		// Rule: transport/ must not import store/
		errors = append(errors, checkImports(featureDir, "transport", []string{"/store"}, wd)...)

		// Rule: store/ must not import transport/, commands/, queries/
		errors = append(errors, checkImports(featureDir, "store", []string{"/transport", "/commands", "/queries"}, wd)...)

		// Rule: commands/ must not import store/ or transport/
		errors = append(errors, checkImports(featureDir, "commands", []string{"/store", "/transport"}, wd)...)

		// Rule: queries/ must not import store/ or transport/
		errors = append(errors, checkImports(featureDir, "queries", []string{"/store", "/transport"}, wd)...)

		// Rule: feature root files (e.g. internal/user/user.go) must not import own transport/ or store/
		errors = append(errors, checkRootImports(featureDir, []string{"/transport", "/store"}, wd)...)
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: review AST boundary import violations", what, why, strings.Join(errors, "; ")),
		}
	}
	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - no forbidden imports found (verified via AST fallback)", what, why),
	}
}

// checkRootImports parses Go files directly in the featureDir (not subdirectories)
// and checks for forbidden import path fragments.
func checkRootImports(featureDir string, forbidden []string, wd string) []string {
	var violations []string
	fset := token.NewFileSet()

	entries, err := os.ReadDir(featureDir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(featureDir, e.Name())
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			continue
		}

		f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			continue
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if !strings.Contains(importPath, "internal/") {
				continue
			}
			for _, frag := range forbidden {
				if strings.Contains(importPath, frag) {
					pos := fset.Position(imp.Pos())
					relPath := pos.Filename
					if wd != "" {
						relPath = strings.TrimPrefix(relPath, wd+"/")
					}
					violations = append(violations, fmt.Sprintf("D5: %s:%d imports %s", relPath, pos.Line, importPath))
				}
			}
		}
	}
	return violations
}

// checkImports parses Go files in a subdirectory and checks for forbidden import path fragments.
func checkImports(featureDir, subDir string, forbidden []string, wd string) []string {
	dir := filepath.Join(featureDir, subDir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	var violations []string
	fset := token.NewFileSet()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return nil //nolint:nilerr // skip unparseable files
		}

		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if !strings.Contains(importPath, "internal/") {
				continue
			}
			for _, frag := range forbidden {
				if strings.Contains(importPath, frag) {
					pos := fset.Position(imp.Pos())
					relPath := pos.Filename
					if wd != "" {
						relPath = strings.TrimPrefix(relPath, wd+"/")
					}
					violations = append(violations, fmt.Sprintf("D5: %s:%d imports %s", relPath, pos.Line, importPath))
				}
			}
		}
		return nil
	})
	if err != nil {
		violations = append(violations, fmt.Sprintf("walk error in %s: %v", dir, err))
	}

	return violations
}
