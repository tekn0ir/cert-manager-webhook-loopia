#!/bin/sh
# Downloads the binaries that the cert-manager ACME DNS-01 conformance test suite needs.
#
# Since cert-manager v1.11 the test fixture starts a complete Kubernetes control
# plane through controller-runtime's envtest, which requires the etcd,
# kube-apiserver and kubectl binaries. They are downloaded with setup-envtest,
# pinned to the same patch version as k8s.io/client-go in go.mod.
#
# Usage, both forms install setup-envtest and download the binaries into testdata/bin first:
#
#   sh testdata/scripts/fetch-test-binaries.sh         # prints the assets directory
#   sh testdata/scripts/fetch-test-binaries.sh --env   # prints the exports to evaluate
set -eu

BIN_DIR="$(pwd)/testdata/bin"
ENVTEST_K8S_VERSION="${ENVTEST_K8S_VERSION:-$(go list -m -f '{{.Version}}' k8s.io/client-go | sed 's/^v0\./1./')}"
# setup-envtest is versioned independently of controller-runtime, 'latest' is used so that
# the script keeps working on newer Go versions, override with SETUP_ENVTEST_VERSION if needed.
SETUP_ENVTEST_VERSION="${SETUP_ENVTEST_VERSION:-latest}"

# setup-envtest is a Go tool itself, it is installed into the directory on PATH.
GOPATH_BIN="$(go env GOPATH)/bin"
export PATH="${GOPATH_BIN}:${PATH}"
go install "sigs.k8s.io/controller-runtime/tools/setup-envtest@${SETUP_ENVTEST_VERSION}" >&2

mkdir -p "$BIN_DIR"
ASSETS_DIR="$(setup-envtest use --bin-dir "$BIN_DIR" -p path "$ENVTEST_K8S_VERSION")"

# controller-runtime finds the binaries through KUBEBUILDER_ASSETS, while the
# cert-manager test fixture looks them up on the PATH, so both are needed.
if [ "${1:-}" = "--env" ]; then
	printf 'export KUBEBUILDER_ASSETS="%s"\n' "$ASSETS_DIR"
	# shellcheck disable=SC2016
	# '$PATH' is deliberately not expanded here, it is meant to be expanded by the eval of the caller.
	printf 'export PATH="%s:$PATH"\n' "$ASSETS_DIR"
else
	printf '%s\n' "$ASSETS_DIR"
fi
