# Web UI Developer Agent

You are working on the OPCT web UI report — a Vue.js 2 single-page application that displays OpenShift conformance test results. The report is served by a Go HTTP server on port 9090.

> **Note:** The AI chat widget and `internal/chat/` backend are experimental and tracked in [PR #213](https://github.com/redhat-openshift-ecosystem/opct/pull/213). They are not on `main` yet — skip chat-specific tasks unless working on that branch.

## Architecture Overview

### Frontend (single HTML file)
- **Template**: `data/templates/report/report.html` — Go template with `[[`/`]]` delimiters
- **Styles**: `data/templates/report/report.css`
- **Framework**: Vue.js 2 (CDN), Bootstrap (CDN), axios (CDN)
- **Charts**: Chart.js 4.x + chartjs-plugin-zoom (CDN)

### Backend
- **Report server**: `pkg/cmd/report/report.go` — Go HTTP server on port 9090
- **Data generation**: `internal/report/data.go` — renders templates, produces `opct-report.json`
- **Metrics**: `internal/openshift/mustgathermetrics/` — generates Plotly-format JSON chart data

### Data flow
1. `opct report -s <dir> <archive>` extracts results, generates JSON, renders HTML templates
2. Go HTTP server serves `<dir>/` as static files
3. Vue app fetches `opct-report.json` via axios, renders pages via `v-html` into `menuBody`

## Critical Rules

### Layout isolation
- Pages that need multi-column layouts (e.g., etcd split-pane) MUST use `v-if`/`v-else` to conditionally render the split container
- NEVER add `d-flex` or width constraints to a shared container — it breaks all other pages
- Clear inline styles in `changeMenuCleanup()` when navigating away

### Vue rendering
- Content is injected as HTML strings via `v-html` — `<script>` tags inside are ignored
- Use `this.$nextTick(() => { ... })` to run code after DOM update
- Chart.js needs `<canvas>` elements (not `<div>`)
- Always clean up state in `changeMenuCleanup()` when adding new data properties

### AI Attribution
All commits include `Co-Authored-By: Claude <noreply@anthropic.com>`. All GitHub comments end with `— AI Claude`.

### Testing workflow
- ALWAYS rebuild (`make build`) after Go changes
- ALWAYS regenerate report after template changes (use an explicit path variable):
  ```bash
  TEST_ID=<your-test-archive-id>
  REPORT_DIR=~/opct/tmp/${TEST_ID}__report
  REPORT_ROOT="${HOME}/opct/tmp"
  REAL_DIR="$(realpath -m "${REPORT_DIR}")"
  case "${REAL_DIR}" in
    "${REPORT_ROOT}"/*) ;;
    *) echo "refusing to delete outside ${REPORT_ROOT}: ${REAL_DIR}" >&2; exit 1 ;;
  esac
  [ -L "${REPORT_DIR}" ] && { echo "refusing to delete symlink: ${REPORT_DIR}" >&2; exit 1; }
  rm -rf "${REPORT_DIR:?}"
  opct report -s "${REPORT_DIR}" --skip-server <archive>
  ```
- Test with `python3 -m http.server` or the Go server (see `webui-report-test` skill)
- Check at least 3 pages (Summary, Checks, etcd) after any layout change

## File Reference

| Area | File | Purpose |
|------|------|---------|
| Frontend | `data/templates/report/report.html` | Vue.js SPA template (~1600 lines) |
| Styles | `data/templates/report/report.css` | All CSS including split-pane, charts |
| Report server | `pkg/cmd/report/report.go` | CLI command, HTTP server setup |
| Report data | `internal/report/data.go` | Report struct, template rendering, JSON serialization |
| Metrics gen | `internal/openshift/mustgathermetrics/plotly.go` | Plotly JSON chart data generation |

## CDN Libraries (loaded in report.html head)

| Library | Version | Purpose |
|---------|---------|---------|
| Vue.js | 2.7.16 | Frontend framework |
| Bootstrap | 4.6.2 | Layout and components |
| BootstrapVue | 2.23.1 | Vue Bootstrap components |
| axios | 1.7.9 | HTTP client |
| Chart.js | 4.4.9 | Charts (etcd page) |
| chartjs-adapter-date-fns | 3.0.0 | Time-series X axis |
| Hammer.js | 2.0.8 | Touch gestures for zoom |
| chartjs-plugin-zoom | 2.2.0 | Drag-to-zoom, pan |
| marked.js | 15.0.7 | Markdown rendering (chat — PR #213) |

**Always pin CDN URLs to specific versions** (e.g., `vue@2.7.16`, not `vue@2` or `@latest`). The main report template currently uses unpinned URLs — when modifying CDN references, pin to the versions above for reproducible UI behavior.

## Common Tasks

### Add a new page/menu item
1. Add button in the left nav menu (around line 43-78)
2. Add case in `changeMenu()` switch (around line 267-306)
3. Create `changeMenuNewPage()` method that builds `this.menuBody`
4. Use `this.pageHeadline` at the top for consistent styling

### Add a new chart
1. Define chart metadata in Vue `data()` with `id`, `path`, `title`
2. Build `<canvas>` placeholders in the target panel
3. Fetch JSON via `axios.get()` in `$nextTick()` callback
4. Create Chart.js instance with `responsive: true`
5. Map Plotly JSON format (`data[].x`, `data[].y`, `data[].name`) to Chart.js datasets

## Experimental: Chat API (PR #213 only)

When working on the chat feature branch ([PR #213](https://github.com/redhat-openshift-ecosystem/opct/pull/213)):

### Additional files

| Area | File | Purpose |
|------|------|---------|
| Chat handler | `internal/chat/handler.go` | HTTP handlers, SSE streaming, Claude API loop |
| Chat tools | `internal/chat/tools.go` | Tool definitions and execution against report data |
| Chat sessions | `internal/chat/session.go` | Session CRUD, JSON persistence |
| Chat prompt | `internal/chat/prompt.go` | System prompt with file override |

### Chat backend rules
- Vertex AI: prefer `GOOGLE_CLOUD_LOCATION` over `CLOUD_ML_REGION` (which may be `global`)
- Model IDs: use alias form like `claude-sonnet-4-5` (not dated versions)
- Tool results are read from the report directory filesystem — tools must handle missing files gracefully
- SSE events: `text` (streaming tokens), `tool_call` (tool invocation), `done` (final text), `error`
- Floating widget states: use CSS classes with `!important` for minimize/maximize — never toggle `display` on individual children via JS
- Test chat with Go server (not python) — python can't serve the API endpoints

### Add a new chat tool (PR #213)
1. Add tool definition in `internal/chat/tools.go` `ToolDefinitions()` function
2. Add input struct if needed (with jsonschema tags)
3. Add case in `Execute()` switch
4. Implement the execution method reading from `te.reportDir`
5. Tool results are returned as JSON strings to Claude
