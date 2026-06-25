package gates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecurityGate_BcryptLowCost(t *testing.T) {
	dir := t.TempDir()
	code := `package auth

import "golang.org/x/crypto/bcrypt"

func hash(pw []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(pw, 10)
}
`
	if err := os.WriteFile(filepath.Join(dir, "auth.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SecurityGate{InternalDir: dir}
	errs := g.runASTChecks()
	if len(errs) == 0 {
		t.Error("expected bcrypt cost violation, got none")
	}
}

func TestSecurityGate_BcryptOKCost(t *testing.T) {
	dir := t.TempDir()
	code := `package auth

import "golang.org/x/crypto/bcrypt"

func hash(pw []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(pw, 14)
}
`
	if err := os.WriteFile(filepath.Join(dir, "auth.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SecurityGate{InternalDir: dir}
	errs := g.runASTChecks()
	if len(errs) != 0 {
		t.Errorf("expected no violations for cost 14, got: %v", errs)
	}
}

func TestSecurityGate_SQLConcat(t *testing.T) {
	dir := t.TempDir()
	code := `package repo

import "fmt"

func query(id string) string {
	return fmt.Sprintf("SELECT id FROM users WHERE id = %s", id)
}
`
	if err := os.WriteFile(filepath.Join(dir, "repo.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SecurityGate{InternalDir: dir}
	errs := g.runASTChecks()
	if len(errs) == 0 {
		t.Error("expected SQL injection warning, got none")
	}
}

func TestSecurityGate_Run(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand

	g := &SecurityGate{}

	// 1. Tool available, PASS
	lookPath = func(name string) (string, error) { return name, nil }
	t.Setenv("MOCK_FAIL", "0")
	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected security tool PASS, got: %s (%s)", res.Status, res.Message)
	}

	// 2. Tool available, FAIL
	t.Setenv("MOCK_FAIL", "1")
	res = g.Run()
	if res.Status != "FAIL" {
		t.Errorf("expected security tool FAIL, got: %s", res.Status)
	}
}

func TestSecurityGate_CheckOrigin(t *testing.T) {
	dir := t.TempDir()
	code := `package ws
import "github.com/gorilla/websocket"
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
`
	if err := os.WriteFile(filepath.Join(dir, "ws.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SecurityGate{InternalDir: dir}
	errs := g.runASTChecks()
	if len(errs) == 0 {
		t.Error("expected WebSocket CheckOrigin violation, got none")
	}
	if !strings.Contains(errs[0], "WebSocket CheckOrigin returns true unconditionally") {
		t.Errorf("expected message to mention CheckOrigin, got: %s", errs[0])
	}
}

func TestSecurityGate_CheckOriginAssignmentAndDecl(t *testing.T) {
	dir := t.TempDir()
	code := `package ws
func init() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
}
func myCheckOrigin(r *http.Request) bool {
	return true
}
`
	if err := os.WriteFile(filepath.Join(dir, "ws.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SecurityGate{InternalDir: dir}
	errs := g.runASTChecks()
	if len(errs) < 2 {
		t.Errorf("expected at least 2 WebSocket CheckOrigin violations (assignment + func decl), got %d: %v", len(errs), errs)
	}
}

func TestSecurityGate_BcryptCostResolution(t *testing.T) {
	// Test constants and variables cost resolving
	dir := t.TempDir()
	code := `package auth
import "golang.org/x/crypto/bcrypt"
const GoodCost = 13
const BadCost = 10
var LowCost = 9
const DefaultCostVal = bcrypt.DefaultCost
const MinCostVal = bcrypt.MinCost

func hash1(pw []byte) {
	bcrypt.GenerateFromPassword(pw, GoodCost)
}
func hash2(pw []byte) {
	bcrypt.GenerateFromPassword(pw, BadCost)
}
func hash3(pw []byte) {
	bcrypt.GenerateFromPassword(pw, LowCost)
}
func hash4(pw []byte) {
	bcrypt.GenerateFromPassword(pw, DefaultCostVal)
}
func hash5(pw []byte) {
	bcrypt.GenerateFromPassword(pw, MinCostVal)
}
func hash6(pw []byte) {
	bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
}
`
	if err := os.WriteFile(filepath.Join(dir, "auth.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SecurityGate{InternalDir: dir}
	errs := g.runASTChecks()
	// Should flag: hash2 (BadCost=10), hash3 (LowCost=9), hash4 (DefaultCostVal=bcrypt.DefaultCost=10), hash5 (MinCostVal=bcrypt.MinCost=4), hash6 (bcrypt.DefaultCost=10)
	if len(errs) != 5 {
		t.Errorf("expected 5 bcrypt violations, got %d: %v", len(errs), errs)
	}
}

func TestSecurityGate_EdgeCases(t *testing.T) {
	dir := t.TempDir()
	code := `package edge
import "os"
import "golang.org/x/crypto/bcrypt"
import "fmt"
type Dummy struct{}
const CostString = "ten"
const CostFloat = 12.34
const CostChain = GoodCost
const GoodCost = 14

func getCost() int { return 12 }

func test() {
	// unresolved cost
	_ = bcrypt.GenerateFromPassword(nil, getCost())
	
	// float cost
	_ = bcrypt.GenerateFromPassword(nil, CostFloat)
	
	// string cost
	_ = bcrypt.GenerateFromPassword(nil, CostString)
	
	// chain cost
	_ = bcrypt.GenerateFromPassword(nil, CostChain)
	
	// assignment not func lit
	CheckOrigin = nil
	
	// assignment to other field
	otherField = func() {}

	// isCheckOrigin default case: dereference assignment
	*CheckOrigin = nil

	// checkSQLConcat: not fmt Sprintf
	log.Printf("SELECT")

	// checkSQLConcat: Sprintf with no args
	fmt.Sprintf()

	// checkSQLConcat: Sprintf with non-basic-lit first arg
	fmt.Sprintf(CostString)
}
`
	if err := os.WriteFile(filepath.Join(dir, "edge.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}

	g := &SecurityGate{InternalDir: dir}
	_ = g.runASTChecks()
}

func TestSecurityGate_HasThirdPartyVulns(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"no vulnerabilities", "", false},
		{"standard library only", "Vulnerability #1: Standard library", false},
		{"third party found", "Vulnerability #1: some-lib@v1.0.0", true},
		{"multiple with stdlib", "Vulnerability #1: Standard library\nVulnerability #2: third-party-lib", true},
		{"third party found (different format)", "Vulnerability #1: golang.org/x/crypto", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasThirdPartyVulns(tt.output)
			if got != tt.want {
				t.Errorf("hasThirdPartyVulns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecurityGate_RunProdPath(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand
	lookPath = func(name string) (string, error) { return "", os.ErrNotExist }

	g := &SecurityGate{}
	t.Setenv("MOCK_FAIL", "0")

	res := g.Run()
	if res.Status != "PASS" && res.Status != "FAIL" {
		t.Errorf("expected PASS or FAIL, got: %s (%s)", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "AST only") {
		t.Errorf("expected 'AST only' suffix in message, got: %s", res.Message)
	}
}

func TestSecurityGate_RunProdPath_WithTools(t *testing.T) {
	oldExec := ExecCommand
	oldLook := lookPath
	defer func() {
		ExecCommand = oldExec
		lookPath = oldLook
	}()
	ExecCommand = mockExecCommand
	lookPath = func(name string) (string, error) { return name, nil }

	g := &SecurityGate{}
	t.Setenv("MOCK_FAIL", "0")

	res := g.Run()
	if res.Status != "PASS" {
		t.Errorf("expected PASS with tools, got: %s (%s)", res.Status, res.Message)
	}
	if !strings.Contains(res.Message, "gosec + govulncheck + AST") {
		t.Errorf("expected all tools in message, got: %s", res.Message)
	}
}
