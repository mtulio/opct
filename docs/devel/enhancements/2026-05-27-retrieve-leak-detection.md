# Plan: Add leak detection to `opct retrieve` using leaktk patterns

## Context

The `opct retrieve` command downloads conformance results from the cluster and saves them as a tar.gz archive. Currently it applies file-level patches (redacting `internalRegistryPullSecret`) and removes large unnecessary files (`packagemanifests`). However, there's no content-based scanning for leaked credentials — AWS keys, SSH private keys, OpenShift tokens, etc. could be present in must-gather logs, resource dumps, or config files.

Goal: Embed high-priority leak patterns from [leaktk/patterns](https://github.com/leaktk/patterns) directly into the existing tarball processing pipeline in `cleaner.go`, scanning file contents during the retrieve stream processing without impacting timing.

## Approach: Embedded patterns (no external dependency)

LeakTK is CLI-only (pre-v1.0, no stable Go library). Instead, we'll extract ~25 high-value regex patterns from leaktk/patterns TOML and compile them as Go structs. This integrates directly into the existing `processTarHeader()` in `cleaner.go` — zero external dependencies, no subprocess overhead.

## Implementation

### 1. New file: `internal/cleaner/leakpatterns.go`

Define leak pattern structs and the curated pattern set:

```go
type LeakPattern struct {
    ID          string
    Description string
    Regex       *regexp.Regexp
    Keywords    []string       // pre-filter optimization
}

type LeakFinding struct {
    File        string
    Pattern     string
    Line        int
    Match       string         // redacted preview (first/last chars only)
}
```

Each pattern must include a comment referencing the source:
```go
// Source: https://github.com/leaktk/patterns
// Pattern ID: sOZiHxUBVFc (leaktk v8.27.0)
```

**Priority patterns to embed (~25 rules):**

| Category | Pattern ID | Description |
|----------|-----------|-------------|
| OpenShift/K8s | `sOZiHxUBVFc` | OpenShift User Token |
| OpenShift/K8s | `vAAom0bPHy8` | Kubernetes Service Account JWT |
| OpenShift/K8s | `gpfGmO3HH64` | Container Registry Authentication |
| AWS | `LAJoYTdoQH4` | AWS IAM Unique Identifier |
| AWS | `9j_rmwDeioM` | AWS Secret Access Key |
| Azure | `zl044yuux24` | Azure AD Client Secret |
| Azure | — | Azure Storage Account Access Key |
| GCP | — | GCP API Key |
| Private keys | `ePK9whPQPpY` | Private Key (PEM header) |
| Private keys | `RVee3wT2Z4I` | Base64 Encoded OpenSSH Private Key |
| Generic | — | Generic password/secret in key=value |
| Generic | — | Generic token in YAML/JSON |
| GitHub | — | GitHub Personal Access Token |

Fetch the actual regex patterns from https://github.com/leaktk/patterns/blob/main/target/main.toml at implementation time.

### 2. New file: `internal/cleaner/leakscanner.go`

Scanner function that takes file content bytes and returns findings:

```go
func ScanContentForLeaks(filename string, content []byte) []LeakFinding
```

- Skip binary files (check for null bytes in first 512 bytes)
- Skip files > 10MB (avoid scanning large tarballs-within-tarballs)
- Apply keyword pre-filter before regex (optimization from leaktk)
- Scan line-by-line for regex matches
- Return findings with file path, pattern description, line number

### 3. Modify: `internal/cleaner/cleaner.go`

In `processTarHeader()`, after reading file content and before writing to `tarWriter`:

```go
// After existing patch/skip logic, before writing:
if len(content) > 0 && len(content) < maxLeakScanSize {
    findings := ScanContentForLeaks(header.Name, content)
    for _, f := range findings {
        log.Warnf("Potential leak detected in %s (line %d): %s", f.File, f.Line, f.Pattern)
    }
}
```

**Behavior:** Log warnings only (don't block retrieve). The user sees what was found and can investigate. Future enhancement: add `--fail-on-leak` flag.

Add a new `LeakScanRules` variable alongside existing `JSONPatchRules` and `RemoveFilePatternRules`.

### 4. Add summary at end of retrieve

After archive is saved, print a leak scan summary:

```
INFO Leak scan: 3 potential findings in 2 files
WARN   resources/cluster/secrets.json:42 — AWS Secret Access Key
WARN   must-gather/logs/pod.log:188 — OpenShift User Token  
WARN   install-config.txt:15 — Private Key (PEM)
```

### 5. Tests: `internal/cleaner/leakscanner_test.go`

- Test each pattern against known test vectors (from leaktk pattern examples)
- Test false positive suppression (allowlist keywords)
- Test binary file skip
- Test large file skip
- Benchmark to ensure scanning doesn't add significant time

## Files to modify

| File | Change |
|------|--------|
| `internal/cleaner/leakpatterns.go` | New — pattern definitions |
| `internal/cleaner/leakscanner.go` | New — scanning logic |
| `internal/cleaner/leakscanner_test.go` | New — tests |
| `internal/cleaner/cleaner.go` | Hook scanner into `processTarHeader()` |

## Performance considerations

- **Keyword pre-filter:** Each pattern has keywords (e.g., `["akia", "asia"]` for AWS keys). Check if any keyword is present in the file content (case-insensitive) before running the full regex. This eliminates >99% of regex invocations.
- **Binary skip:** Don't scan binary files (tar archives, images, compressed data).
- **Size limit:** Skip files > 10MB.
- **Existing pipeline:** The tarball is already read file-by-file in memory. Scanning adds a regex pass on the already-loaded bytes — no additional I/O.

Based on the current retrieve timing (~14 seconds), the regex scanning should add <1 second for typical archives.

## Enhancement document

Save this plan as `docs/devel/enhancements/2026-05-27-retrieve-leak-detection.md` (create the `enhancements` directory if it doesn't exist).

## Skill updates

Add leak scanning knowledge to `opct-developer` agent (Related Skills section) and consider adding a reference in the `ci-triage` agent for awareness when reviewing partner archives.

The `leakpatterns.go` file header must reference the upstream pattern source:
```go
// Package cleaner provides leak detection patterns for scanning OPCT archives.
//
// Patterns are sourced from the leaktk/patterns project:
//   https://github.com/leaktk/patterns
//
// To update patterns, fetch the latest merged TOML from:
//   https://github.com/leaktk/patterns/blob/main/target/main.toml
//
// LeakTK documentation:
//   https://github.com/leaktk/leaktk
//   https://github.com/leaktk/leaktk/blob/main/docs/scan.md
```

## Verification

1. `make build && make test`
2. Run retrieve on the existing test archive: `./build/opct-linux-amd64 retrieve --log-level=debug`
3. Verify no false positives on known-clean archives
4. Plant a test secret (e.g., `AKIA...` in a file) and verify detection
5. Benchmark: compare retrieve time with/without leak scanning
