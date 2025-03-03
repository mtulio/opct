# Running Native Kubernetes e2e with opct and openshift-tests (The Hard Way)

This guide describes how to manually run the e2e tests using `opct`, `sonobuoy`, and `openshift-tests`.

This is an advanced step; use it if you are looking to expand the tests or use the
test orchestration provided by opct/sonobuoy to create custom tests.

## Prerequisites

- Install OPCT

- Grant permissions to the test environment

```bash
oc adm policy add-scc-to-group privileged system:authenticated system:serviceaccounts
oc adm policy add-scc-to-group anyuid system:authenticated system:serviceaccounts
```

- Install yq

```bash
VERSION="v4.44.6"
BINARY=yq_linux_amd64
wget -O $HOME/bin/yq https://github.com/mikefarah/yq/releases/download/${VERSION}/${BINARY} &&\
    chmod +x $HOME/bin/yq
```

## Running Kubernetes conformance tests provided by `sonobuoy`

To begin with, you need to define the group of tests it will run.

Sonobuoy provides rich documentation on how to explore it; take a look at: https://sonobuoy.io/docs/main/e2eplugin/

In this example, we'll trigger the test with 'LoadBalancers' in the name.

Steps:

- Run the tool, focusing on 'LoadBalancers':

```sh
./opct sonobuoy run --e2e-focus='LoadBalancers' --dns-namespace=openshift-dns --dns-pod-labels=dns.operator.openshift.io/daemonset-dns=default
```

- Check the status:

```sh
$ /home/mtulio/opct/bin/opct-devel sonobuoy status
```

- Check if the environment has been created

```sh
oc get pods -n sonobuoy -w
```

- Check the logs:

```sh
$ oc logs -l sonobuoy-plugin=e2e -n sonobuoy -c e2e
```

- Collect the results:

```sh
RESULT_FILE=$(./opct sonobuoy retrieve)
```

- Explore the results:

```sh
./opct sonobuoy results $RESULT_FILE -mode full
```

- Explore more:

```sh
./opct sonobuoy results $RESULT_FILE -p e2e

./opct sonobuoy results $RESULT_FILE -p e2e -m dump | yq e '.items[].items[].items[] | select(.status=="passed")' -
```

## Running `openshift-tests` directly

- Extract `openshift-tests` from the release image:

```sh
# OpenShift release version and arch
VERSION=v4.18.3
ARCH=$(uname -m)
RELEASE_IMAGE="quay.io/openshift-release-dev/ocp-release:${VERSION}-${ARCH}"
TESTS_IMAGE=$(oc adm release info --image-for=tests -a ${PULL_SECRET_FILE} ${RELEASE_IMAGE})

oc image extract $TESTS_IMAGE -a ${PULL_SECRET_FILE} \
    --file="/usr/bin/openshift-tests"
```

- Extract the tests with filters (e.g., Load Balancer):

```sh
./openshift-tests run all --dry-run  | grep '\[sig-network\] LoadBalancers' > ./tests-lb.txt

$ wc -l ./tests-lb.txt
20 ./tests-lb.txt
```

### Running in parallel mode (default):

```sh
./openshift-tests run --junit-dir ./junits -f ./tests-lb.txt | tee -a tests-lb-run.txt

grep -E ^'(passed|skipped|failed)' ./tests-lb-run.txt
grep ^passed ./tests-lb-run.txt
grep ^failed ./tests-lb-run.txt
grep ^skipped ./tests-lb-run.txt
```

### Running in serial mode (Parallel==1):

- Serial execution:

```sh
./openshift-tests run --junit-dir ./junits-serial -f ./tests-lb.txt --max-parallel-tests 1 | tee -a tests-lb-run-serial.txt

grep ^passed ./tests-lb-run-serial.txt
grep ^failed ./tests-lb-run-serial.txt
grep ^skipped ./tests-lb-run-serial.txt
```
