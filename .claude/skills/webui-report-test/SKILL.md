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

## Workflow

### 1. Build

```bash
make build
```

### 2. Regenerate report

```bash
TEST_ID=4.20.0-0.nightly-2026-05-04-230007-20260510-HighlyAvailable-aws-External
rm -rf ~/opct/tmp/${TEST_ID}__report
./build/opct-linux-amd64 report -s ~/opct/tmp/${TEST_ID}__report --skip-server ~/opct/results/${TEST_ID}.tar.gz
```

### 3. Serve for browser testing

```bash
# Kill any existing server first
pkill -f "python3 -m http.server 9090" 2>/dev/null
sleep 0.5

# Serve (run in background)
python3 -m http.server 9090 --directory ~/opct/tmp/${TEST_ID}__report
```

### 4. Test in browser

Open http://localhost:9090 and verify:
- The page you modified works correctly
- At least 2-3 other pages (Summary, Checks, Network) have no layout regressions
- Charts load and are interactive (if applicable)
- Browser resize works for responsive layouts

## Key architecture notes

- Report is a Vue.js 2 SPA with Go template delimiters `[[` / `]]`
- Content rendered via `this.menuBody` string + `v-html` directive
- Split-pane pages use `menuBodyRight` with `v-if`/`v-else` — never shared containers
- Chart data comes from static JSON files at `./metrics/`, NOT from `opct-report.json`
- Gate chart rendering on `this.report.summary.features.hasMetricsData`
- Chart.js uses `<canvas>` elements; render via `$nextTick()` after DOM update
