# opct adm parse-etcd-logs

Extract the information from JUnit files to get insights from a test execution.

## Options

```txt
--8<-- "docs/assets/output/opct-adm-parse-junit.txt"
```

## Usage

!!! info "Added to OPCT in release: v0.6+"

- Command: `opct adm parse-junit <path-to-junit-file>`

Args:

- `<path-to-junit-file>`: the path to the JUnit file.

## Examples

- Create the test list:

```sh
cat << EOF >./test-list.txt
[sig-scheduling] SchedulerPredicates [Serial] validates that NodeSelector is respected if not matching [Conformance] [Suite:openshift/conformance/serial/minimal] [Suite:k8s]"
....
EOF
```

- Run `openshift-tests` targetting the test list:

> More information how to use [`openshift-tests`](./../../guides/features/running-e2e.md)

```sh
./openshift-tests run \
    -f test-list.txt \
    --monitor="etcd-log-analyzer" \
    --junit-dir=./results
```

- Run the parser command:

```bash
$ ./build/opct-linux-amd64 adm parse-junit ./results/junit_e2e__20250313-192600.xml
Summary:
- File: ./results/junit_e2e__20250313-192600.xml
- Total: 26
- Pass: 23
- Skipped: 0
- Failures: 3

JUnit Attributes:
- XMLName: { testsuite}
- Name: openshift-tests
- Tests: 26
- Skipped: 0
- Failures: 3
- Time: 1557
- Property: {TestVersion 4.19.0-202502102307.p0.gc22aad1.assembly.stream.el9-c22aad1}

#> Passed tests (23): 
[sig-api-machinery] Watchers should observe an object deletion if it stops meeting the requirements of the selector [Conformance] [Suite:openshift/conformance/parallel/minimal] [Suite:k8s]
[sig-scheduling] SchedulerPredicates [Serial] validates that NodeAffinity is respected if not matching [Suite:openshift/conformance/serial] [Suite:k8s]
[sig-scheduling] SchedulerPredicates [Serial] validates that NodeSelector is respected if not matching [Conformance] [Suite:openshift/conformance/serial/minimal] [Suite:k8s]
[sig-instrumentation] MetricsGrabber should grab all metrics slis from API server. [Suite:openshift/conformance/parallel] [Suite:k8s]
[sig-scheduling] SchedulerPredicates [Serial] validates resource limits of pods that are allowed to run [Conformance] [Suite:openshift/conformance/serial/minimal] [Suite:k8s]
[sig-scheduling] SchedulerPredicates [Serial] validates pod overhead is considered along with resource limits of pods that are allowed to run verify pod overhead is accounted for [Suite:openshift/conformance/serial] [Suite:k8s]
[sig-cli] Kubectl client Kubectl diff should check if kubectl diff finds a difference for Deployments [Conformance] [Suite:openshift/conformance/parallel/minimal] [Suite:k8s]
[sig-network][Feature:Router] The HAProxy router should enable openshift-monitoring to pull metrics [Skipped:Disconnected] [Suite:openshift/conformance/parallel]
[sig-network][Feature:Router] The HAProxy router should expose prometheus metrics for a route [apigroup:route.openshift.io] [Skipped:Disconnected] [Suite:openshift/conformance/parallel]
[sig-auth][Feature:OAuthServer] OAuth server has the correct token and certificate fallback semantics [apigroup:user.openshift.io] [Suite:openshift/conformance/parallel]
[sig-auth][Feature:OpenShiftAuthorization][Serial] authorization TestAuthorizationResourceAccessReview should succeed [apigroup:authorization.openshift.io] [Suite:openshift/conformance/serial]
[sig-cli] oc builds new-build [apigroup:build.openshift.io] [Skipped:Disconnected] [Suite:openshift/conformance/parallel]
[sig-network][Feature:Router] The HAProxy router should expose the profiling endpoints [Skipped:Disconnected] [Suite:openshift/conformance/parallel]
[sig-api-machinery][Feature:APIServer] anonymous browsers should get a 403 from / [Suite:openshift/conformance/parallel]
[sig-arch] Managed cluster should ensure platform components have system-* priority class associated [Suite:openshift/conformance/parallel]
[sig-node][apigroup:config.openshift.io] CPU Partitioning cluster infrastructure should be configured correctly [Suite:openshift/conformance/parallel]
[sig-network][Feature:Router] The HAProxy router should expose a health check on the metrics port [Skipped:Disconnected] [Suite:openshift/conformance/parallel]
[sig-builds][Feature:Builds] oc new-app should succeed with an imagestream [apigroup:build.openshift.io] [Skipped:Disconnected] [Suite:openshift/conformance/parallel]
[sig-auth][Feature:UserAPI] users can manipulate groups [apigroup:user.openshift.io][apigroup:authorization.openshift.io][apigroup:project.openshift.io] [Suite:openshift/conformance/parallel]
[sig-cli] oc basics can get version information from API [Suite:openshift/conformance/parallel]
[sig-auth][Feature:OAuthServer] well-known endpoint should be reachable [apigroup:route.openshift.io] [apigroup:oauth.openshift.io] [Suite:openshift/conformance/parallel]
Cluster should be stable after installation is complete
Cluster should be stable before test is started

#> Failed tests (3): 
"[sig-instrumentation] Prometheus [apigroup:image.openshift.io] when installed on the cluster shouldn't report any alerts in firing state apart from Watchdog and AlertmanagerReceiversNotConfigured [Early][apigroup:config.openshift.io] [Skipped:Disconnected] [Suite:openshift/conformance/parallel]"
"[sig-network] services when running openshift ipv4 cluster ensures external ip policy is configured correctly on the cluster [apigroup:config.openshift.io] [Serial] [Suite:openshift/conformance/serial]"
"[sig-network][Feature:tap] should create a pod with a tap interface [apigroup:k8s.cni.cncf.io] [Suite:openshift/conformance/parallel]"

#> Skipped tests (0):
```
