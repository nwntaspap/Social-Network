/*
StackGate validates the Go compiler version, module configuration,
SQLite database parameters, frontend scaffold (Next.js/Tailwind/Bun),
and platform service directory structure (Gate #1).
*/

package gates

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StackGate validates the entire development and runtime stack (Gate #1).
type StackGate struct {
	GoModPath   string // path to go.mod, defaults to "<RootDir>/go.mod"
	EnvPath     string // path to .env, defaults to "<RootDir>/.env"
	FrontendDir string // path to frontend dir, defaults to "frontend-next" or "frontend"
	RootDir     string // root directory of the repository, defaults to GitRepoRoot()
}

func (g *StackGate) Name() string { return "stack" }

type pkgJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func (g *StackGate) Run() Result {
	rootDir := g.RootDir
	if rootDir == "" {
		rootDir = GitRepoRoot()
		if rootDir == "" {
			rootDir = "."
		}
	}

	goModPath := g.GoModPath
	if goModPath == "" {
		goModPath = filepath.Join(rootDir, "go.mod")
	}

	envPath := g.EnvPath
	if envPath == "" {
		envPath = filepath.Join(rootDir, ".env")
	}

	what := "Go version, module path, frontend framework (Next.js/Bun), SQLite WAL and busy timeout configuration, and platform interface readiness"
	why := "to ensure compile compatibility, correct database lock settings, and Next.js/Bun/platform readiness"

	var errors []string
	var checkedDetails []string

	// 1. Go version and module check
	goVer, modPath, err := g.checkGoMod(goModPath)
	if err != nil {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %v | debug: run 'cat %s'", what, why, err, goModPath),
		}
	}
	checkedDetails = append(checkedDetails, "Go "+goVer)
	if modPath != "social-network" {
		errors = append(errors, "expected module 'social-network' in go.mod, got '"+modPath+"'")
	}

	// 2. Database validation in .env
	dbDetail, err := g.checkDatabase(envPath)
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		checkedDetails = append(checkedDetails, dbDetail)
	}

	// 3. Frontend validation (Next.js & Bun)
	feDetail, err := g.checkFrontend(rootDir)
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		checkedDetails = append(checkedDetails, feDetail)
	}

	// 4. Platform Services
	platformDetail, err := g.checkPlatform(rootDir)
	if err != nil {
		errors = append(errors, err.Error())
	} else {
		checkedDetails = append(checkedDetails, platformDetail)
	}

	if len(errors) > 0 {
		return Result{
			Gate:    g.Name(),
			Status:  "FAIL",
			Message: fmt.Sprintf("checked: %s | why: %s | status: FAIL - %s | debug: run 'cat %s' or check '%s' settings", what, why, strings.Join(errors, "; "), goModPath, envPath),
		}
	}

	return Result{
		Gate:    g.Name(),
		Status:  "PASS",
		Message: fmt.Sprintf("checked: %s | why: %s | status: OK - %s", what, why, strings.Join(checkedDetails, ", ")),
	}
}

func (g *StackGate) checkGoMod(goModPath string) (string, string, error) {
	// #nosec G304
	goModFile, err := os.Open(goModPath)
	if err != nil {
		return "", "", fmt.Errorf("cannot open go.mod: %w", err)
	}
	defer goModFile.Close()

	var goVersion, modulePath string
	scanner := bufio.NewScanner(goModFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if val, ok := strings.CutPrefix(line, "go "); ok {
			goVersion = val
		}
		if val, ok := strings.CutPrefix(line, "module "); ok {
			modulePath = val
		}
	}

	if !strings.HasPrefix(goVersion, "1.25") {
		return goVersion, modulePath, fmt.Errorf("expected Go 1.25 in go.mod, got %s", goVersion)
	}
	return goVersion, modulePath, nil
}

func (g *StackGate) checkDatabase(envPath string) (string, error) {
	envVars, err := parseEnvFile(envPath)
	if err != nil {
		return "", fmt.Errorf("cannot open environment config: %w", err)
	}

	dbDriver := envVars["DB_DRIVER"]
	if dbDriver == "" {
		dbDriver = envVars["DATABASE_DRIVER"]
	}
	if dbDriver == "" {
		dbDriver = "sqlite3"
	}

	switch dbDriver {
	case "sqlite3", "sqlite":
		dbPragma := envVars["DB_PRAGMA"]
		if !strings.Contains(dbPragma, "_journal_mode=WAL") {
			return "", errors.New("SQLite journal mode is not set to WAL (missing _journal_mode=WAL in DB_PRAGMA inside .env)")
		}
		if !strings.Contains(dbPragma, "_busy_timeout=5000") {
			return "", errors.New("SQLite busy timeout is not configured (missing _busy_timeout=5000 in DB_PRAGMA inside .env)")
		}
		return "SQLite (WAL & busy_timeout=5000)", nil
	case "postgres", "postgresql":
		return "PostgreSQL (portable driver)", nil
	default:
		return "", fmt.Errorf("unsupported DB_DRIVER: %s", dbDriver)
	}
}

func (g *StackGate) checkFrontend(rootDir string) (string, error) {
	var frontendDir string
	if g.FrontendDir != "" {
		frontendDir = filepath.Join(rootDir, g.FrontendDir)
	} else {
		if _, err := os.Stat(filepath.Join(rootDir, "frontend-next")); err == nil {
			frontendDir = filepath.Join(rootDir, "frontend-next")
		} else {
			frontendDir = filepath.Join(rootDir, "frontend")
		}
	}

	pkgJSONPath := filepath.Join(frontendDir, "package.json")
	if _, err := os.Stat(pkgJSONPath); os.IsNotExist(err) {
		return "frontend is not scaffolded yet", nil
	}

	// #nosec G304
	pkgData, err := os.ReadFile(pkgJSONPath)
	if err != nil {
		return "", fmt.Errorf("cannot read frontend package.json: %w", err)
	}

	var pkg pkgJSON
	if err := json.Unmarshal(pkgData, &pkg); err != nil {
		return "", fmt.Errorf("cannot parse frontend package.json: %w", err)
	}

	var feErrors []string
	if !hasDependency(pkg, "next") {
		feErrors = append(feErrors, "Next.js dependency missing in frontend package.json")
	}
	if !hasDependency(pkg, "tailwindcss") {
		feErrors = append(feErrors, "Tailwind CSS dependency missing in frontend package.json")
	}

	// Check Bun lockfile
	hasBunLock := false
	if _, err := os.Stat(filepath.Join(frontendDir, "bun.lockb")); err == nil {
		hasBunLock = true
	} else if _, err := os.Stat(filepath.Join(frontendDir, "bun.lock")); err == nil {
		hasBunLock = true
	}
	if !hasBunLock {
		feErrors = append(feErrors, "Bun lockfile (bun.lockb or bun.lock) not found in frontend directory")
	}

	if len(feErrors) > 0 {
		return "", fmt.Errorf("%s", strings.Join(feErrors, "; "))
	}

	return "Next.js/Tailwind/Bun verified", nil
}

func (g *StackGate) checkPlatform(rootDir string) (string, error) {
	platformDir := filepath.Join(rootDir, "internal", "platform")
	if _, err := os.Stat(platformDir); os.IsNotExist(err) {
		return "", errors.New("missing internal/platform directory")
	}
	return "platform interface ready", nil
}

func parseEnvFile(path string) (map[string]string, error) {
	// #nosec G304
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return env, scanner.Err()
}

func hasDependency(pkg pkgJSON, dep string) bool {
	if pkg.Dependencies != nil {
		if _, ok := pkg.Dependencies[dep]; ok {
			return true
		}
	}
	if pkg.DevDependencies != nil {
		if _, ok := pkg.DevDependencies[dep]; ok {
			return true
		}
	}
	return false
}
