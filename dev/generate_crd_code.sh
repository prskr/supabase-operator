#!/usr/bin/env bash

set -o errexit

# --- begin runfiles.bash initialization v3 ---
# Copy-pasted from the Bazel Bash runfiles library v3.
set -uo pipefail
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

CONTROLLER_GEN="$(rlocation "${CONTROLLER_GEN}")"

"${CONTROLLER_GEN}" \
	rbac:roleName=manager-role crd webhook \
	paths="./..." \
	"output:crd:artifacts:config=${BUILD_WORKSPACE_DIRECTORY}/config/crd/bases"

"${CONTROLLER_GEN}" \
	object:headerFile="hack/boilerplate.go.txt" \
	paths="./api/v1alpha1/..." \
	"output:dir=${BUILD_WORKSPACE_DIRECTORY}/api/v1alpha1"
