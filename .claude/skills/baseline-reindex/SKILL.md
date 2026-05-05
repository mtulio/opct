---
name: baseline-reindex
description: Re-index the OPCT baseline results stored in S3 and invalidate the CloudFront cache. Use when baseline results are missing, after new results are published, or when the index appears corrupted.
allowed-tools: Bash, Read, WebFetch
---

# Baseline Reindex

Re-index the OPCT baseline results stored in S3 and invalidate the CloudFront cache.

## When to use

- After new baseline results have been published to S3
- When the index.json is missing entries or appears corrupted
- After manual S3 object additions/deletions

## Prerequisites

- AWS credentials configured (via env vars, default profile, or instance role)
- `OPCT_ENABLE_ADM_BASELINE=1` environment variable set

## Architecture

The baseline indexer operates on an S3 bucket (`opct-archive`) under the prefix `api/v0/result/summary/`.

**Object naming convention:** `{ocpVersion}_{platformType}_{timestamp}.json`
- Example: `4.19_None_20260501214947.json`

**Special objects:**
- `index.json` — master index with all results and latest map
- `{version}_{platform}_latest.json` — alias copies of the latest result per version/platform

**Incremental indexing flow:**
1. Lists all S3 objects (paginated via `ListObjectsPages`)
2. Loads existing `index.json` from S3
3. Builds a set of already-indexed paths
4. Only fetches and parses metadata for new objects
5. Recalculates latest entries across all results
6. Writes updated `index.json` and copies latest aliases
7. Invalidates CloudFront cache paths

Falls back to full reindex if no existing index is found.

## Key files

- `internal/report/baseline/indexer.go` — `CreateBaselineIndex()`, `ListObjects()`, `loadIndexFromS3()`, `fetchObjectMetadata()`
- `internal/report/baseline/aws.go` — S3 and CloudFront client creation
- `internal/report/baseline/baseline.go` — `BaselineConfig`, S3 client setup, API readers
- `internal/report/baseline/data.go` — `BaselineData`, metadata extraction from `setup.api`
- `internal/report/baseline/uploader.go` — `UploadBaseline()` for publishing results

## Commands

```bash
# Build
make build-linux-amd64

# Run the indexer
OPCT_ENABLE_ADM_BASELINE="1" opct adm baseline indexer --log-level=debug

# List baseline results
opct adm baseline list

# Publish a result
OPCT_ENABLE_ADM_BASELINE="1" opct adm baseline publish --log-level=debug <artifact.tar.gz>
```

## S3 bucket structure

```
opct-archive/
├── uploads/                              # Raw artifact archives
│   └── *.tar.gz
└── api/v0/result/summary/               # Processed summaries
    ├── index.json                        # Master index (~530KB)
    ├── 4.13_None_20240729023635.json     # Individual results (~40KB each)
    ├── 4.13_None_latest.json             # Latest alias (copy of newest)
    └── ...
```

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `OPCT_ENABLE_ADM_BASELINE` | unset | Must be `1` to enable admin baseline features |
| `OPCT_EXP_BUCKET_NAME` | `opct-archive` | Override S3 bucket name |
| `OPCT_EXP_BUCKET_REGION` | `us-east-1` | Override S3 bucket region |
| `AWS_SHARED_CREDENTIALS_FILE` | default | Path to AWS credentials file (used in CI) |

## CI integration

The indexer runs in OpenShift CI after baseline results are published:
- Step reference: https://steps.ci.openshift.org/reference/opct-conformance-test-results
- Script: `opct-conformance-test-results-commands.sh`
- Credentials: `/var/run/vault/opct/.awscred-vmc-ci-opct-uploader` (via `AWS_SHARED_CREDENTIALS_FILE`)
- CloudFront distribution: `E3MJR7MT6EHHJC`
- Local clone of CI step registry: check `ci-operator/step-registry/opct/` in the `openshift/release` repo

## Known issues and history

- **OPCT-399**: S3 pagination was capped at 1000 objects, `_latest.json` files polluted the index, CloudFront credentials required hardcoded profile. Fixed in PR #212.
- Index currently has ~1,156 objects spanning OCP versions 4.13–4.22.
