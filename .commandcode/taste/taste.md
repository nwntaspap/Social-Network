# Taste (Continuously Learned by [CommandCode][cmd])

[cmd]: https://commandcode.ai/

# makefile

- `make gates` should only run Go build and the gates binary (`go build ./...` + `go run cmd/gates/main.go --all`), not additional CI targets like `be-ci-new` or `fe-ci` which are already included in the gates. Confidence: 0.65
