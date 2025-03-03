# `opct adm baseline [actions]`

Administrative commands to manipulate baseline results.

Baseline results are conformance executions accepted for
use as reference results during the review process.

The baseline results are automated executions that are
automatically published to the result services.

The `report` command automatically consumes the latest valid result from a specific
`OpenShift Version` and `Platform Type` in the filter pipeline (`Failed Filter API`),
making inferences about common failures in that specific release which **may** not be directly
related to the environment being validated.

The filter helps to isolate:

- Flaky tests in CI environment
- Permanent failures in the platform and release tested
- Test environment issues

## Options

```txt
--8<-- "docs/assets/output/opct-adm-baseline.txt"
```

## Usage

```sh
opct adm baseline [flags]
opct adm baseline [command]
```

Overview of commands:

- `opct adm baseline list`: List available baselines.
- `opct adm baseline get`: Get a summary of a specific baseline.
- `opct adm baseline publish`: (restricted) Publish artifacts to the OPCT services.
- `opct adm baseline indexer`: (restricted) Rebuild the index of the report service to serve the latest baseline summary.

## Available Commands

- `get`: Get a baseline result to be used in the review process.
- `indexer`: (Administrative usage) Rebuild the indexer for baseline in the backend.
- `list`: List all available baseline results by OpenShift version, provider, and platform type.
- `publish`: Publish a baseline result to be used in the review process.

## Global Flags

- `-h, --help`: Help for baseline
- `--kubeconfig string`: Kubeconfig for target OpenShift cluster
- `--log-level string`: Logging level (default "info")

## Examples

- List the latest summary's artifacts by version and platform type:

```bash
$ opct adm baseline list
+---------------+--------+---------+--------------+------------------------------+
| ID            | TYPE   | RELEASE | PLATFORMTYPE | NAME                         |
+---------------+--------+---------+--------------+------------------------------+
| 4.13_None     | latest | 4.13    | None         | 4.13_None_20240729023635     |
| 4.14_External | latest | 4.14    | External     | 4.14_External_20250301032412 |
| 4.14_None     | latest | 4.14    | None         | 4.14_None_20250301031949     |
| 4.15_External | latest | 4.15    | External     | 4.15_External_20250302082341 |
| 4.15_None     | latest | 4.15    | None         | 4.15_None_20250302032525     |
| 4.16_External | latest | 4.16    | External     | 4.16_External_20250302141824 |
| 4.16_None     | latest | 4.16    | None         | 4.16_None_20250302200651     |
| 4.17_AWS      | latest | 4.17    | AWS          | 4.17_AWS_20240813215230      |
| 4.17_External | latest | 4.17    | External     | 4.17_External_20250303140431 |
| 4.17_None     | latest | 4.17    | None         | 4.17_None_20250302205650     |
| 4.18_AWS      | latest | 4.18    | AWS          | 4.18_AWS_20241220033058      |
| 4.18_External | latest | 4.18    | External     | 4.18_External_20250201044656 |
| 4.18_None     | latest | 4.18    | None         | 4.18_None_20250303103134     |
+---------------+--------+---------+--------------+------------------------------+
```

- Review the summary for a latest artifact from a specific release:

```bash
$ opct-devel adm baseline get --platform=External --release=4.18

INFO[2025-03-03T16:40:34-03:00] Getting latest baseline result by release and platform: 4.18/External 
INFO[2025-03-03T16:40:35-03:00] Baseline result processed from archive: 4.18.0-0.nightly-2025-01-31-141502-20250201-HighlyAvailable-vsphere-External.tar.gz 
>> Example serializing and extracting plugin failures for  20-openshift-conformance-validated
[0]: [sig-instrumentation] Prometheus [apigroup:image.openshift.io] when installed on the cluster shouldn't report any alerts in firing state apart from Watchdog and AlertmanagerReceiversNotConfigured [Early][apigroup:config.openshift.io] [Skipped:Disconnected] [Suite:openshift/conformance/parallel]
[1]: [sig-olmv1][OCPFeatureGate:NewOLM][Skipped:Disconnected] OLMv1 operator installation should install a cluster extension [Suite:openshift/conformance/parallel]
[2]: [sig-network] LoadBalancers [Feature:LoadBalancer] should be able to preserve UDP traffic when server pod cycles for a LoadBalancer service on the same nodes [Skipped:alibabacloud] [Skipped:aws] [Skipped:baremetal] [Skipped:ibmcloud] [Skipped:kubevirt] [Skipped:nutanix] [Skipped:openstack] [Skipped:ovirt] [Skipped:vsphere] [Suite:openshift/conformance/parallel] [Suite:k8s]
....
```

- Publish artifacts to the OPCT services (**administrative only**):

```bash
export PROCESS_FILES="4.15.0-rc.7-20240221-HighlyAvailable-vsphere-None.tar.gz
4.15.0-rc.7-20240221-HighlyAvailable-vsphere-External.tar.gz
4.15.0-rc.1-20240110-HighlyAvailable-vsphere-External.tar.gz
4.15.0-20240228-HighlyAvailable-vsphere-None.tar.gz
4.15.0-20240228-HighlyAvailable-vsphere-External.tar.gz"

# Upload each baseline artifact
for PF in $PROCESS_FILES;
do
    opct adm baseline publish --log-level=debug "$HOME/opct/$PF";
done

# re-index
opct adm baseline indexer

# Expire CloudFront cache if you received an error (manual step):
# - AWS Console: AWS CloudFront > Distributions > Select Distribution > Invalidations > Create new expiring '/*'
# - AWS CLI: $ aws cloudfront create-invalidation --distribution-id <id> --paths /*

# Check the latest baseline data
opct-devel adm baseline list --all

# check all baseline data
opct-devel adm baseline list
```
