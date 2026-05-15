---
name: ci-triage
description: Triage an OPCT periodic CI job failure — analyze test failures, check job history, query flake rates, check existing Jira bugs, and draft new bugs for real failures.
allowed-tools: Bash, Read, WebFetch, WebSearch
---

# CI Triage

Investigate a failing OPCT periodic CI job, determine if failures are real or transient, and draft Jira bugs for real failures.

## When to use

- You receive a Slack notification about an OPCT periodic job failure
- You want to check if a failing job needs a new Jira bug
- Periodic triage review of OPCT CI health

## Usage

Provide a Prow job URL:

```
investigate the job failure https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-redhat-openshift-ecosystem-opct-main-4.18-platform-none-vsphere-upgrade/2054789108642877440
```

Or shorter:

```
triage https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-release-main-nightly-4.22-opct-platform-external-aws/2053264314218844160
```

## What the agent does

1. **Parses metadata** from the job URL (OCP version, platform, provider, workflow)
2. **Fetches job details** via `ci:fetch-prowjob-json`
3. **Analyzes test failures** via `ci:prow-job-analyze-test-failure`
4. **Checks job history** to find consecutive failures and first failure date
5. **Queries flake rates** via `ci:fetch-test-report` (Sippy)
6. **Checks existing Jira bugs** via `ci:check-if-jira-regression-is-ongoing`
7. **Classifies each failure**: KNOWN_FLAKE, EXISTING_BUG, or NEW_FAILURE
8. **Drafts a Jira bug** for new failures (asks for approval before filing)
9. **Presents a triage summary** table

## Example job URLs

OPCT repository jobs:
```
https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-redhat-openshift-ecosystem-opct-main-4.18-platform-none-vsphere-upgrade/2054789108642877440
https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-redhat-openshift-ecosystem-opct-main-4.18-platform-external-vsphere-upgrade/2054955197750317056
```

Release repository nightly jobs:
```
https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-release-main-nightly-4.19-opct-external-aws-ccm/2051551527281102848
https://prow.ci.openshift.org/view/gs/test-platform-results/logs/periodic-ci-openshift-release-main-nightly-4.22-opct-platform-external-aws/2053264314218844160
```

## Expected output

```
## Triage Summary: periodic-ci-...-4.18-platform-none-vsphere-upgrade

Job: https://prow.ci.openshift.org/view/gs/...
History: https://prow.ci.openshift.org/job-history/gs/...
Status: FAILED (failing since 2026-03-15, 12 consecutive failures)

| # | Test | Classification | Action |
|---|------|---------------|--------|
| 1 | [sig-network] ... | KNOWN_FLAKE (12%) | No action |
| 2 | [sig-auth] ... | EXISTING_BUG | OCPBUGS-1234 |
| 3 | [sig-storage] ... | NEW_FAILURE | Bug draft below |

### Bug Draft (pending approval)
Title: OPCT/CI job failure: 4.18-platform-none-vsphere-upgrade
Project: OCPBUGS
...

Shall I file this bug?
```

## After triage

- **Approve bug**: tell the agent to file it
- **Dismiss**: if you determine the failure is not actionable
- **Investigate further**: ask the agent to dig deeper into specific test failures

## Jira bug fields (auto-populated)

| Field | Source |
|-------|--------|
| Title | `OPCT/CI job failure: {VERSION}-{PLATFORM}-{PROVIDER}-{WORKFLOW}` |
| Release Blocker | Rejected |
| Labels | splatteam, needs-refinement |
| Components | OPCT / Other |
| Parent | OPCT-400 |
| Affects Version | Extracted from job name |

## Marketplace skills used

- `ci:fetch-prowjob-json` — job metadata
- `ci:prow-job-analyze-test-failure` — failure analysis
- `ci:prow-job-analyze-install-failure` — install failure analysis (if applicable)
- `ci:fetch-test-report` — Sippy flake rates
- `ci:check-if-jira-regression-is-ongoing` — existing bug check
- `jira:create-bug` + `jira:ocpbugs` — bug creation
