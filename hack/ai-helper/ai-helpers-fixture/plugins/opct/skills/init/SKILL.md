---
name: OPCT Session Initializer
description: |
  Initialize OPCT sessions by extracting and organizing result tarballs containing test results
  and must-gather diagnostic data. Use this skill when the user wants to create a new OPCT
  session from a tarball file for analysis.

  Triggers: "init opct session", "extract opct results", "create opct session", "initialize opct",
  "setup opct analysis", "opct init", "extract must-gather from opct", "organize opct data"
---

# OPCT Session Initializer Skill

Initialize OPCT (OpenShift Platform Compatibility Tool) sessions by extracting and organizing result tarballs containing test results and must-gather diagnostic data.

## Overview

This skill provides session initialization for:
- **OPCT Result Tarballs**: Extract and organize OPCT test execution results
- **Must-Gather Data**: Extract diagnostic data from OPCT artifacts collector
- **Session Management**: Create structured directories for analysis workflows
- **Data Organization**: Separate results and must-gather data for independent analysis

## OPCT Result Tarball Structure

OPCT result tarballs typically contain:
```
opct_results.tar.gz
├── opct-report.json                    # Main OPCT report
├── opct-report-schema.json             # Report schema (optional)
├── plugins/                            # Plugin execution results
│   └── 99-openshift-artifacts-collector/
│       └── results/
│           └── global/
│               └── artifacts_must-gather.tar.xz  # Must-gather archive
└── [other OPCT artifacts]
```

## Instructions

### 1. Validate Input Parameters

**Required Parameters:**
- `tarball_file`: Path to OPCT result tarball (.tar.gz)
- `session_dir`: Optional custom session directory name

**Validation Steps:**
```bash
# Check if tarball exists and is readable
if [ ! -f "$tarball_file" ]; then
    echo "Error: Tarball file '$tarball_file' not found"
    exit 1
fi

# Check if tarball is readable
if [ ! -r "$tarball_file" ]; then
    echo "Error: Cannot read tarball file '$tarball_file'"
    exit 1
fi
```

### 2. Determine Session Directory

**Auto-Generated Name:**
```bash
# Extract base name without .tar.gz extension
if [ -z "$session_dir" ]; then
    session_dir=$(basename "$tarball_file" .tar.gz)
fi
```

**Custom Name:**
- Use provided `session_dir` parameter
- Validate directory name for filesystem compatibility

### 3. Create Session Structure

**Directory Creation:**
```bash
# Create main session directory
mkdir -p "$session_dir"

# Create subdirectories
mkdir -p "$session_dir/results"
mkdir -p "$session_dir/must-gather"
```

**Overwrite Protection:**
```bash
# Check if session directory already exists
if [ -d "$session_dir" ] && [ "$(ls -A "$session_dir" 2>/dev/null)" ]; then
    echo "Warning: Session directory '$session_dir' already exists and is not empty"
    echo "This will overwrite existing data. Continue? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Operation cancelled"
        exit 0
    fi
fi
```

### 4. Extract OPCT Results

**Full Tarball Extraction:**
```bash
# Extract entire tarball to results directory
tar -xzf "$tarball_file" -C "$session_dir/results"

# Verify extraction
if [ $? -ne 0 ]; then
    echo "Error: Failed to extract OPCT results from '$tarball_file'"
    exit 1
fi
```

### 5. Extract Must-Gather Data

**Locate Must-Gather Archive:**
```bash
must_gather_archive="$session_dir/results/plugins/99-openshift-artifacts-collector/results/global/artifacts_must-gather.tar.xz"

if [ ! -f "$must_gather_archive" ]; then
    echo "Warning: Must-gather archive not found at expected location:"
    echo "  $must_gather_archive"
    echo "Skipping must-gather extraction"
else
    echo "Found must-gather archive, extracting..."

    # Extract must-gather to dedicated directory
    tar -xf "$must_gather_archive" -C "$session_dir/must-gather"

    if [ $? -ne 0 ]; then
        echo "Error: Failed to extract must-gather data"
        exit 1
    fi
fi
```

### 6. Verify Session Creation

**Structure Verification:**
```bash
# Check if results directory has content
if [ ! "$(ls -A "$session_dir/results" 2>/dev/null)" ]; then
    echo "Error: Results directory is empty after extraction"
    exit 1
fi

# Check if must-gather directory has content (if extracted)
if [ -d "$session_dir/must-gather" ] && [ "$(ls -A "$session_dir/must-gather" 2>/dev/null)" ]; then
    echo "Must-gather data successfully extracted"
else
    echo "Must-gather data not available or empty"
fi
```

## Session Directory Structure

After successful initialization:

```
{session_dir}/
├── results/                           # OPCT test results
│   ├── opct-report.json              # Main OPCT report
│   ├── opct-report-schema.json       # Report schema (if available)
│   ├── plugins/                      # Plugin execution results
│   │   └── 99-openshift-artifacts-collector/
│   │       └── results/
│   │           └── global/
│   │               └── artifacts_must-gather.tar.xz
│   └── [other OPCT artifacts]
└── must-gather/                      # Extracted must-gather data
    └── must-gather-opct/             # Must-gather directory
        └── [must-gather contents]
```

## Error Handling

### Common Error Scenarios

1. **Tarball Not Found**:
   ```bash
   Error: Tarball file '/path/to/file.tar.gz' not found
   ```

2. **Permission Issues**:
   ```bash
   Error: Cannot read tarball file '/path/to/file.tar.gz'
   Error: Cannot create session directory (permission denied)
   ```

3. **Extraction Failures**:
   ```bash
   Error: Failed to extract OPCT results from 'file.tar.gz'
   Error: Failed to extract must-gather data
   ```

4. **Missing Must-Gather**:
   ```bash
   Warning: Must-gather archive not found at expected location
   ```

### Recovery Actions

- **Existing Directory**: Prompt for confirmation before overwriting
- **Partial Extraction**: Report what was successfully extracted
- **Missing Dependencies**: Check for required tools (tar, xz)

## Output Format

**Success Output:**
```
================================================================================
OPCT SESSION INITIALIZED SUCCESSFULLY
================================================================================

Session Directory: /path/to/session_dir
Results Extracted: /path/to/session_dir/results/
Must-Gather Extracted: /path/to/session_dir/must-gather/

Next Steps:
- Analyze OPCT results: /opct:analyze /path/to/session_dir/results/
- Analyze must-gather: /must-gather:analyze /path/to/session_dir/must-gather/
- Generate report: /opct:report /path/to/session_dir/
```

**Warning Output:**
```
================================================================================
OPCT SESSION INITIALIZED WITH WARNINGS
================================================================================

Session Directory: /path/to/session_dir
Results Extracted: /path/to/session_dir/results/
Must-Gather: Not available (archive not found)

Note: Must-gather data was not found in the expected location.
OPCT results are available for analysis.
```

## Integration Points

### With OPCT Analysis Commands
- `/opct:analyze` - Analyze extracted test results
- `/opct:report` - Generate comprehensive analysis reports

### With Must-Gather Analysis
- `/must-gather:analyze` - Analyze extracted diagnostic data
- Cross-reference test failures with cluster issues

### With Session Management
- Session directories can be reused across multiple analysis commands
- Maintains separation between test results and diagnostic data

## Best Practices

1. **Session Naming**: Use descriptive names that include timestamps or test identifiers
2. **Storage Management**: Ensure sufficient disk space for both results and must-gather
3. **Backup Strategy**: Consider backing up important sessions before analysis
4. **Cleanup**: Remove old sessions to manage disk usage

## Troubleshooting

### Common Issues

1. **Large Tarball Files**:
   - Ensure sufficient disk space
   - Monitor extraction progress for very large files

2. **Corrupted Archives**:
   - Verify tarball integrity before extraction
   - Check for partial downloads

3. **Permission Problems**:
   - Ensure write permissions in target directory
   - Check file ownership and access rights

4. **Missing Tools**:
   - Verify `tar` and `xz` commands are available
   - Install required packages if missing

## Next Steps After Initialization

1. **Verify OPCT Report**: Check if `opct-report.json` is present and valid
2. **Analyze Test Results**: Use OPCT analysis commands to examine failures
3. **Correlate with Must-Gather**: Cross-reference test failures with cluster diagnostics
4. **Generate Reports**: Create comprehensive analysis reports
5. **Share Findings**: Export results for team review and action
