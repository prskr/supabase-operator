#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# --- begin runfiles.bash initialization v3 ---
# Copy-pasted from the Bazel Bash runfiles library v3.
set +e
f=bazel_tools/tools/bash/runfiles/runfiles.bash
# shellcheck disable=SC1090
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null ||
	source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null ||
	source "$0.runfiles/$f" 2>/dev/null ||
	source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null ||
	source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null ||
	{
		echo >&2 "ERROR: cannot find $f"
		exit 1
	}
f=
set -e
# --- end runfiles.bash initialization v3 ---

GO_BIN="$(rlocation "${GO_BIN}")"
SETUP_ENVTEST="$(rlocation "${SETUP_ENVTEST}")"

# Change to the workspace root
cd "${BUILD_WORKSPACE_DIRECTORY}"

# Prepare KUBEBUILDER_ASSETS
KUBEBUILDER_ASSETS="$("${SETUP_ENVTEST}" use -p path)"

echo "Updating snapshots using ${GO_BIN}..."
echo "KUBEBUILDER_ASSETS: ${KUBEBUILDER_ASSETS}"

export KUBEBUILDER_ASSETS="${KUBEBUILDER_ASSETS}"
# Try 'always' as it might bypass some CI detection if go-snaps thinks it's in CI
export UPDATE_SNAPS=always

"${GO_BIN}" test -count=1 -v ./internal/controller/...
