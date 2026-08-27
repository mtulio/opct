---
name: webui-report-test
description: Build, regenerate, and serve the OPCT web UI report for browser testing. Use when modifying report.html, report.css, or any template/chart changes.
allowed-tools: Bash, Read
---

# Web UI Report Test

Build the project, regenerate the report from a test archive, and serve it for browser testing.

## When to use

- After modifying `data/templates/report/report.html` or `report.css`
- After changing report data generation in `internal/report/data.go`
- After changing metrics generation in `internal/openshift/mustgathermetrics/`
- When testing any web UI layout, chart, or table changes
- Chat API testing: only when [PR #213](https://github.com/redhat-openshift-ecosystem/opct/pull/213) is merged or checked out locally

## Workflow

### 1. Build

```bash
make build
```

### 2. Generate report and serve

Set up paths, server cleanup, and validate before any destructive cleanup:

```bash
TEST_ID=4.20.0-0.nightly-2026-05-04-230007-20260510-HighlyAvailable-aws-External
REPORT_DIR=~/opct/tmp/${TEST_ID}__report
REPORT_ROOT="${HOME}/opct/tmp"
HTTP_PID=""
REPORT_PID=""

cleanup_servers() {
  [ -n "${HTTP_PID}" ] && kill "${HTTP_PID}" 2>/dev/null
  [ -n "${REPORT_PID}" ] && kill "${REPORT_PID}" 2>/dev/null
}
trap cleanup_servers EXIT

REAL_DIR="$(realpath -m "${REPORT_DIR}")"
case "${REAL_DIR}" in
  "${REPORT_ROOT}"/*) ;;
  *) echo "refusing to delete outside ${REPORT_ROOT}: ${REAL_DIR}" >&2; exit 1 ;;
esac
[ -L "${REPORT_DIR}" ] && { echo "refusing to delete symlink: ${REPORT_DIR}" >&2; exit 1; }

cleanup_servers
```

**Default (frontend/layout changes — no chat API):**
```bash
rm -rf "${REPORT_DIR:?}"
./build/opct-linux-amd64 report -s "${REPORT_DIR}" --skip-server ~/opct/results/${TEST_ID}.tar.gz

python3 -m http.server 9090 --bind 127.0.0.1 --directory "${REPORT_DIR}" &
HTTP_PID=$!
echo "Static server PID: ${HTTP_PID} (http://127.0.0.1:9090)"
```

**With chat API (PR #213 only — Go server blocks if run in foreground):**
```bash
rm -rf "${REPORT_DIR:?}"
REPORT_LOG="/tmp/opct-report-${TEST_ID}.log"
./build/opct-linux-amd64 report -s "${REPORT_DIR}" \
  --server-address 127.0.0.1:9090 \
  ~/opct/results/${TEST_ID}.tar.gz \
  > "${REPORT_LOG}" 2>&1 &
REPORT_PID=$!
echo "Report server PID: ${REPORT_PID} (logs: ${REPORT_LOG}, http://127.0.0.1:9090)"
```

When finished testing, stop servers with `kill "${HTTP_PID}" "${REPORT_PID}"` or exit the shell (the `EXIT` trap cleans up owned PIDs).

### 3. Test in browser

Open http://127.0.0.1:9090 and verify:

**Always check:**
- The page you modified works correctly
- At least 2-3 other pages (Summary, Checks, Network) have no layout regressions
- Browser resize works for responsive layouts

**For chart changes:**
- Charts load and are interactive (zoom, expand)
- etcd split-pane layout works, divider is draggable

**For layout changes (headline, tables):**
- Verify styling on ALL pages that use `pageHeadline` (Summary, Checks, etcd, Network, Runtime, etc.)

**For chatbot changes (PR #213 only):**
- Chat bubble appears in bottom-right
- `/api/v1/chat/status` returns enabled status
- Streaming responses work
- Minimize/maximize/close buttons work
- Stop button cancels mid-stream responses
- Sessions auto-save and load correctly

## Important notes

- The Go server (`opct report` without `--skip-server`) blocks the terminal — always run it in background (see above) or use a second terminal
- Python server only serves static files — no chat API; bind to `127.0.0.1` for local testing
- Use PID tracking and the `EXIT` trap instead of `pkill` — avoids killing unrelated processes
- Chat requires [PR #213](https://github.com/redhat-openshift-ecosystem/opct/pull/213) merged or checked out locally, plus Vertex AI or Anthropic API credentials (see CLAUDE.md)
- Report generation takes ~15 seconds for the test archive
- Always validate `REPORT_DIR` is under `~/opct/tmp` and not a symlink before `rm -rf`
