# Issue: Rename Module from github.com/arnald/forum to social-network

## What to build

Rename the Go module name in `go.mod` from `github.com/arnald/forum` to `social-network` and update all project imports inside `.go` files accordingly to align with the social network project's domain specification.

## Acceptance Criteria

- [ ] `go.mod` contains the module definition `module social-network`.
- [ ] All `.go` source files under `cmd/`, `internal/`, and `db/` have their imports updated from `"github.com/arnald/forum/...` to `"social-network/...`.
- [ ] Running `go build ./cmd/server` and `go build ./cmd/client` compiles successfully.
- [ ] Running `go mod tidy` and `go vet ./...` succeeds without any errors.

## Execution Guide

To perform this refactoring automatically on a Linux/WSL environment:

```bash
# 1. Update module declaration in go.mod
sed -i 's|module github.com/arnald/forum|module social-network|g' go.mod

# 2. Update imports in all Go source files recursively
find . -type f -name "*.go" -exec sed -i 's|github.com/arnald/forum|social-network|g' {} +

# 3. Clean up go.mod and go.sum dependencies
go mod tidy

# 4. Verify formatting and compile safety
go vet ./...
go build ./cmd/server && go build ./cmd/client
```
