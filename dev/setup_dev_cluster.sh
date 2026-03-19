#!/usr/bin/env bash

# see https://kind.sigs.k8s.io/docs/user/local-registry/

set -o errexit

# --- Tool paths passed in from Bazel ---
KIND="${1:?Usage: $0 <kind_binary> <kubectl_binary>}"
KUBECTL="${2:?Usage: $0 <kind_binary> <kubectl_binary>}"
CLUSTER_CONFIG="${3:?Usage: $0 <kind> <kubectl> <cluster_config>}"

# 1. Create registry container unless it already exists
reg_name='kind-registry'
reg_port='5005'

if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --network bridge --name "${reg_name}" \
    registry:2
fi

# 2. Create kind cluster (no more nested bazel run)
"${KIND}" create cluster --config "${CLUSTER_CONFIG}"

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
