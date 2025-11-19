# CLAUDE.md - Development Instructions for AI Assistants

This document provides structured instructions for common development activities on the OPCT project, designed for AI assistants (like Claude) to execute consistently and correctly.

> **Note**: This document is intended for non-release development tasks. For release procedures, refer to [docs/devel/release.md](docs/devel/release.md) and [docs/devel/update.md](docs/devel/update.md).

## Table of Contents

- [Project Structure](#project-structure)
- [Development Tasks](#development-tasks)
  - [Go Version Bump](#go-version-bump)
  - [Dependency Management](#dependency-management)
- [Validation Procedures](#validation-procedures)
- [Contributing Guidelines](#contributing-guidelines)

---

## Project Structure

OPCT is organized into the following key components:

- **CLI (`cmd/opct`)**: Client-side utility to orchestrate conformance workflows
- **Internal packages (`internal/`)**: Core business logic and utilities
- **Public packages (`pkg/`)**: Public APIs and client interfaces
- **Plugins**: Container-based workflow steps (separate repository)
- **Documentation (`docs/`)**: User and developer documentation

### Key Directories

```
.
├── .github/workflows/      # GitHub Actions CI/CD workflows
├── cmd/opct/              # CLI entrypoint
├── internal/              # Internal packages
├── pkg/                   # Public packages
├── hack/                  # Build scripts and Containerfile
├── docs/                  # Documentation
│   ├── devel/            # Developer guides
│   └── user/             # User guides
└── test/                  # Test utilities
```

---

## Development Tasks

### Go Version Bump

**When to use**: Update project to use a newer Go version available in the build environment.

**Command pattern**: "Update to Go 1.X.Y" or "Bump Go version to 1.X.Y"

#### Files to Update

1. **`go.mod`**
   - Update `go` directive (e.g., `go 1.25.0`)
   - Update `toolchain` directive to match environment (e.g., `toolchain go1.25.4`)

2. **`.github/workflows/*.yaml` (ALL workflow files)**
   - Search for ALL files with `GO_VERSION`: `grep -rn "GO_VERSION" .github/workflows/`
   - Update `GO_VERSION` environment variable in:
     - `go.yaml` (line ~16)
     - `pre_linters.yaml` (line ~19)
     - `pre_reviewer.yaml` (line ~15)
     - `e2e.yaml` (line ~13)
   - Format: `GO_VERSION: 1.25` (major.minor only)
   - **Important**: Always search for ALL occurrences, as new workflows may be added

3. **`hack/Containerfile`**
   - Update builder image: `FROM docker.io/golang:1.25-alpine AS builder`
   - Check base image for latest stable: `FROM quay.io/fedora/fedora-minimal:XX`
   - To find latest Fedora stable: Check [Fedora releases](https://fedoraproject.org/wiki/Releases)

#### Procedure

```bash
# 1. Check current environment Go version
go version  # e.g., go version go1.25.4 linux/amd64

# 2. Update go.mod
# - Set go directive to major.minor.patch (e.g., 1.25.0)
# - Set toolchain to exact version from environment (e.g., go1.25.4)

# 3. Update ALL .github/workflows/*.yaml files
# - Search for all GO_VERSION references: grep -rn "GO_VERSION" .github/workflows/
# - Update in go.yaml, pre_linters.yaml, pre_reviewer.yaml, e2e.yaml
# - Set GO_VERSION to major.minor only (e.g., 1.25)

# 4. Update hack/Containerfile
# - Update golang builder image to match major.minor
# - Optionally update fedora-minimal base image

# 5. Resolve dependencies
go mod tidy

# 6. Validate changes
make test
make vet
make test-lint  # May show pre-existing YAML lint issues - that's OK

# 7. Commit changes
git add go.mod go.sum .github/workflows/*.yaml hack/Containerfile
git commit -m "chore: bump Go version to X.Y.Z

Updated Go version from X.Y.Z to A.B.C with toolchain A.B.D
to use the latest Go version available in the build environment.

Changes:
- Updated go directive to A.B.C
- Updated toolchain to goA.B.D
- Updated CI workflows GO_VERSION to A.B (4 workflow files)
- Updated hack/Containerfile golang image to A.B-alpine
- Resolved dependencies with go mod tidy

Validation:
- ✅ make test - all tests passed
- ✅ make vet - no issues found

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

#### Expected Results

- **Build**: `make build` should succeed
- **Tests**: All existing tests should pass
- **Linting**: Go code should pass linting (YAML workflow linting may have pre-existing issues)
- **CI**: Workflows should use new Go version

#### Common Issues

- **YAML linting errors**: These are typically pre-existing issues in `.github/workflows/*.yaml` files and unrelated to Go version changes
- **Dependency conflicts**: Run `go mod tidy` to resolve; if issues persist, check for incompatible dependencies
- **Toolchain mismatch**: Ensure `toolchain` directive matches the exact patch version in your environment

---

### Dependency Management

**When to use**: Update, add, or remove Go dependencies.

#### Reducing Dependencies (SBOM Optimization)

**Reference commit**: `26e8975` - "refactor: replace pkg/errors with stdlib fmt.Errorf"

##### Replace `github.com/pkg/errors` with stdlib

**Common pattern**: Replace external error handling libraries with Go's built-in error wrapping.

```go
// Before
import "github.com/pkg/errors"
errors.New("message")
errors.Wrap(err, "context")
errors.Wrapf(err, "context: %v", value)

// After
import "fmt"
fmt.Errorf("message")
fmt.Errorf("context: %w", err)
fmt.Errorf("context %v: %w", value, err)
```

**Files typically affected**:
- `pkg/cmd/**/*.go`
- `pkg/run/**/*.go`
- `internal/**/*.go`

**Validation steps**:
1. Search for all `errors.New`, `errors.Wrap`, `errors.Wrapf` usages
2. Replace with `fmt.Errorf` equivalents using `%w` for wrapping
3. Remove `"github.com/pkg/errors"` import
4. Remove from `go.mod` direct dependencies
5. Run `go mod tidy`
6. Run `make test` and `make build`

##### Upgrade YAML library (v2 → v3)

**Pattern**: Consolidate duplicate YAML libraries.

```go
// Before
import "gopkg.in/yaml.v2"

// After
import "gopkg.in/yaml.v3"
```

**Note**: API is mostly compatible; check for any breaking changes in complex YAML operations.

#### Dependency Analysis Commands

```bash
# Count total dependencies
go list -m all | wc -l

# Find direct dependencies
grep -A 100 "^require (" go.mod | grep -v "^)"

# Check why a dependency is needed
go mod why <package>

# Find duplicate/similar dependencies
go mod graph | grep <pattern>

# View dependency tree for a specific package
go mod graph | grep "^github.com/your/project " | head -30
```

---

## Validation Procedures

### Standard Validation Checklist

Before committing any changes, run:

```bash
# 1. Resolve dependencies
go mod tidy

# 2. Build the project
make build

# 3. Run tests
make test

# 4. Run go vet
make vet

# 5. Run linting (optional, may have pre-existing issues)
make test-lint
```

### Expected Test Results

- **make test**: All tests should pass (currently ~10 packages)
- **make vet**: Should complete with no output (no issues)
- **make test-lint**: May show YAML workflow formatting issues (pre-existing)
  - Go code linting is currently commented out in Makefile
  - YAML issues in `.github/workflows/*.yaml` are acceptable if unrelated to changes

### Test Output Interpretation

```bash
# Good test run
ok  	github.com/redhat-openshift-ecosystem/opct/internal/opct/summary	0.014s

# Failed test (requires investigation)
FAIL	github.com/redhat-openshift-ecosystem/opct/internal/opct/summary [build failed]

# No test files (acceptable)
?   	github.com/redhat-openshift-ecosystem/opct/cmd/opct	[no test files]
```

---

## Contributing Guidelines

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>: <description>

[optional body]

[optional footer(s)]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependency updates, tooling, etc.)

**Examples**:

```
chore: bump Go version to 1.25.0

feat: add support for custom plugin manifests

refactor: replace pkg/errors with stdlib fmt.Errorf

fix: correct error handling in baseline reporting
```

### AI Assistant Footer

Include the following footer in commits made by AI assistants:

```
🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### Branch Naming

- Feature branches: `feature/<description>`
- Bug fixes: `fix/<description>`
- Development tasks: `dev/<description>`
- Review/testing: `review-<description>`

### Pull Request Guidelines

1. Create a descriptive PR title following Conventional Commits
2. Include a detailed description of changes
3. Reference related issues with `Fixes #123` or `Relates to #123`
4. Ensure all CI checks pass
5. Request review from maintainers

---

## Additional Resources

- **Release Process**: [docs/devel/release.md](docs/devel/release.md)
- **Component Updates**: [docs/devel/update.md](docs/devel/update.md)
- **Developer Guide**: [docs/devel/guide.md](docs/devel/guide.md)
- **User Documentation**: [https://redhat-openshift-ecosystem.github.io/opct/](https://redhat-openshift-ecosystem.github.io/opct/)
- **Contributing Guide**: [CONTRIBUTING.md](CONTRIBUTING.md)

---

## Document Maintenance

This document should be updated when:
- New development patterns are established
- Common tasks are identified that need standardization
- AI assistants encounter repeated questions or issues
- Development workflows change significantly

**Last Updated**: 2025-11-19
**Maintainer**: OPCT Development Team
