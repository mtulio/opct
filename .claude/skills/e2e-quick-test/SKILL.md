---
name: e2e-quick-test
description: Quick end-to-end validation on a live OpenShift cluster with minimal test count for rapid development iteration.
allowed-tools: Bash, Read
---

# E2E Quick Test

Rapid validation workflow for testing OPCT changes on a live OpenShift cluster.

## Prerequisites

- `oc` CLI installed and configured (same cluster as `KUBECONFIG`)
- `KUBECONFIG` environment variable set to a valid cluster
- Permissions to manage the `opct` namespace and read cluster events
- Dedicated compute node configured (requires cluster-scoped node permissions):
  ```bash
  opct adm e2e-dedicated taint-node --yes
  ```
  Lists worker `Node` resources cluster-wide and applies the required label and `NoSchedule` taint. Requires permission to `get`/`list`/`patch` nodes — **cluster-admin recommended**.
- `gh` CLI installed (only when checking out PR branches via `gh pr checkout`)

## Quick Test Command

```bash
make build-linux-amd64; ./build/opct-linux-amd64 destroy; ./build/opct-linux-amd64 run -w --dev-count=1
```

**What it does:**
1. Rebuilds the CLI binary with latest local changes
2. Destroys any existing OPCT run to ensure clean state
3. Runs OPCT workflow with:
   - `-w` (wait mode): blocks until completion
   - `--dev-count=1`: runs only 1 test per plugin (development mode)

**Expected runtime:** ~5-10 minutes (vs. hours for full conformance)

**Note:** `-w` blocks the terminal until the run completes. Open a second terminal for `watch`/`oc` commands below, or omit `-w` and poll with `opct status`.

## Development Workflow

### Testing a PR branch locally

```bash
# 1. Fetch and checkout PR branch (requires gh CLI)
gh pr checkout ${PR_NUMBER}  # or: git fetch origin pull/${PR_NUMBER}/head:pr-${PR_NUMBER} && git checkout pr-${PR_NUMBER}

# 2. Quick validate
make build-linux-amd64
./build/opct-linux-amd64 destroy
./build/opct-linux-amd64 run -w --dev-count=1

# 3. Monitor plugin progress (in another terminal, or remove -w from the run command):
watch -n 5 'oc get pods -n opct'

# 4. Check for specific issues (e.g., cache-init errors)
oc get events -n opct --field-selector involvedObject.kind=Pod --sort-by='.lastTimestamp'
```

### Validating template changes

When modifying plugin templates in `data/templates/plugins/*.yaml`:

```bash
# 1. Make template changes
# 2. Rebuild (embeds new templates)
make build-linux-amd64

# 3. Clean previous run
./build/opct-linux-amd64 destroy

# 4. Test with minimal workload
./build/opct-linux-amd64 run -w --dev-count=1

# 5. Verify init containers succeeded
oc describe pod -n opct -l sonobuoy-component=plugin | grep -A 10 "Init Containers"
```

### Testing specific scenarios

**Cache-init validation:**
```bash
# After modifying cache-init in plugin templates
make build-linux-amd64 && \
./build/opct-linux-amd64 destroy && \
./build/opct-linux-amd64 run -w --dev-count=1

# Check cache-init container status
oc get pods -n opct -l sonobuoy-component=plugin -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.initContainerStatuses[?(@.name=="cache-init")].state}{"\n"}{end}'
```

**Retrieve/download validation:**
```bash
# Test retrieve functionality
./build/opct-linux-amd64 retrieve --watch

# Test with signal handling (Ctrl+C during download)
./build/opct-linux-amd64 retrieve -w  # Press Ctrl+C during "Extracting..."
```

## Common Issues

| Issue | Diagnosis | Fix |
|-------|-----------|-----|
| `Init:Error` on plugin pods | `oc logs <pod> -c <init-container>` | Check init container permissions/commands |
| `ImagePullBackOff` | `oc describe pod <pod>` | Verify image references in templates |
| Pods stuck in `Pending` | `oc describe pod <pod>` | Check node resources, taints/tolerations |
| `destroy` fails with "not found" | Expected if no prior run | Safe to ignore, continue with `run` |
| No dedicated node | `oc get nodes -l node-role.kubernetes.io/worker` | Run `opct adm e2e-dedicated taint-node --yes` |

## Full Test vs. Quick Test

| Mode | Command | Runtime | Use Case |
|------|---------|---------|----------|
| **Quick** | `--dev-count=1` | ~5-10 min | Development iteration, template changes, plugin fixes |
| **Full** | (no flag) | ~2-6 hours | Final validation, conformance certification |

## Cleanup

```bash
# Remove OPCT namespace and all resources
./build/opct-linux-amd64 adm cleanup

# Or just destroy current run (keeps namespace)
./build/opct-linux-amd64 destroy
```

## Related Skills

- `/go-validate` — Validate Go code before building
- `/opct-runtime` — Understand plugin execution architecture
- `/ci-triage` — Debug CI failures
