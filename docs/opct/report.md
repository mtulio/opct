# `opct report`

`opct report` generates a comprehensive report from the conformance workflow executed on an OpenShift/OKD cluster.

The report is available in two formats:

- Command Line Interface (CLI)
- Web User Interface (Web UI)

The `report` command is the most important command used to review the results.

The command reads the archive/result file and:
- provides a summary of standard conformance suites required to validate an OpenShift/OKD installation
- extracts must-gather information, such as the state of objects and logs
- extracts counters of error patterns from workloads (must-gather)
- applies SLOs (Checks) created from results executed in well-known (reference) deployments
- reports failed SLOs, providing guidance on failed items to help/guide the OpenShift/OKD installation
- exports counters and exposes them to be used as indicators when reviewing an OpenShift/OKD installation
- builds a local Web UI app allowing you to navigate throughout the results extracted from the archive, such as e2e test failure logs, e2e test metadata, e2e test documentation, CAMGI report, metrics report, and counters

To begin, see the `Usage` section.

## Usage

- Basic usage:
```sh
opct report <archive.tar.gz>
```

- Advanced usage exposing/serving the WebUI:

```sh
opct report <archive.tar.gz> --save-to /tmp/results --log-level=debug
```

## Options

```txt
--8<-- "docs/assets/output/opct-report.txt"
```

## Examples

### CLI Report

To generate a CLI report, run the following command:

```sh
./opct report ./opct_202502230112_624d66d9-6354-40d1-b847-82a55f57d444.tar
```

The output will be similar to:

```txt
--8<-- "docs/assets/output/opct-report_example-cli.txt"
```

### Web UI Report

To generate a Web UI report, follow these steps:

1. Run the command:

```sh
./opct report --save-to ./results ./opct_202502230112_624d66d9-6354-40d1-b847-82a55f57d444.tar
```

2. Open your browser and navigate to [localhost:9090](http://localhost:9090).

![opct report](../assets/output/opct-report_example-webui.jpg)
