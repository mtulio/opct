# Using Claude Code CLI for OPCT Development

This guide shows how to use [Claude Code](https://claude.com/claude-code), Anthropic's official CLI for Claude, to automate common development tasks in the OPCT project using the instructions in [`CLAUDE.md`](../../CLAUDE.md).

## Overview

Claude Code is an interactive CLI tool that helps with software engineering tasks. The `CLAUDE.md` file in the project root contains structured instructions that Claude Code uses to execute tasks consistently and correctly.

## Prerequisites

1. **Install Claude Code**: Follow the [official installation guide](https://github.com/anthropics/claude-code)
2. **API Key**: Set up your Anthropic API key as instructed
3. **Project Context**: Ensure you're in the OPCT project root directory

## Getting Started

Start Claude Code in the project directory:

```bash
cd /path/to/opct
claude-code
```

## Common Development Tasks

### 1. Bumping Go Version

**Scenario**: You've updated your Go installation and want to update the project to use the new version.

**Example Prompts**:

```
Update to Go 1.26
```

```
Bump Go version to 1.26.0
```

```
I just updated Go to 1.26.2, please update the project to use this version
```

**What Claude Code will do**:
- Check your current Go version from the environment
- Update `go.mod` (go directive and toolchain)
- Update all 4 `.github/workflows/*.yaml` files (GO_VERSION)
- Update `hack/Containerfile` (golang builder image)
- Update `hack/go-imports.sh` and `hack/go-staticcheck.sh` (golang container images)
- Run `go mod tidy`
- Validate with `make test`, `make vet`, `make build`
- Commit changes with a conventional commit message

**Expected Output**:
```
✅ Updated go.mod: go 1.26.0, toolchain go1.26.2
✅ Updated 4 workflow files: GO_VERSION: 1.26
✅ Updated hack/Containerfile: golang:1.26-alpine
✅ Updated hack scripts: golang:1.26
✅ Tests passed
✅ Committed: "chore: bump Go version to 1.26.0"
```

---

### 2. Updating Kubernetes and OpenShift Client Libraries

**Scenario**: New stable releases of Kubernetes or OpenShift are available, and you want to update the client libraries.

**Example Prompts**:

```
Bump the dependencies to the latest stable versions
```

```
Update k8s and openshift client libraries
```

```
Update OpenShift client libraries to release-4.23
```

```
Update Kubernetes client-go to v0.35
```

**What Claude Code will do**:
- Check latest stable branches on GitHub:
  - `github.com/openshift/api/branches` for latest `release-X.Y`
  - `github.com/kubernetes/client-go/branches` for latest `release-X.Y`
- Update OpenShift libraries (e.g., `release-4.22` → `release-4.23`)
- Update Kubernetes libraries (e.g., `v0.34.x` → `v0.35.x`)
- Update go.mod comments to reflect correct branch references
- Run `go mod tidy`
- Validate with `make test`, `make vet`, `make build`
- Commit changes with detailed version information

**Expected Output**:
```
✅ Checked GitHub branches for latest versions
✅ OpenShift: release-4.22 → release-4.23
✅ Kubernetes: v0.34.1 → v0.35.0
✅ Updated go.mod and go.sum
✅ Tests passed
✅ Committed: "chore: bump k8s and openshift client libraries to latest stable"
```

---

### 3. Reducing Dependencies (SBOM Optimization)

**Scenario**: You want to reduce the project's dependency footprint by removing unnecessary or duplicate dependencies.

**Example Prompts**:

```
I want to decrease the dependencies (SBOM) of the project. Please review and provide suggestions.
```

```
Analyze dependencies and suggest optimizations
```

```
Replace github.com/pkg/errors with stdlib
```

```
Find and remove duplicate YAML libraries
```

**What Claude Code will do**:
- Analyze `go.mod` for direct and indirect dependencies
- Identify duplicates, unofficial libraries, or replaceable dependencies
- Suggest specific changes with impact analysis
- If approved, execute the changes:
  - Replace external packages with stdlib equivalents
  - Consolidate duplicate libraries
  - Update all affected Go files
  - Run `go mod tidy`
  - Validate with tests and builds
  - Commit changes

**Expected Output**:
```
Analysis:
- Found 177 total dependencies, 22 direct
- Identified duplicates: yaml.v2 + yaml.v3, 2 UUID libraries
- Suggested Phase 1 optimizations:
  ✅ Replace pkg/errors with stdlib fmt.Errorf
  ✅ Upgrade yaml.v2 to yaml.v3

Changes:
✅ Updated 11 Go files
✅ Removed 2 direct dependencies
✅ Tests passed
✅ Committed: "refactor: replace pkg/errors with stdlib fmt.Errorf"
```

---

### 4. Code Refactoring

**Example Prompts**:

```
Refactor the error handling in pkg/run/run.go to use errors.Join
```

```
Extract the duplicate code in internal/opct/summary/*.go into a shared function
```

```
Update all functions in pkg/cmd/* to follow the project's error handling pattern
```

**What Claude Code will do**:
- Analyze the specified code
- Identify refactoring opportunities
- Propose changes with explanations
- Execute approved changes
- Ensure no behavioral changes (tests still pass)
- Commit with descriptive messages

---

### 5. Adding New Features

**Example Prompts**:

```
Add a new command 'opct adm validate-cluster' that checks cluster prerequisites
```

```
Implement support for exporting metrics to JSON format in addition to CSV
```

**What Claude Code will do**:
- Analyze existing code patterns
- Design the feature following project conventions
- Implement the feature with:
  - New command/function code
  - Tests for new functionality
  - Documentation updates
  - Integration with existing code
- Validate all tests pass
- Commit changes with detailed descriptions

---

### 6. Bug Fixes

**Example Prompts**:

```
The build is failing with error X, please investigate and fix
```

```
Tests are failing in pkg/status/status_test.go, please debug and fix
```

```
The report command crashes when processing large archives, please fix
```

**What Claude Code will do**:
- Analyze the error/failure
- Investigate the code to find root cause
- Propose and implement a fix
- Add tests to prevent regression
- Validate the fix with tests and builds
- Commit with bug fix description

---

## Advanced Usage

### Multi-Step Tasks

You can provide complex, multi-step instructions:

```
I want to add a new feature that:
1. Adds a --format flag to the report command
2. Supports JSON, YAML, and CSV formats
3. Refactors the existing CSV code into a formats package
4. Adds tests for each format
5. Updates the documentation
```

Claude Code will break this down, plan the implementation, and execute each step.

### Asking Questions During Execution

Claude Code may ask clarifying questions:

```
> I noticed there are two possible approaches for implementing this:
> 1. Use the existing renderer pattern
> 2. Create a new formatter interface
> Which approach would you prefer?
```

### Reviewing Changes Before Commit

You can ask Claude Code to show you changes before committing:

```
Please show me the diff before committing
```

```
What files will be modified?
```

---

## Best Practices

### 1. Be Specific with Prompts

**Good**:
```
Update to Go 1.26.0
```

**Less Ideal**:
```
Update Go
```

### 2. Reference Existing Patterns

**Good**:
```
Add a new adm subcommand following the pattern in pkg/cmd/adm/baseline
```

**Less Ideal**:
```
Add a new command
```

### 3. Request Validation

**Good**:
```
Update dependencies and make sure all tests pass
```

**Less Ideal**:
```
Update dependencies
```
(Claude Code will validate by default based on CLAUDE.md, but being explicit helps)

### 4. Incremental Changes

For complex tasks, break them into steps:

```
First, please analyze the current error handling patterns in the codebase
```

Then after review:

```
Now update all pkg/cmd/* files to use the consistent pattern
```

---

## Understanding Claude Code's Workflow

Based on the instructions in `CLAUDE.md`, Claude Code typically follows this workflow:

1. **Understand the Request**: Analyzes your prompt and determines the task type
2. **Plan**: Creates a plan based on CLAUDE.md instructions
3. **Execute**: Runs the necessary commands and makes code changes
4. **Validate**: Runs tests, linting, and builds
5. **Commit**: Creates a conventional commit with detailed description
6. **Report**: Shows you what was done and the results

### Example Session

```
$ claude-code

You: Update to Go 1.26

Claude Code: I'll update the project to use Go 1.26. Let me check your current Go version...

Claude Code: Current environment: go version go1.26.1 linux/amd64

Claude Code: I'll update:
- go.mod (go 1.26.0, toolchain go1.26.1)
- 4 workflow files (.github/workflows/*.yaml)
- hack/Containerfile
- hack/*.sh scripts

[Progress indicators...]

Claude Code: ✅ All files updated
✅ Dependencies resolved
✅ Tests passed (10 packages)
✅ Build successful

Committed: chore: bump Go version to 1.26.0
```

---

## Troubleshooting

### Claude Code can't find instructions

**Issue**: Claude Code doesn't seem to follow the project conventions.

**Solution**: Make sure `CLAUDE.md` exists in the project root:
```bash
ls -l CLAUDE.md
```

### Changes are incomplete

**Issue**: Some files weren't updated during a Go version bump.

**Solution**: Be explicit in your prompt:
```
You missed updating .github/workflows/e2e.yaml, please complete the Go bump
```

Claude Code learns from this feedback and will update CLAUDE.md to prevent future issues.

### Tests fail after dependency updates

**Issue**: Tests fail after updating client libraries.

**Solution**:
```
The tests are failing, please investigate and fix the issues
```

Claude Code will:
- Analyze the test failures
- Identify API changes in updated dependencies
- Update the code to be compatible
- Validate all tests pass

---

## Updating CLAUDE.md

As the project evolves, you can ask Claude Code to update its own instructions:

```
Please add instructions to CLAUDE.md for updating the container base image to the latest Fedora version
```

```
Update the dependency bump procedure in CLAUDE.md to include checking for breaking changes
```

This creates a self-improving documentation loop where the AI assistant's capabilities grow with the project.

---

## Examples from Real Usage

### Example 1: Complete Go Version Bump

**Prompt**:
```
I just updated the go from my build env to 1.25, please bump the project to this version
```

**Result**:
- Updated 6 files (go.mod, 4 workflows, Containerfile, 2 hack scripts)
- Resolved dependencies
- All tests passed
- Clean commit with detailed changelog

### Example 2: Dependency Optimization

**Prompt**:
```
I want to decrease the dependencies (SBOM) of opct project. Please review and provide suggestions.
```

**Result**:
- Analyzed 177 dependencies
- Identified 10 optimizations across 2 phases
- Executed Phase 1 (with user approval):
  - Replaced pkg/errors in 11 files
  - Upgraded yaml.v2 to yaml.v3
  - Removed 2 dependencies
  - All tests passed

### Example 3: Multi-Step Update

**Prompt**:
```
Bump the dependencies to the latest stable versions
```

**Result**:
- Checked GitHub branches for latest releases
- Updated OpenShift: release-4.18 → release-4.22
- Updated Kubernetes: v0.32.2 → v0.34.1
- Updated 6 related dependencies
- All validations passed
- Comprehensive commit with version details

---

## Additional Resources

- **CLAUDE.md**: Project-specific instructions for Claude Code
- **[Claude Code Documentation](https://github.com/anthropics/claude-code)**: Official documentation
- **[Conventional Commits](https://www.conventionalcommits.org/)**: Commit message format used by this project
- **OPCT Development Docs**: See other documents in `docs/devel/` for manual procedures

---

## Feedback and Improvements

If you find that Claude Code doesn't handle a task correctly:

1. Provide corrective feedback in the session
2. Ask Claude Code to update CLAUDE.md with the learnings
3. The next developer (or AI session) will benefit from the improved instructions

Example:
```
You missed updating the hack scripts. Please:
1. Update them now
2. Add this to CLAUDE.md so it doesn't happen again
```

This collaborative approach continuously improves the automation.
