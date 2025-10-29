---
description: Initialize OPCT session from result tarball - extracts results and must-gather data
argument-hint: [tarball-file] [session-dir]
---

## Name
opct:init

## Synopsis
```
/opct:init [tarball-file] [session-dir]
```

## Description

The `opct:init` command initializes a new OPCT session by extracting and organizing data from an OPCT result tarball. This command creates a structured session directory containing both the test results and must-gather diagnostic data, enabling comprehensive analysis of OpenShift Platform Compatibility Tool (OPCT) execution results.

The command performs the following operations:
1. **Session Directory Creation**: Creates a session directory based on the tarball filename (without .tar.gz extension)
2. **Results Extraction**: Extracts the OPCT results to `{session-dir}/results/`
3. **Must-Gather Extraction**: Extracts must-gather data from `plugins/99-openshift-artifacts-collector/results/global/artifacts_must-gather.tar.xz` to `{session-dir}/must-gather/`

## Prerequisites

**Required Files:**
- OPCT result tarball file (typically ending with .tar.gz)
- The tarball must contain the standard OPCT structure with results and must-gather artifacts

**Required Tools:**
- `tar` command for extraction
- `xz` command for decompressing must-gather archives

## Session Directory Structure

After successful initialization, the session directory will contain:

```
{session-dir}/
├── results/                    # OPCT test results and metadata
│   ├── opct-report.json        # Main OPCT report
│   ├── opct-report-schema.json # Report schema (if available)
│   └── plugins/                # Plugin execution results
│       └── 99-openshift-artifacts-collector/
│           └── results/
│               └── global/
│                   └── artifacts_must-gather.tar.xz
└── must-gather/                # Extracted must-gather diagnostic data
    └── must-gather-opct/       # Must-gather directory structure
        └── [must-gather contents]
```

## Implementation

The command uses the `init` skill which performs:

1. **Validate Input**: Check if tarball file exists and is readable
2. **Determine Session Directory**:
   - If `session-dir` is provided, use it
   - Otherwise, derive from tarball filename (remove .tar.gz extension)
3. **Create Session Structure**: Create the session directory and subdirectories
4. **Extract Results**: Extract the entire tarball to `{session-dir}/results/`
5. **Extract Must-Gather**:
   - Locate `artifacts_must-gather.tar.xz` in the results
   - Extract it to `{session-dir}/must-gather/`
6. **Verify Extraction**: Confirm both extractions completed successfully

## Error Handling

**Common Error Scenarios:**
- **Tarball Not Found**: Clear error message if tarball file doesn't exist
- **Permission Issues**: Check for read permissions on tarball and write permissions on target directory
- **Missing Must-Gather**: Warning if must-gather archive is not found in expected location
- **Extraction Failures**: Detailed error messages for tar/xz command failures

**Recovery Actions:**
- If session directory already exists, prompt user for confirmation before overwriting
- If must-gather extraction fails, continue with results extraction and report the issue

## Return Value

The command outputs:
- **Success**: Confirmation of session creation with directory paths
- **Session Summary**: Brief overview of extracted contents
- **Next Steps**: Suggested commands for analysis

## Examples

1. **Basic initialization with auto-generated session directory**:
   ```
   /opct:init opct_202510091157_67175c7a-0cc0-42c2-877f-d2652daa75fc.tar.gz
   ```
   Creates session directory: `opct_202510091157_67175c7a-0cc0-42c2-877f-d2652daa75fc/`

2. **Initialize with custom session directory**:
   ```
   /opct:init opct_results.tar.gz my-analysis-session
   ```
   Creates session directory: `my-analysis-session/`

3. **Initialize from full path**:
   ```
   /opct:init /path/to/opct_202510091157_67175c7a-0cc0-42c2-877f-d2652daa75fc.tar.gz
   ```

## Arguments

- **$1** (tarball-file): **Required**. Path to the OPCT result tarball file
- **$2** (session-dir): **Optional**. Custom name for the session directory. If not provided, derived from tarball filename

## Notes

- **Session Naming**: Session directories follow OPCT result file naming standards (without .tar.gz extension)
- **Must-Gather Location**: Must-gather data is expected at `plugins/99-openshift-artifacts-collector/results/global/artifacts_must-gather.tar.xz`
- **Overwrite Protection**: Existing session directories require confirmation before overwriting
- **Tool Integration**: Initialized sessions work seamlessly with other OPCT analysis commands
- **Storage Requirements**: Ensure sufficient disk space for both results and must-gather data extraction

## Related Commands

- `/opct:analyze` - Analyze OPCT test results and failures
- `/opct:report` - Generate comprehensive OPCT analysis reports
- `/must-gather:analyze` - Analyze must-gather diagnostic data
