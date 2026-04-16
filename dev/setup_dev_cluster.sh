#!/usr/bin/env bash

# see https://kind.sigs.k8s.io/docs/user/local-registry/

set -euo pipefail

# --- begin runfiles.bash initialization v3 ---
# Copy-pasted from the Bazel Bash runfiles library v3.
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

# --- Tool paths passed in from Bazel ---
KIND="$(rlocation "${KIND_BIN}")"
KUBECTL="$(rlocation "${KUBECTL_BIN}")"
KUSTOMIZE="$(rlocation "${KUSTOMIZE_BIN}")"
CLUSTER_CONFIG="$(rlocation "${CLUSTER_CONFIG}")"
DEX_IDP_CONFIG="$(rlocation "${DEX_IDP_CONFIG}")"
CONFIG_DEPS="$(rlocation "${CONFIG_DEPS}")"

# 1. Create registry container unless it already exists
reg_name='kind-registry'
reg_port='5005'

if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
	docker run \
		-d --restart=always -p "127.0.0.1:${reg_port}:5000" --network bridge --name "${reg_name}" \
		registry:2
fi

idp_name='studio-idp'
idp_port='5556'

if [ "$(docker inspect -f '{{.State.Running}}' "${idp_name}" 2>/dev/null || true)" != 'true' ]; then
	docker run \
		-d --restart=always -p "127.0.0.1:${idp_port}:5556" --network bridge --name "${idp_name}" \
		-v "${DEX_IDP_CONFIG}":/etc/dex/config.yaml \
		ghcr.io/dexidp/dex \
		dex serve /etc/dex/config.yaml
fi

# 2. Create kind cluster only if it doesn't already exist
if ! "${KIND}" get clusters | grep -q "supabase-operator-debug"; then
	"${KIND}" create cluster --config "${CLUSTER_CONFIG}"
fi

# 3. Add the registry config to the nodes
REGISTRY_DIR="/etc/containerd/certs.d/localhost:${reg_port}"
for node in $("${KIND}" get nodes --name supabase-operator-debug); do
	docker exec "${node}" mkdir -p "${REGISTRY_DIR}"
	cat <<EOF | docker exec -i "${node}" cp /dev/stdin "${REGISTRY_DIR}/hosts.toml"
[host."http://${reg_name}:5000"]
EOF
done

# 4. Connect the registry to the cluster network if not already connected
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${reg_name}")" = 'null' ]; then
	docker network connect "kind" "${reg_name}"
fi

# Connect the IDP to the cluster network if not already connected
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${idp_name}")" = 'null' ]; then
	docker network connect "kind" "${idp_name}"
fi

# 5. Document the local registry
cat <<EOF | "${KUBECTL}" apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

# Retry the kustomize build and apply command up to 5 times with 3-second intervals
max_attempts=5
attempt=1
success=false

while [ $attempt -le $max_attempts ]; do
	if "${KUSTOMIZE}" build "$(dirname "${CONFIG_DEPS}")" | "${KUBECTL}" apply --server-side=true --force-conflicts -f -; then
		success=true
		break
	else
		echo "Attempt $attempt failed. Retrying in 3 seconds..."
		sleep 3
		attempt=$((attempt + 1))
	fi
done

if [ "$success" = false ]; then
	echo "Failed to apply kustomize configuration after $max_attempts attempts."
	exit 1
fi
