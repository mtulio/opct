---
name: go-validate
description: Run the full Go validation suite (tidy, build, test, vet) before committing. Use after any Go code changes.
allowed-tools: Bash, Read
---

# Go Validate

Run the standard validation checklist before committing Go changes.

## When to use

- Before committing any Go code changes
- After dependency updates (`go.mod`, `go.sum`)
- After modifying embedded templates (`data/templates/`)
- When CI fails and you need to reproduce locally

## Steps

```bash
# 1. Resolve dependencies
go mod tidy

# 2. Build
make build

# 3. Run tests
make test

# 4. Run vet
make vet

# 5. Lint
make test-lint
```

## Expected results

- `make build`: produces `build/opct-linux-amd64`
- `make test`: all packages pass (currently ~10 packages)
- `make vet`: no output (clean)
- `make test-lint`: Go code should pass; if YAML workflow lint fails, verify each failure predates your change — do not ignore failures in modified workflow files

## Common issues

| Issue | Fix |
|-------|-----|
| `no required module provides package` | Add with an explicit version aligned to `go.mod` (e.g., `go get example.com/pkg@v1.2.3`), then `go mod tidy` |
| `toolchain` directive in go.mod | Remove it — causes CI compatibility issues |
| Import cycle | Check `internal/` vs `pkg/` boundaries |
| `go.sum` mismatch | Delete `go.sum` and run `go mod tidy` to regenerate |
