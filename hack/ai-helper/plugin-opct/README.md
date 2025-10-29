# OPCT Plugin

Comprehensive automation and analysis tools for OpenShift Platform Compatibility Tool (OPCT) workflows, including session management, test result analysis, and must-gather integration.

## Overview

The OPCT plugin provides AI-powered assistance for OpenShift Platform Compatibility Tool workflows, enabling automated analysis of conformance test results, cluster diagnostics, and validation reports. This plugin integrates seamlessly with OPCT's report endpoint and must-gather data to provide comprehensive insights into OpenShift cluster compatibility and health.

## Key Features

- **Session Management**: Initialize and organize OPCT result sessions from tarball files
- **Test Result Analysis**: Comprehensive analysis of OPCT test execution results
- **Must-Gather Integration**: Extract and analyze cluster diagnostic data
- **Report Generation**: Generate detailed analysis reports with actionable insights
- **Validation Checks**: Analyze OPCT validation rules and SLO/SLI metrics
- **Failure Investigation**: Deep-dive analysis of test failures with root cause identification

## Available Commands

### `/opct:init`
Initialize OPCT sessions from result tarballs.

**Usage:**
```bash
/opct:init <tarball-file> [session-dir]
```

**Examples:**
```bash
# Auto-generate session directory from tarball name
/opct:init opct_202510091157_67175c7a-0cc0-42c2-877f-d2652daa75fc.tar.gz

# Use custom session directory
/opct:init opct_results.tar.gz my-analysis-session

# Full path example
/opct:init /path/to/opct_202510091157_67175c7a.tar.gz
```

**Features:**
- Extracts OPCT results to structured session directories
- Automatically extracts must-gather data from artifacts collector
- Creates organized directory structure for analysis workflows
- Validates tarball integrity and handles extraction errors
- Provides session summary and next steps guidance

## Session Structure

OPCT sessions are organized as follows:

```
{session-dir}/
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

## OPCT Integration

### Report Endpoint Integration
The plugin integrates with OPCT's localhost report endpoint:
- **Main Report**: `http://localhost:9090/opct-report.json`
- **Schema**: `http://localhost:9090/opct-report-schema.json`
- **Test Failures**: `http://localhost:9090/failures-<job>/<test_id>-failure.txt`
- **System Output**: `http://localhost:9090/failures-<job>/<test_id>-systemOut.txt`

### Data Sources
- **Test Results**: Individual test execution results with status and metadata
- **Validation Checks**: OPCT validation rules and SLO/SLI metrics
- **Error Counters**: Test error pattern analysis
- **Node Information**: Cluster node configuration and capacity
- **Plugin Results**: Execution results from OPCT plugins

## Analysis Capabilities

### Test Result Analysis
- **Component Grouping**: Organize failures by SIG (sig-auth, sig-network, etc.)
- **Priority Classification**: Identify test importance via [Early], [Late], [Serial] tags
- **Root Cause Analysis**: Extract failure patterns from test logs
- **Platform-Specific Issues**: Detect BareMetal, assisted-installer specific problems
- **Pattern Recognition**: Identify common errors, timeouts, resource constraints

### Must-Gather Integration
- **Cluster Health**: Analyze cluster operators, nodes, and pods
- **Network Diagnostics**: OVN/SDN health and connectivity analysis
- **Storage Analysis**: Persistent volume and claim status
- **Event Correlation**: Cross-reference test failures with cluster events
- **Performance Issues**: etcd latency, storage I/O analysis

### Validation Checks
- **SLO/SLI Analysis**: Service Level Objective and Indicator evaluation
- **Rule Violations**: OPCT validation rule compliance checking
- **Threshold Monitoring**: Target vs. current metric comparison
- **Documentation Links**: Reference to detailed check documentation

## Workflow Integration

### With Must-Gather Plugin
```bash
# Initialize OPCT session
/opct:init opct_results.tar.gz

# Analyze must-gather data
/must-gather:analyze session-dir/must-gather/
```

### With Report Server
```bash
# Start OPCT report server
opct report -s session-dir/results/ opct_results.tar.gz

# Analyze via HTTP endpoint
curl http://localhost:9090/opct-report.json
```

## Error Handling

### Common Scenarios
- **Tarball Validation**: File existence, readability, and integrity checks
- **Extraction Failures**: Detailed error messages for tar/xz command failures
- **Missing Must-Gather**: Graceful handling when must-gather archive is not found
- **Permission Issues**: Clear error messages for file system access problems
- **Overwrite Protection**: Confirmation prompts for existing session directories

### Recovery Actions
- **Partial Extraction**: Report what was successfully extracted
- **Missing Dependencies**: Check for required tools (tar, xz)
- **Corrupted Archives**: Verify tarball integrity before extraction
- **Storage Issues**: Monitor disk space for large extractions

## Best Practices

### Session Management
1. **Descriptive Naming**: Use timestamps or test identifiers in session names
2. **Storage Planning**: Ensure sufficient disk space for results and must-gather
3. **Backup Strategy**: Consider backing up important sessions
4. **Cleanup**: Remove old sessions to manage disk usage

### Analysis Workflow
1. **Initialize Session**: Start with `/opct:init` to organize data
2. **Verify Extraction**: Check that both results and must-gather are available
3. **Analyze Results**: Use OPCT analysis commands to examine failures
4. **Correlate Data**: Cross-reference test failures with cluster diagnostics
5. **Generate Reports**: Create comprehensive analysis reports
6. **Share Findings**: Export results for team review and action

## Troubleshooting

### Common Issues
- **Large Files**: Monitor extraction progress for very large tarballs
- **Corrupted Archives**: Verify tarball integrity before extraction
- **Permission Problems**: Check file ownership and access rights
- **Missing Tools**: Ensure `tar` and `xz` commands are available
- **Network Issues**: Verify OPCT report server connectivity

### Debug Information
- **Verbose Output**: Enable detailed logging for troubleshooting
- **File Validation**: Check file permissions and integrity
- **Tool Availability**: Verify required commands are installed
- **Storage Space**: Monitor available disk space during extraction

## Related Plugins

- **Must-Gather Plugin**: Analyze cluster diagnostic data
- **OpenShift Plugin**: OpenShift development workflow automation
- **CI Plugin**: Continuous integration job analysis
- **Utils Plugin**: General-purpose development utilities

## Documentation

- **OPCT Documentation**: https://github.com/redhat-openshift-ecosystem/opct
- **OpenShift Conformance**: https://github.com/cncf/k8s-conformance
- **OPCT Validation Rules**: https://github.com/redhat-openshift-ecosystem/opct/blob/main/docs/review/rules.md
- **OpenShift Origin Tests**: https://github.com/openshift/origin

## Contributing

When contributing to the OPCT plugin:
1. Follow the existing command structure and documentation format
2. Ensure scripts handle errors gracefully with clear messages
3. Test with various OPCT result tarball formats
4. Update documentation for new features or changes
5. Consider integration with other plugins for enhanced workflows
